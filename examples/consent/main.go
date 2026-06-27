// Example: Human-in-the-Loop Consent Flow
//
// This example demonstrates how to use the AAuth protocol for
// operations that require human consent.
//
// Run with:
//
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aistandardsio/oaiaf"
)

func main() {
	// Create an agent with AAuth authorization (human-in-the-loop)
	agent := oaiaf.NewAgent("consent-agent",
		oaiaf.WithName("Consent Example Agent"),
		oaiaf.WithAuthServer("https://auth.example.com"),
		oaiaf.WithProtocol(oaiaf.ProtocolAAuth),
	)

	fmt.Printf("Agent ID: %s\n", agent.ID())
	fmt.Printf("Agent Name: %s\n", agent.Name())

	// Create a custom AAuth provider with a consent handler
	provider := oaiaf.NewAAuthProvider(agent)
	provider.ConsentHandler = func(consentURI string) (bool, error) {
		// In a real application, this would:
		// 1. Display the consent URI to the user
		// 2. Wait for user approval/denial
		// 3. Return the result

		fmt.Printf("\n=== CONSENT REQUIRED ===\n")
		fmt.Printf("Please visit the following URL to approve:\n")
		fmt.Printf("  %s\n", consentURI)
		fmt.Printf("========================\n\n")

		// For this example, we'll simulate approval
		fmt.Println("(Simulating user approval...)")
		return true, nil
	}

	// Make an authorized request that requires consent
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.example.com/payment", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	fmt.Println("\nAttempting authorized request (may require consent)...")

	// Using the custom provider directly
	token, err := provider.AcquireToken(ctx, []string{"payment:send"})
	if err != nil {
		// Expected to fail in this example since there's no real auth server
		fmt.Printf("Token acquisition failed (expected in demo): %v\n", err)
		return
	}

	// Add the token to the request
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	fmt.Printf("Got access token: %s...\n", token.AccessToken[:20])
	fmt.Printf("Token expires in: %d seconds\n", token.ExpiresIn)
}
