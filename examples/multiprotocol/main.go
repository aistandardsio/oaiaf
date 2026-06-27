// Example: Multi-Protocol Agent Authorization
//
// This example demonstrates how to use multiple authorization protocols
// (ID-JAG, AAuth, AIMS) with the OAIAF framework.
//
// Run with:
//
//	go run main.go
package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aistandardsio/oaiaf"
)

func main() {
	// Generate agent credentials for ID-JAG
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	// Common auth server
	authServer := "https://auth.example.com"

	fmt.Println("=== Multi-Protocol Authorization Demo ===")
	fmt.Println()

	// Demonstrate ID-JAG (automated machine-to-machine auth)
	demonstrateIDJAG(authServer, privateKey)

	// Demonstrate AAuth (human-in-the-loop consent)
	demonstrateAAuth(authServer)

	// Demonstrate AIMS (workload identity with SPIFFE)
	demonstrateAIMS(authServer)

	fmt.Println()
	fmt.Println("=== Protocol Selection Demo ===")
	fmt.Println()
	demonstrateProtocolSelection(authServer, privateKey)
}

func demonstrateIDJAG(authServer string, privateKey *ecdsa.PrivateKey) {
	fmt.Println("--- ID-JAG Protocol ---")
	fmt.Println("Use case: Automated machine-to-machine authorization")
	fmt.Println("Authentication: JWT assertion signed by agent private key")
	fmt.Println()

	agent := oaiaf.NewAgent("automated-agent",
		oaiaf.WithName("Automated Processing Agent"),
		oaiaf.WithAuthServer(authServer),
		oaiaf.WithCredentials(privateKey, "agent-key-1"),
		oaiaf.WithProtocol(oaiaf.ProtocolIDJAG),
	)

	fmt.Printf("Agent: %s (%s)\n", agent.Name(), agent.ID())
	fmt.Printf("Protocol: ID-JAG\n")
	fmt.Println("Flow: Agent -> creates JWT assertion -> exchanges for access token")
	fmt.Println()

	// Note: This would succeed with a real auth server
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/data", nil)
	_, err := agent.AuthorizedRequest(ctx, "read:data", req)
	if err != nil {
		fmt.Printf("(Expected demo error: %v)\n\n", err)
	}
}

func demonstrateAAuth(authServer string) {
	fmt.Println("--- AAuth Protocol ---")
	fmt.Println("Use case: Operations requiring human consent")
	fmt.Println("Authentication: User approval via consent URI")
	fmt.Println()

	agent := oaiaf.NewAgent("consent-agent",
		oaiaf.WithName("Human Consent Agent"),
		oaiaf.WithAuthServer(authServer),
		oaiaf.WithProtocol(oaiaf.ProtocolAAuth),
	)

	fmt.Printf("Agent: %s (%s)\n", agent.Name(), agent.ID())
	fmt.Printf("Protocol: AAuth\n")
	fmt.Println("Flow: Agent -> requests authorization -> user consents -> token issued")
	fmt.Println()

	// Create provider with custom consent handler
	provider := oaiaf.NewAAuthProvider(agent)
	provider.ConsentHandler = func(consentURI string) (bool, error) {
		fmt.Printf("  [Consent Required] Please approve at: %s\n", consentURI)
		fmt.Println("  (In a real app, user would click approve)")
		return true, nil // Simulate approval
	}

	ctx := context.Background()
	_, err := provider.AcquireToken(ctx, []string{"payment:send"})
	if err != nil {
		fmt.Printf("(Expected demo error: %v)\n\n", err)
	}
}

