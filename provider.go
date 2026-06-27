package oaiaf

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Provider defines the interface for authorization providers.
// Each protocol (ID-JAG, AAuth, AIMS) has its own provider implementation.
type Provider interface {
	// Protocol returns the protocol identifier.
	Protocol() Protocol

	// AcquireToken requests an access token for the given scopes.
	AcquireToken(ctx context.Context, scopes []string) (*TokenResponse, error)
}

// TokenResponse represents the response from a token acquisition request.
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"-"`
	Scope        string    `json:"scope,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

// ErrorResponse represents an OAuth error response.
type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// ProviderOption configures a provider.
type ProviderOption func(any)

// IDJAGProvider implements the ID-JAG protocol for automated authorization.
type IDJAGProvider struct {
	agent      *Agent
	httpClient *http.Client
}

// NewIDJAGProvider creates a new ID-JAG provider for the agent.
func NewIDJAGProvider(agent *Agent, opts ...ProviderOption) *IDJAGProvider {
	p := &IDJAGProvider{
		agent:      agent,
		httpClient: agent.httpClient,
	}
	return p
}

// Protocol returns the ID-JAG protocol identifier.
func (p *IDJAGProvider) Protocol() Protocol {
	return ProtocolIDJAG
}

// AcquireToken requests an access token using ID-JAG token exchange.
func (p *IDJAGProvider) AcquireToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	if p.agent.authServer == "" {
		return nil, fmt.Errorf("authorization server not configured")
	}
	if p.agent.privateKey == nil {
		return nil, fmt.Errorf("signing credentials not configured")
	}

	// Create ID-JAG assertion JWT
	assertion, err := p.createAssertion(scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to create assertion: %w", err)
	}

	// Perform token exchange
	tokenURL := strings.TrimSuffix(p.agent.authServer, "/") + "/token"

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	form.Set("subject_token", assertion)
	form.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")
	form.Set("scope", strings.Join(scopes, " "))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("token error: %s - %s", errResp.Error, errResp.Description)
		}
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Calculate expiration time
	if tokenResp.ExpiresIn > 0 {
		tokenResp.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	return &tokenResp, nil
}

// createAssertion creates a signed ID-JAG assertion JWT.
func (p *IDJAGProvider) createAssertion(scopes []string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"iss":   p.agent.id,
		"sub":   p.agent.id,
		"aud":   p.agent.authServer,
		"iat":   now.Unix(),
		"exp":   now.Add(5 * time.Minute).Unix(),
		"jti":   fmt.Sprintf("%d", now.UnixNano()),
		"scope": strings.Join(scopes, " "),
	}

	// Determine signing method based on key type
	var signingMethod jwt.SigningMethod
	switch p.agent.privateKey.(type) {
	case *ecdsa.PrivateKey:
		signingMethod = jwt.SigningMethodES256
	case *rsa.PrivateKey:
		signingMethod = jwt.SigningMethodRS256
	case ed25519.PrivateKey:
		signingMethod = jwt.SigningMethodEdDSA
	default:
		return "", fmt.Errorf("unsupported key type: %T", p.agent.privateKey)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	if p.agent.keyID != "" {
		token.Header["kid"] = p.agent.keyID
	}

	return token.SignedString(p.agent.privateKey)
}

// AAuthProvider implements the AAuth protocol for human-in-the-loop authorization.
type AAuthProvider struct {
	agent      *Agent
	httpClient *http.Client

	// ConsentHandler is called when human consent is required.
	// It receives the consent URI and should display it to the user.
	// Returns true if consent was granted, false otherwise.
	ConsentHandler func(consentURI string) (bool, error)

	// PollInterval is the interval for polling consent status.
	PollInterval time.Duration

	// Timeout is the maximum time to wait for consent.
	Timeout time.Duration
}

// NewAAuthProvider creates a new AAuth provider for the agent.
func NewAAuthProvider(agent *Agent, opts ...ProviderOption) *AAuthProvider {
	p := &AAuthProvider{
		agent:        agent,
		httpClient:   agent.httpClient,
		PollInterval: 2 * time.Second,
		Timeout:      10 * time.Minute,
	}
	return p
}

