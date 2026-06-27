package oaiaf

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent("test-agent")

	if agent.ID() != "test-agent" {
		t.Errorf("expected ID test-agent, got %s", agent.ID())
	}

	if agent.Name() != "test-agent" {
		t.Errorf("expected Name test-agent (fallback to ID), got %s", agent.Name())
	}
}

func TestNewAgentWithOptions(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	client := &http.Client{Timeout: 30 * time.Second}

	agent := NewAgent("test-agent",
		WithName("Test Agent"),
		WithAuthServer("https://auth.example.com"),
		WithCredentials(privateKey, "key-1"),
		WithHTTPClient(client),
		WithProtocol(ProtocolIDJAG),
	)

	if agent.Name() != "Test Agent" {
		t.Errorf("expected Name Test Agent, got %s", agent.Name())
	}

	if agent.authServer != "https://auth.example.com" {
		t.Errorf("expected authServer https://auth.example.com, got %s", agent.authServer)
	}

	if agent.privateKey != privateKey {
		t.Error("expected privateKey to be set")
	}

	if agent.keyID != "key-1" {
		t.Errorf("expected keyID key-1, got %s", agent.keyID)
	}

	if agent.httpClient != client {
		t.Error("expected httpClient to be set")
	}

	if agent.protocol != ProtocolIDJAG {
		t.Errorf("expected protocol idjag, got %s", agent.protocol)
	}
}

func TestClearTokenCache(t *testing.T) {
	agent := NewAgent("test-agent")

	// Add a token to cache
	agent.tokenCache["scope1"] = &CachedToken{
		AccessToken: "token1",
		ExpiresAt:   time.Now().Add(time.Hour),
		Scopes:      []string{"scope1"},
	}

	if len(agent.tokenCache) != 1 {
		t.Errorf("expected 1 cached token, got %d", len(agent.tokenCache))
	}

	agent.ClearTokenCache()

	if len(agent.tokenCache) != 0 {
		t.Errorf("expected 0 cached tokens after clear, got %d", len(agent.tokenCache))
	}
}

func TestGetTokenFromCache(t *testing.T) {
	agent := NewAgent("test-agent")

	// Add a valid token to cache
	agent.tokenCache["scope1"] = &CachedToken{
		AccessToken: "cached-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		Scopes:      []string{"scope1"},
	}

	token, err := agent.getToken(context.Background(), "scope1")
	if err != nil {
		t.Fatalf("getToken failed: %v", err)
	}

	if token.AccessToken != "cached-token" {
		t.Errorf("expected cached-token, got %s", token.AccessToken)
	}
}

func TestGetTokenExpiredCache(t *testing.T) {
	// Create a mock server that returns tokens
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := TokenResponse{
			AccessToken: "new-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "scope1",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	agent := NewAgent("test-agent",
		WithAuthServer(ts.URL),
		WithCredentials(privateKey, "key-1"),
	)

	// Add an expired token to cache
	agent.tokenCache["scope1"] = &CachedToken{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-time.Hour), // Expired
		Scopes:      []string{"scope1"},
	}

	token, err := agent.getToken(context.Background(), "scope1")
	if err != nil {
		t.Fatalf("getToken failed: %v", err)
	}

	// Should have acquired a new token
	if token.AccessToken != "new-token" {
		t.Errorf("expected new-token, got %s", token.AccessToken)
	}
}

