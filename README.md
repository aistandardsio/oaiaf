# OAIAF - OpenAI Agent Framework

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/aistandardsio/agent-protocols/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/aistandardsio/agent-protocols/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/aistandardsio/agent-protocols/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/aistandardsio/agent-protocols/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/aistandardsio/agent-protocols/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/aistandardsio/agent-protocols/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/aistandardsio/agent-protocols
 [goreport-url]: https://goreportcard.com/report/github.com/aistandardsio/agent-protocols
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/aistandardsio/agent-protocols
 [docs-godoc-url]: https://pkg.go.dev/github.com/aistandardsio/agent-protocols
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://aistandards.io/agent-protocols
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=aistandardsio%2Fagent-protocols
 [loc-svg]: https://tokei.rs/b1/github/grokify/agent-protocols
 [repo-url]: https://github.com/aistandardsio/agent-protocols
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/aistandardsio/agent-protocols/blob/main/LICENSE

OAIAF provides a framework for building AI agents with built-in support for agent authorization protocols including ID-JAG, AAuth, and AIMS.

## Features

- **Protocol Support** - Built-in support for ID-JAG, AAuth, and AIMS protocols
- **Token Management** - Automatic token caching and renewal
- **Human-in-the-Loop** - Support for consent flows when required
- **HTTP Integration** - Easy-to-use HTTP client with automatic authorization

## Installation

```bash
go get github.com/aistandardsio/oaiaf
```

## Quick Start

```go
package main

import (
    "context"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "log"
    "net/http"

    "github.com/aistandardsio/oaiaf"
)

func main() {
    // Generate or load credentials
    privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    keyID := "agent-key-1"

    // Create an agent
    agent := oaiaf.NewAgent("my-agent",
        oaiaf.WithName("My AI Agent"),
        oaiaf.WithAuthServer("https://auth.example.com"),
        oaiaf.WithCredentials(privateKey, keyID),
    )

    // Make an authorized request
    ctx := context.Background()
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/user/email", nil)

    resp, err := agent.AuthorizedRequest(ctx, "read:email", req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    log.Printf("Response status: %s", resp.Status)
}
```

## Supported Protocols

### ID-JAG (Identity Assertion Authorization Grant)

Automated, policy-based authorization for trusted operations:

```go
agent := oaiaf.NewAgent("my-agent",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithProtocol(oaiaf.ProtocolIDJAG),
)
```

### AAuth (Agent Authorization)

Human-in-the-loop consent for sensitive operations:

```go
agent := oaiaf.NewAgent("my-agent",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithProtocol(oaiaf.ProtocolAAuth),
)
```

### AIMS (Agent Identity and Messaging System)

Workload identity using SPIFFE-based authentication:

```go
agent := oaiaf.NewAgent("my-agent",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithProtocol(oaiaf.ProtocolAIMS),
)
```

## Token Management

OAIAF automatically handles token acquisition, caching, and renewal:

```go
// Tokens are automatically cached
resp1, _ := agent.AuthorizedRequest(ctx, "read:email", req1)  // Acquires token
resp2, _ := agent.AuthorizedRequest(ctx, "read:email", req2)  // Uses cached token

// Clear the cache if needed
agent.ClearTokenCache()
```

## Examples

See the [examples](examples/) directory for complete working examples:

- [basic](examples/basic/) - Basic ID-JAG authorization
- [consent](examples/consent/) - AAuth human-in-the-loop consent flow
- [multiprotocol](examples/multiprotocol/) - Using multiple protocols dynamically

## API Reference

### Agent Options

| Option | Description |
|--------|-------------|
| `WithName(name)` | Set agent display name |
| `WithAuthServer(url)` | Authorization server URL |
| `WithCredentials(key, keyID)` | Signing credentials for ID-JAG |
| `WithProtocol(protocol)` | Default protocol (ProtocolIDJAG, ProtocolAAuth, ProtocolAIMS) |
| `WithHTTPClient(client)` | Custom HTTP client |
| `WithProvider(provider)` | Custom authorization provider |

### Provider Interface

Custom providers can be created by implementing the `Provider` interface:

```go
type Provider interface {
    Protocol() Protocol
    AcquireToken(ctx context.Context, scopes []string) (*TokenResponse, error)
}
```

### AIMS/SPIFFE Configuration

For workload identity with SPIFFE:

```go
provider := oaiaf.NewAIMSProvider(agent,
    oaiaf.WithSPIFFEID("spiffe://example.com/agent/my-agent"),
    oaiaf.WithTrustBundle(trustBundle),
    oaiaf.WithSVID(certificate),
)

// Or fetch from workload API
provider := oaiaf.NewAIMSProvider(agent,
    oaiaf.WithWorkloadSocket("/var/run/spiffe/agent.sock"),
)
err := provider.FetchSVIDFromWorkloadAPI(ctx)
```

## Documentation

- [Agent Protocols](https://github.com/aistandardsio/agent-protocols) - Protocol implementations
- [ID-JAG Spec](https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/) - ID-JAG specification
- [AAuth Spec](https://datatracker.ietf.org/doc/draft-hardt-oauth-aauth-protocol/) - AAuth specification

## License

MIT License - see [LICENSE](LICENSE) for details.
