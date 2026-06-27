// Example: Basic Agent Authorization
//
// This example demonstrates how to create an AI agent that makes
// authorized requests using the ID-JAG protocol.
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
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aistandardsio/oaiaf"
)

func main() {
	// Generate agent credentials (in production, these would be loaded from secure storage)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	// Create an agent with ID-JAG authorization
	agent := oaiaf.NewAgent("example-agent",
		oaiaf.WithName("Example AI Agent"),
		oaiaf.WithAuthServer("https://auth.example.com"),
		oaiaf.WithCredentials(privateKey, "agent-key-1"),
		oaiaf.WithProtocol(oaiaf.ProtocolIDJAG),
	)

	fmt.Printf("Agent ID: %s\n", agent.ID())
	fmt.Printf("Agent Name: %s\n", agent.Name())

	// Make an authorized request
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/user/profile", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Note: This will fail without a real auth server, but demonstrates the API
	fmt.Println("\nAttempting authorized request...")
	resp, err := agent.AuthorizedRequest(ctx, "read:profile", req)
	if err != nil {
		// Expected to fail in this example since there's no real auth server
		fmt.Printf("Request failed (expected in demo): %v\n", err)
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", string(body))
}