func TestSplitScopes(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"read:email", []string{"read:email"}},
		{"read:email read:profile", []string{"read:email", "read:profile"}},
		{"  read:email   read:profile  ", []string{"read:email", "read:profile"}},
	}

	for _, tt := range tests {
		result := splitScopes(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitScopes(%q) = %v, expected %v", tt.input, result, tt.expected)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("splitScopes(%q)[%d] = %q, expected %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestIDJAGProviderProtocol(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewIDJAGProvider(agent)

	if provider.Protocol() != ProtocolIDJAG {
		t.Errorf("expected protocol idjag, got %s", provider.Protocol())
	}
}

func TestIDJAGProviderAcquireToken(t *testing.T) {
	// Create a mock token server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			http.Error(w, "bad content type", http.StatusBadRequest)
			return
		}

		// Return a token
		w.Header().Set("Content-Type", "application/json")
		resp := TokenResponse{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "read:email",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	agent := NewAgent("test-agent",
		WithAuthServer(ts.URL),
		WithCredentials(privateKey, "key-1"),
	)

	provider := NewIDJAGProvider(agent)

	token, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err != nil {
		t.Fatalf("AcquireToken failed: %v", err)
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("expected access token test-access-token, got %s", token.AccessToken)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("expected token type Bearer, got %s", token.TokenType)
	}

	if token.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set")
	}
}

func TestIDJAGProviderMissingAuthServer(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewIDJAGProvider(agent)

	_, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err == nil {
		t.Error("expected error for missing auth server")
	}
}

func TestIDJAGProviderMissingCredentials(t *testing.T) {
	agent := NewAgent("test-agent",
		WithAuthServer("https://auth.example.com"),
	)
	provider := NewIDJAGProvider(agent)

	_, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err == nil {
		t.Error("expected error for missing credentials")
	}
}

func TestAAuthProviderProtocol(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewAAuthProvider(agent)

	if provider.Protocol() != ProtocolAAuth {
		t.Errorf("expected protocol aauth, got %s", provider.Protocol())
	}
}

func TestAAuthProviderImmediateToken(t *testing.T) {
	// Create a mock server that returns an immediate token
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/authorize" {
			w.Header().Set("Content-Type", "application/json")
			resp := AuthorizationResponse{
				AccessToken: "immediate-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read:email",
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	agent := NewAgent("test-agent",
		WithAuthServer(ts.URL),
	)

	provider := NewAAuthProvider(agent)

	token, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err != nil {
		t.Fatalf("AcquireToken failed: %v", err)
	}

	if token.AccessToken != "immediate-token" {
		t.Errorf("expected access token immediate-token, got %s", token.AccessToken)
	}
}

func TestAAuthProviderConsentFlow(t *testing.T) {
	consentCalls := 0
	statusCalls := 0

	// Use a variable to store the server URL since we need it in the handler
	var serverURL string

	// Create a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/authorize":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			resp := AuthorizationResponse{
				ConsentURI: "http://example.com/consent/123",
				StatusURI:  serverURL + "/consent/status/123",
				MissionID:  "123",
				Interval:   1,
			}
			_ = json.NewEncoder(w).Encode(resp)
			consentCalls++

		case "/consent/status/123":
			statusCalls++
			w.Header().Set("Content-Type", "application/json")
			// First call returns pending, second returns approved
			if statusCalls < 2 {
				_ = json.NewEncoder(w).Encode(ConsentStatusResponse{Status: "pending"})
			} else {
				_ = json.NewEncoder(w).Encode(ConsentStatusResponse{
					Status:      "approved",
					AccessToken: "consent-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
					Scope:       "read:email",
				})
			}

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Set the server URL now that we have it
	serverURL = ts.URL

	agent := NewAgent("test-agent",
		WithAuthServer(ts.URL),
	)

	provider := NewAAuthProvider(agent)
	provider.PollInterval = 10 * time.Millisecond // Speed up polling for test

	token, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err != nil {
		t.Fatalf("AcquireToken failed: %v", err)
	}

	if token.AccessToken != "consent-token" {
		t.Errorf("expected access token consent-token, got %s", token.AccessToken)
	}

	if consentCalls != 1 {
		t.Errorf("expected 1 authorize call, got %d", consentCalls)
	}

	if statusCalls < 2 {
		t.Errorf("expected at least 2 status calls, got %d", statusCalls)
	}
}

func TestAAuthProviderMissingAuthServer(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewAAuthProvider(agent)

	_, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err == nil {
		t.Error("expected error for missing auth server")
	}
}

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name     string
		protocol Protocol
		expected Protocol
	}{
		{"default", "", ProtocolIDJAG},
		{"idjag", ProtocolIDJAG, ProtocolIDJAG},
		{"aauth", ProtocolAAuth, ProtocolAAuth},
		{"aims", ProtocolAIMS, ProtocolAIMS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent("test-agent", WithProtocol(tt.protocol))
			provider := agent.getProvider()

			if provider.Protocol() != tt.expected {
				t.Errorf("expected protocol %s, got %s", tt.expected, provider.Protocol())
			}
		})
	}
}

func TestAuthorizedRequest(t *testing.T) {
	// Create a mock API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer apiServer.Close()

	// Create a mock auth server
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "read:email",
		})
	}))
	defer authServer.Close()

	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	agent := NewAgent("test-agent",
		WithAuthServer(authServer.URL),
		WithCredentials(privateKey, "key-1"),
	)

	req, _ := http.NewRequest(http.MethodGet, apiServer.URL+"/resource", nil)
	resp, err := agent.AuthorizedRequest(context.Background(), "read:email", req)
	if err != nil {
		t.Fatalf("AuthorizedRequest failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