func demonstrateAIMS(authServer string) {
	fmt.Println("--- AIMS Protocol ---")
	fmt.Println("Use case: Workload identity in cloud/container environments")
	fmt.Println("Authentication: SPIFFE SVID with mTLS")
	fmt.Println()

	// In production, you would load the SVID from SPIFFE workload API
	fmt.Printf("SPIFFE ID: spiffe://example.com/agent/worker\n")
	fmt.Printf("Protocol: AIMS\n")
	fmt.Println("Flow: Agent -> presents SVID via mTLS -> token issued")
	fmt.Println()

	agent := oaiaf.NewAgent("workload-agent",
		oaiaf.WithName("Workload Agent"),
		oaiaf.WithAuthServer(authServer),
		oaiaf.WithProtocol(oaiaf.ProtocolAIMS),
	)

	// Create AIMS provider with SPIFFE configuration
	trustBundle := x509.NewCertPool()
	provider := oaiaf.NewAIMSProvider(agent,
		oaiaf.WithSPIFFEID("spiffe://example.com/agent/worker"),
		oaiaf.WithTrustBundle(trustBundle),
	)

	fmt.Printf("SPIFFE ID: %s\n", provider.SPIFFEID())

	// In production with SPIFFE workload API:
	// socketPath := os.Getenv("SPIFFE_ENDPOINT_SOCKET")
	// provider := oaiaf.NewAIMSProvider(agent,
	//     oaiaf.WithWorkloadSocket(socketPath),
	// )
	// err := provider.FetchSVIDFromWorkloadAPI(ctx)

	ctx := context.Background()
	_, err := provider.AcquireToken(ctx, []string{"read:metrics"})
	if err != nil {
		fmt.Printf("(Expected demo error: %v)\n\n", err)
	}
}

func demonstrateProtocolSelection(authServer string, privateKey *ecdsa.PrivateKey) {
	fmt.Println("Dynamic Protocol Selection:")
	fmt.Println("Choose protocol based on operation type")
	fmt.Println()

	// Use protocol based on operation sensitivity
	operations := []struct {
		name     string
		scope    string
		protocol oaiaf.Protocol
		reason   string
	}{
		{
			name:     "Read public data",
			scope:    "read:public",
			protocol: oaiaf.ProtocolIDJAG,
			reason:   "Low sensitivity, automated OK",
		},
		{
			name:     "Send payment",
			scope:    "payment:send",
			protocol: oaiaf.ProtocolAAuth,
			reason:   "High sensitivity, needs human approval",
		},
		{
			name:     "Access service mesh",
			scope:    "mesh:access",
			protocol: oaiaf.ProtocolAIMS,
			reason:   "Workload-to-workload, uses SPIFFE",
		},
	}

	for _, op := range operations {
		agent := oaiaf.NewAgent("dynamic-agent",
			oaiaf.WithAuthServer(authServer),
			oaiaf.WithCredentials(privateKey, "key-1"),
			oaiaf.WithProtocol(op.protocol),
		)

		fmt.Printf("Operation: %s\n", op.name)
		fmt.Printf("  Scope: %s\n", op.scope)
		fmt.Printf("  Protocol: %s\n", op.protocol)
		fmt.Printf("  Reason: %s\n\n", op.reason)

		_ = agent // Used to show configuration
	}

	// Custom provider injection
	fmt.Println("Custom Provider Injection:")
	fmt.Println("Override default provider for custom auth flows")
	fmt.Println()

	customHTTPClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
		},
	}

	agent := oaiaf.NewAgent("custom-agent",
		oaiaf.WithAuthServer(authServer),
		oaiaf.WithHTTPClient(customHTTPClient),
	)

	// Create ID-JAG provider with custom configuration
	provider := oaiaf.NewIDJAGProvider(agent)

	// Inject provider into agent
	customAgent := oaiaf.NewAgent("injected-agent",
		oaiaf.WithAuthServer(authServer),
		oaiaf.WithProvider(provider),
	)

	fmt.Printf("Agent with injected provider: %s\n", customAgent.ID())
	fmt.Println("(Provider: custom ID-JAG instance)")

	// Environment-based protocol selection
	fmt.Println()
	fmt.Println()
	fmt.Println("Environment-Based Selection:")
	fmt.Println("Select protocol from environment variable")
	fmt.Println()

	protocolEnv := os.Getenv("AGENT_AUTH_PROTOCOL")
	if protocolEnv == "" {
		protocolEnv = "idjag" // Default
	}

	var protocol oaiaf.Protocol
	switch protocolEnv {
	case "aauth":
		protocol = oaiaf.ProtocolAAuth
	case "aims":
		protocol = oaiaf.ProtocolAIMS
	default:
		protocol = oaiaf.ProtocolIDJAG
	}

	fmt.Printf("AGENT_AUTH_PROTOCOL=%s -> Protocol: %s\n", protocolEnv, protocol)
}
