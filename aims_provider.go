package oaiaf

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AIMSProvider implements the AIMS protocol using SPIFFE workload identity.
// AIMS (Agent Identity and Messaging System) uses SPIFFE SVIDs for workload
// identity and mTLS for authentication.
type AIMSProvider struct {
	agent      *Agent
	httpClient *http.Client

	// SPIFFE configuration
	spiffeID       string // e.g., spiffe://example.com/agent/my-agent
	trustBundle    *x509.CertPool
	svid           *tls.Certificate
	workloadSocket string // Path to SPIFFE Workload API socket
}

// AIMSOption configures the AIMS provider.
type AIMSOption func(*AIMSProvider)

// WithSPIFFEID sets the expected SPIFFE ID for the agent.
func WithSPIFFEID(id string) AIMSOption {
	return func(p *AIMSProvider) {
		p.spiffeID = id
	}
}

// WithTrustBundle sets the trust bundle for validating peer certificates.
func WithTrustBundle(bundle *x509.CertPool) AIMSOption {
	return func(p *AIMSProvider) {
		p.trustBundle = bundle
	}
}

// WithSVID sets the X.509 SVID certificate for mTLS authentication.
func WithSVID(cert *tls.Certificate) AIMSOption {
	return func(p *AIMSProvider) {
		p.svid = cert
	}
}

// WithWorkloadSocket sets the path to the SPIFFE Workload API socket.
// If set, the provider will attempt to fetch SVIDs from the Workload API.
func WithWorkloadSocket(path string) AIMSOption {
	return func(p *AIMSProvider) {
		p.workloadSocket = path
	}
}

// NewAIMSProvider creates a new AIMS provider for the agent.
func NewAIMSProvider(agent *Agent, opts ...AIMSOption) *AIMSProvider {
	p := &AIMSProvider{
		agent:      agent,
		httpClient: agent.httpClient,
	}

	for _, opt := range opts {
		opt(p)
	}

	// If SVID and trust bundle are configured, create an mTLS client
	if p.svid != nil && p.trustBundle != nil {
		p.httpClient = p.createMTLSClient()
	}

	return p
}

// Protocol returns the AIMS protocol identifier.
func (p *AIMSProvider) Protocol() Protocol {
	return ProtocolAIMS
}

// AcquireToken requests an access token using AIMS/SPIFFE authentication.
// The token is obtained by presenting the X.509 SVID to the authorization server.
func (p *AIMSProvider) AcquireToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	if p.agent.authServer == "" {
		return nil, fmt.Errorf("authorization server not configured")
	}

	// Verify we have credentials
	if p.svid == nil {
		return nil, fmt.Errorf("SVID not configured; use WithSVID or WithWorkloadSocket")
	}

	// Perform token request with mTLS
	tokenURL := strings.TrimSuffix(p.agent.authServer, "/") + "/token"

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	form.Set("subject_token_type", "urn:spiffe:params:oauth:token-type:svid")
	form.Set("scope", strings.Join(scopes, " "))

	// Add SPIFFE ID if configured
	if p.spiffeID != "" {
		form.Set("subject_token", p.spiffeID)
	}

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

// SPIFFEID returns the configured SPIFFE ID.
func (p *AIMSProvider) SPIFFEID() string {
	return p.spiffeID
}

// createMTLSClient creates an HTTP client configured for mTLS using the SVID.
func (p *AIMSProvider) createMTLSClient() *http.Client {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*p.svid},
		RootCAs:      p.trustBundle,
		MinVersion:   tls.VersionTLS12,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// FetchSVIDFromWorkloadAPI attempts to fetch an X.509 SVID from the SPIFFE
// Workload API. This requires a SPIRE agent running on the host.
//
// Note: This is a simplified implementation. In production, you would use
// the github.com/spiffe/go-spiffe/v2/workloadapi package.
func (p *AIMSProvider) FetchSVIDFromWorkloadAPI(ctx context.Context) error {
	if p.workloadSocket == "" {
		return fmt.Errorf("workload socket not configured")
	}

	// In a real implementation, this would:
	// 1. Connect to the Workload API socket
	// 2. Call FetchX509SVID()
	// 3. Store the SVID and trust bundle
	//
	// For now, we return an error indicating this needs the go-spiffe library
	return fmt.Errorf("SVID fetch requires github.com/spiffe/go-spiffe/v2/workloadapi; " +
		"use WithSVID() to provide credentials directly")
}

// ValidatePeerSVID validates a peer's X.509 SVID certificate.
// This can be used by servers to verify incoming agent connections.
func ValidatePeerSVID(cert *x509.Certificate, trustBundle *x509.CertPool, expectedSPIFFEID string) error {
	if cert == nil {
		return fmt.Errorf("no certificate provided")
	}

	// Verify against trust bundle
	opts := x509.VerifyOptions{
		Roots: trustBundle,
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	// Extract and validate SPIFFE ID from URI SAN
	var spiffeID string
	for _, uri := range cert.URIs {
		if strings.HasPrefix(uri.String(), "spiffe://") {
			spiffeID = uri.String()
			break
		}
	}

	if spiffeID == "" {
		return fmt.Errorf("no SPIFFE ID found in certificate")
	}

	if expectedSPIFFEID != "" && spiffeID != expectedSPIFFEID {
		return fmt.Errorf("SPIFFE ID mismatch: expected %s, got %s", expectedSPIFFEID, spiffeID)
	}

	return nil
}

// ParseSPIFFEID parses a SPIFFE ID and returns its components.
type SPIFFEIDComponents struct {
	TrustDomain string
	Path        string
}

// ParseSPIFFEID parses a SPIFFE ID string into its components.
func ParseSPIFFEID(id string) (*SPIFFEIDComponents, error) {
	if !strings.HasPrefix(id, "spiffe://") {
		return nil, fmt.Errorf("invalid SPIFFE ID: must start with spiffe://")
	}

	// Remove scheme
	rest := strings.TrimPrefix(id, "spiffe://")

	// Split into trust domain and path
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("invalid SPIFFE ID: missing trust domain")
	}

	components := &SPIFFEIDComponents{
		TrustDomain: parts[0],
	}

	if len(parts) > 1 {
		components.Path = "/" + parts[1]
	}

	return components, nil
}