// Protocol returns the AAuth protocol identifier.
func (p *AAuthProvider) Protocol() Protocol {
	return ProtocolAAuth
}

// AcquireToken requests an access token using AAuth consent flow.
func (p *AAuthProvider) AcquireToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	if p.agent.authServer == "" {
		return nil, fmt.Errorf("authorization server not configured")
	}

	// Create authorization request
	authResp, err := p.requestAuthorization(ctx, scopes)
	if err != nil {
		return nil, err
	}

	// If we got an immediate token, return it
	if authResp.AccessToken != "" {
		return &TokenResponse{
			AccessToken: authResp.AccessToken,
			TokenType:   authResp.TokenType,
			ExpiresIn:   authResp.ExpiresIn,
			ExpiresAt:   time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
			Scope:       authResp.Scope,
		}, nil
	}

	// Otherwise, wait for consent
	if authResp.StatusURI == "" {
		return nil, fmt.Errorf("no status URI provided for consent polling")
	}

	// Notify consent handler if configured
	if p.ConsentHandler != nil && authResp.ConsentURI != "" {
		approved, err := p.ConsentHandler(authResp.ConsentURI)
		if err != nil {
			return nil, fmt.Errorf("consent handler error: %w", err)
		}
		if !approved {
			return nil, fmt.Errorf("consent was denied")
		}
	}

	// Poll for consent status
	return p.pollConsentStatus(ctx, authResp.StatusURI)
}

// AuthorizationResponse represents the response from an authorization request.
type AuthorizationResponse struct {
	// For immediate approval
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`

	// For deferred consent
	ConsentURI string `json:"consent_uri,omitempty"`
	StatusURI  string `json:"status_uri,omitempty"`
	MissionID  string `json:"mission_id,omitempty"`
	Interval   int    `json:"interval,omitempty"`
}

// ConsentStatusResponse represents the response when polling for consent status.
type ConsentStatusResponse struct {
	Status      string `json:"status"` // pending, approved, denied, expired
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Error       string `json:"error,omitempty"`
	ErrorDesc   string `json:"error_description,omitempty"`
}

func (p *AAuthProvider) requestAuthorization(ctx context.Context, scopes []string) (*AuthorizationResponse, error) {
	authURL := strings.TrimSuffix(p.agent.authServer, "/") + "/authorize"

	reqBody := map[string]any{
		"agent_token": p.agent.id, // Use agent ID as a simple token
		"user_id":     "default",  // Would be set by the caller in a real implementation
		"scope":       strings.Join(scopes, " "),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authorization request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("authorization error: %s - %s", errResp.Error, errResp.Description)
		}
		return nil, fmt.Errorf("authorization request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthorizationResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &authResp, nil
}

func (p *AAuthProvider) pollConsentStatus(ctx context.Context, statusURI string) (*TokenResponse, error) {
	timeout := time.After(p.Timeout)
	ticker := time.NewTicker(p.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("consent timeout exceeded")
		case <-ticker.C:
			status, err := p.checkConsentStatus(ctx, statusURI)
			if err != nil {
				return nil, err
			}

			switch status.Status {
			case "approved":
				return &TokenResponse{
					AccessToken: status.AccessToken,
					TokenType:   status.TokenType,
					ExpiresIn:   status.ExpiresIn,
					ExpiresAt:   time.Now().Add(time.Duration(status.ExpiresIn) * time.Second),
					Scope:       status.Scope,
				}, nil
			case "denied":
				return nil, fmt.Errorf("consent denied: %s", status.ErrorDesc)
			case "expired":
				return nil, fmt.Errorf("consent request expired")
			case "pending":
				// Continue polling
			default:
				return nil, fmt.Errorf("unknown consent status: %s", status.Status)
			}
		}
	}
}

func (p *AAuthProvider) checkConsentStatus(ctx context.Context, statusURI string) (*ConsentStatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("status request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var status ConsentStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	return &status, nil
}
