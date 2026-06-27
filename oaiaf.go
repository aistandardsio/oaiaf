// Package oaiaf provides the OpenAI Agent Framework - a framework for building
// AI agents with built-in support for agent authorization protocols.
//
// OAIAF enables agents to:
//   - Authenticate using ID-JAG, AAuth, and AIMS protocols
//   - Request and manage authorization for sensitive operations
//   - Work with human-in-the-loop consent flows
//   - Integrate with existing identity providers
//
// # Quick Start
//
//	agent := oaiaf.NewAgent("my-agent",
//	    oaiaf.WithAuthServer("https://auth.example.com"),
//	    oaiaf.WithCredentials(privateKey, keyID),
//	)
//
//	// Make an authorized request
//	resp, err := agent.AuthorizedRequest(ctx, "read:email",
//	    oaiaf.Get("https://api.example.com/user/email"),
//	)
package oaiaf

import (
	"context"
	"crypto"
	"net/http"
	"strings"
	"time"
)

// Version is the current version of the OAIAF package.
const Version = "0.1.0"

// Protocol represents a supported authorization protocol.
type Protocol string

// Supported protocols.
const (
	ProtocolIDJAG Protocol = "idjag"
	ProtocolAAuth Protocol = "aauth"
	ProtocolAIMS  Protocol = "aims"
)

// Agent represents an AI agent with authorization capabilities.
type Agent struct {
	id         string
	name       string
	authServer string
	privateKey crypto.PrivateKey
	keyID      string
	httpClient *http.Client
	protocol   Protocol

	// Token caching
	tokenCache map[string]*CachedToken

	// Provider for token acquisition
	provider Provider
}

// CachedToken holds a cached access token.
type CachedToken struct {
	AccessToken string
	ExpiresAt   time.Time
	Scopes      []string
}

// Option configures an Agent.
type Option func(*Agent)

// NewAgent creates a new agent with the given ID and options.
func NewAgent(id string, opts ...Option) *Agent {
	agent := &Agent{
		id:         id,
		httpClient: http.DefaultClient,
		tokenCache: make(map[string]*CachedToken),
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// WithName sets the agent's display name.
func WithName(name string) Option {
	return func(a *Agent) {
		a.name = name
	}
}

// WithAuthServer sets the authorization server URL.
func WithAuthServer(url string) Option {
	return func(a *Agent) {
		a.authServer = url
	}
}

// WithCredentials sets the agent's signing credentials.
func WithCredentials(key crypto.PrivateKey, keyID string) Option {
	return func(a *Agent) {
		a.privateKey = key
		a.keyID = keyID
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(a *Agent) {
		a.httpClient = client
	}
}

// WithProtocol sets the default authorization protocol.
func WithProtocol(protocol Protocol) Option {
	return func(a *Agent) {
		a.protocol = protocol
	}
}

// WithProvider sets a custom authorization provider.
func WithProvider(provider Provider) Option {
	return func(a *Agent) {
		a.provider = provider
	}
}

// ID returns the agent's identifier.
func (a *Agent) ID() string {
	return a.id
}

// Name returns the agent's display name.
func (a *Agent) Name() string {
	if a.name == "" {
		return a.id
	}
	return a.name
}

// AuthorizedRequest makes an HTTP request with the given scope.
// It handles token acquisition, caching, and automatic renewal.
func (a *Agent) AuthorizedRequest(ctx context.Context, scope string, req *http.Request) (*http.Response, error) {
	// Get or acquire token for the scope
	token, err := a.getToken(ctx, scope)
	if err != nil {
		return nil, err
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// Make the request
	return a.httpClient.Do(req)
}

// getToken retrieves a cached token or acquires a new one.
func (a *Agent) getToken(ctx context.Context, scope string) (*CachedToken, error) {
	// Check cache
	if cached, ok := a.tokenCache[scope]; ok {
		if time.Now().Before(cached.ExpiresAt) {
			return cached, nil
		}
		// Token expired, remove from cache
		delete(a.tokenCache, scope)
	}

	// Acquire new token
	token, err := a.acquireToken(ctx, scope)
	if err != nil {
		return nil, err
	}

	// Cache the token
	a.tokenCache[scope] = token

	return token, nil
}

// acquireToken requests a new access token from the authorization server.
func (a *Agent) acquireToken(ctx context.Context, scope string) (*CachedToken, error) {
	// Get or create provider
	provider := a.getProvider()

	// Acquire token using the provider
	resp, err := provider.AcquireToken(ctx, []string{scope})
	if err != nil {
		return nil, err
	}

	// Parse scopes from response
	var scopes []string
	if resp.Scope != "" {
		scopes = splitScopes(resp.Scope)
	} else {
		scopes = []string{scope}
	}

	return &CachedToken{
		AccessToken: resp.AccessToken,
		ExpiresAt:   resp.ExpiresAt,
		Scopes:      scopes,
	}, nil
}

// getProvider returns the configured provider or creates a default one.
func (a *Agent) getProvider() Provider {
	if a.provider != nil {
		return a.provider
	}

	// Create default provider based on protocol
	switch a.protocol {
	case ProtocolAAuth:
		return NewAAuthProvider(a)
	case ProtocolAIMS:
		return NewAIMSProvider(a)
	case ProtocolIDJAG:
		fallthrough
	default:
		return NewIDJAGProvider(a)
	}
}

// splitScopes splits a space-separated scope string.
func splitScopes(scopes string) []string {
	if scopes == "" {
		return nil
	}
	var result []string
	for _, s := range strings.Split(scopes, " ") {
		if s = strings.TrimSpace(s); s != "" {
			result = append(result, s)
		}
	}
	return result
}

// ClearTokenCache removes all cached tokens.
func (a *Agent) ClearTokenCache() {
	a.tokenCache = make(map[string]*CachedToken)
}
