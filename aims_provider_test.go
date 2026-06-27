package oaiaf

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestAIMSProviderProtocol(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewAIMSProvider(agent)

	if provider.Protocol() != ProtocolAIMS {
		t.Errorf("expected protocol aims, got %s", provider.Protocol())
	}
}

func TestAIMSProviderSPIFFEID(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewAIMSProvider(agent, WithSPIFFEID("spiffe://example.com/agent/test"))

	if provider.SPIFFEID() != "spiffe://example.com/agent/test" {
		t.Errorf("expected SPIFFE ID spiffe://example.com/agent/test, got %s", provider.SPIFFEID())
	}
}

func TestAIMSProviderMissingAuthServer(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewAIMSProvider(agent)

	_, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err == nil {
		t.Error("expected error for missing auth server")
	}
}

func TestAIMSProviderMissingSVID(t *testing.T) {
	agent := NewAgent("test-agent",
		WithAuthServer("https://auth.example.com"),
	)
	provider := NewAIMSProvider(agent)

	_, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err == nil {
		t.Error("expected error for missing SVID")
	}
}

func TestAIMSProviderFetchSVIDNotConfigured(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewAIMSProvider(agent)

	err := provider.FetchSVIDFromWorkloadAPI(context.Background())
	if err == nil {
		t.Error("expected error for missing workload socket")
	}
}

func TestAIMSProviderFetchSVIDRequiresLibrary(t *testing.T) {
	agent := NewAgent("test-agent")
	provider := NewAIMSProvider(agent, WithWorkloadSocket("/tmp/spiffe-workload-api.sock"))

	err := provider.FetchSVIDFromWorkloadAPI(context.Background())
	if err == nil {
		t.Error("expected error indicating go-spiffe library is needed")
	}
}

func TestParseSPIFFEID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		trustDomain string
		path        string
		expectErr   bool
	}{
		{
			name:        "simple",
			input:       "spiffe://example.com/agent/test",
			trustDomain: "example.com",
			path:        "/agent/test",
		},
		{
			name:        "no path",
			input:       "spiffe://example.com",
			trustDomain: "example.com",
			path:        "",
		},
		{
			name:        "deep path",
			input:       "spiffe://prod.example.com/region/us-west/agent/my-agent",
			trustDomain: "prod.example.com",
			path:        "/region/us-west/agent/my-agent",
		},
		{
			name:      "invalid prefix",
			input:     "https://example.com/agent",
			expectErr: true,
		},
		{
			name:      "empty",
			input:     "",
			expectErr: true,
		},
		{
			name:      "missing trust domain",
			input:     "spiffe:///agent",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components, err := ParseSPIFFEID(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if components.TrustDomain != tt.trustDomain {
				t.Errorf("expected trust domain %s, got %s", tt.trustDomain, components.TrustDomain)
			}

			if components.Path != tt.path {
				t.Errorf("expected path %s, got %s", tt.path, components.Path)
			}
		})
	}
}

func TestValidatePeerSVID(t *testing.T) {
	// Generate a test CA and certificate
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	caCertDER, _ := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	caCert, _ := x509.ParseCertificate(caCertDER)

	// Create trust bundle
	trustBundle := x509.NewCertPool()
	trustBundle.AddCert(caCert)

	// Generate leaf certificate with SPIFFE ID
	leafKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spiffeURI, _ := url.Parse("spiffe://example.com/agent/test")
	leafTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "Test Agent"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		URIs:                  []*url.URL{spiffeURI},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	leafCertDER, _ := x509.CreateCertificate(rand.Reader, leafTemplate, caCert, &leafKey.PublicKey, caKey)
	leafCert, _ := x509.ParseCertificate(leafCertDER)

	t.Run("valid certificate", func(t *testing.T) {
		err := ValidatePeerSVID(leafCert, trustBundle, "spiffe://example.com/agent/test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("wrong SPIFFE ID", func(t *testing.T) {
		err := ValidatePeerSVID(leafCert, trustBundle, "spiffe://example.com/agent/other")
		if err == nil {
			t.Error("expected error for SPIFFE ID mismatch")
		}
	})

	t.Run("no expected ID", func(t *testing.T) {
		err := ValidatePeerSVID(leafCert, trustBundle, "")
		if err != nil {
			t.Errorf("unexpected error when no expected ID: %v", err)
		}
	})

	t.Run("nil certificate", func(t *testing.T) {
		err := ValidatePeerSVID(nil, trustBundle, "")
		if err == nil {
			t.Error("expected error for nil certificate")
		}
	})

	t.Run("certificate without SPIFFE ID", func(t *testing.T) {
		noSPIFFETemplate := &x509.Certificate{
			SerialNumber:          big.NewInt(3),
			Subject:               pkix.Name{CommonName: "No SPIFFE"},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(time.Hour),
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}
		noSPIFFEDER, _ := x509.CreateCertificate(rand.Reader, noSPIFFETemplate, caCert, &leafKey.PublicKey, caKey)
		noSPIFFECert, _ := x509.ParseCertificate(noSPIFFEDER)

		err := ValidatePeerSVID(noSPIFFECert, trustBundle, "")
		if err == nil {
			t.Error("expected error for certificate without SPIFFE ID")
		}
	})
}

func TestAIMSProviderWithMTLS(t *testing.T) {
	// Generate CA
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caCertDER, _ := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	caCert, _ := x509.ParseCertificate(caCertDER)

	trustBundle := x509.NewCertPool()
	trustBundle.AddCert(caCert)

	// Generate server certificate
	serverKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverCertDER, _ := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)

	serverCert := tls.Certificate{
		Certificate: [][]byte{serverCertDER},
		PrivateKey:  serverKey,
	}

	// Generate client certificate with SPIFFE ID
	clientKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spiffeURI, _ := url.Parse("spiffe://example.com/agent/test")
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "Test Agent"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		URIs:         []*url.URL{spiffeURI},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientCertDER, _ := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)

	clientCert := &tls.Certificate{
		Certificate: [][]byte{clientCertDER},
		PrivateKey:  clientKey,
	}

	// Create mTLS server
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(TokenResponse{
				AccessToken: "aims-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read:email",
			})
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))

	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    trustBundle,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}
	server.StartTLS()
	defer server.Close()

	// Create agent with AIMS provider
	agent := NewAgent("test-agent",
		WithAuthServer(server.URL),
	)

	provider := NewAIMSProvider(agent,
		WithSPIFFEID("spiffe://example.com/agent/test"),
		WithTrustBundle(trustBundle),
		WithSVID(clientCert),
	)

	token, err := provider.AcquireToken(context.Background(), []string{"read:email"})
	if err != nil {
		t.Fatalf("AcquireToken failed: %v", err)
	}

	if token.AccessToken != "aims-token" {
		t.Errorf("expected access token aims-token, got %s", token.AccessToken)
	}
}

func TestGetProviderWithAIMS(t *testing.T) {
	agent := NewAgent("test-agent", WithProtocol(ProtocolAIMS))
	provider := agent.getProvider()

	if provider.Protocol() != ProtocolAIMS {
		t.Errorf("expected protocol aims, got %s", provider.Protocol())
	}
}
