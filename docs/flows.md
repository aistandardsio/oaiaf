# OAIAF Protocol Flows

This document provides detailed sequence diagrams for the authorization protocols supported by OAIAF. Each flow is defined using [PIDL](https://github.com/grokify/pidl) (Protocol Interaction Description Language) and rendered as Mermaid sequence diagrams.

## Overview

OAIAF supports three core authorization protocols:

| Protocol | Use Case | Human Interaction |
|----------|----------|-------------------|
| [ID-JAG](#id-jag-token-exchange) | Automated, policy-based authorization | None |
| [AAuth](#aauth-consent-flow) | Sensitive operations requiring approval | Consent flow |
| [AIMS/SPIFFE](#aimsspiffe-authentication) | Workload identity binding | None |

These protocols can be combined to implement the [Five-Layer Agent Identity Stack](#agent-identity-stack).

## ID-JAG Token Exchange

ID-JAG (Identity Assertion Authorization Grant) enables agents to exchange signed JWT assertions for access tokens without human interaction. This is ideal for automated pipelines and trusted operations.

**Key Characteristics:**

- Agent signs JWT assertion with its private key
- Authorization server validates signature and agent registration
- No human consent required (policy-based)
- Tokens cached by scope for efficiency

**OAIAF Implementation:** `IDJAGProvider`

```mermaid
sequenceDiagram
    title ID-JAG Token Exchange Flow
    autonumber

    participant agent as OAIAF Agent
    participant auth_server as Authorization Server
    participant resource_server as Resource Server

    rect rgb(240, 240, 240)
    note right of agent: Assertion Creation
    agent->>agent: Create JWT Assertion
    end

    rect rgb(240, 240, 240)
    note right of agent: Token Exchange
    agent->>auth_server: POST /token<br/>grant_type=token-exchange<br/>subject_token={jwt}
    auth_server->>auth_server: Validate JWT Signature<br/>Check Agent Registration<br/>Evaluate Policies
    alt assertion_valid
        auth_server-->>agent: 200 OK<br/>{access_token, token_type, expires_in}
    else assertion_invalid
        auth_server-->>agent: 400 Bad Request<br/>{error: invalid_grant}
    end
    agent->>agent: Cache Token
    end

    rect rgb(240, 240, 240)
    note right of agent: Resource Access
    agent->>resource_server: GET /api/resource<br/>Authorization: Bearer {token}
    resource_server->>resource_server: Validate Token
    resource_server-->>agent: 200 OK<br/>{resource data}
    end
```

**Go Example:**

```go
agent := oaiaf.NewAgent("my-agent",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithCredentials(privateKey, keyID),
    oaiaf.WithProtocol(oaiaf.ProtocolIDJAG),
)

resp, err := agent.AuthorizedRequest(ctx, "read:email", req)
```

## AAuth Consent Flow

AAuth (Agent Authorization) provides human-in-the-loop consent for sensitive operations. The agent requests authorization, and a human must approve before access is granted.

**Key Characteristics:**

- Consent required for sensitive scopes
- Polling-based status updates
- Mission-scoped tokens
- Full audit trail

**OAIAF Implementation:** `AAuthProvider`

```mermaid
sequenceDiagram
    title AAuth Human-in-the-Loop Consent Flow
    autonumber

    participant agent as OAIAF Agent
    participant user as Human User
    participant auth_server as Authorization Server
    participant resource_server as Resource Server

    rect rgb(240, 240, 240)
    note right of agent: Authorization Request
    agent->>auth_server: POST /authorize<br/>{agent_token, scope, user_id}
    auth_server->>auth_server: Check Auto-Approval Policy
    auth_server-->>agent: 202 Accepted<br/>{consent_uri, status_uri, mission_id}
    end

    rect rgb(240, 240, 240)
    note right of agent: Consent Flow
    agent->>user: Display Consent URI<br/>(browser, notification, etc.)
    user->>auth_server: GET {consent_uri}<br/>Review Request Details
    auth_server-->>user: Consent Page<br/>(scope, agent info, mission details)
    alt user_approves
        user->>auth_server: POST /consent<br/>{approved: true}
    else user_denies
        user->>auth_server: POST /consent<br/>{approved: false}
    end
    end

    rect rgb(240, 240, 240)
    note right of agent: Token Issuance
    loop Poll until resolved
        agent->>auth_server: GET {status_uri}
        alt pending
            auth_server-->>agent: 200 OK<br/>{status: pending}
        else approved
            auth_server-->>agent: 200 OK<br/>{status: approved, access_token}
        else denied
            auth_server-->>agent: 200 OK<br/>{status: denied, error}
        end
    end
    end

    rect rgb(240, 240, 240)
    note right of agent: Resource Access
    agent->>resource_server: POST /api/action<br/>Authorization: Bearer {token}
    resource_server-->>agent: 200 OK<br/>{result}
    end
```

**Go Example:**

```go
provider := oaiaf.NewAAuthProvider(agent)
provider.ConsentHandler = func(consentURI string) (bool, error) {
    fmt.Printf("Approve at: %s\n", consentURI)
    return waitForUserApproval()
}
provider.Timeout = 10 * time.Minute

agent := oaiaf.NewAgent("my-agent",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithProvider(provider),
)

resp, err := agent.AuthorizedRequest(ctx, "mission:deploy:staging", req)
```

## AIMS/SPIFFE Authentication

AIMS (Agent Identity and Messaging System) uses SPIFFE workload identity to bind agents to their infrastructure. X.509 SVIDs provide cryptographic proof of workload identity.

**Key Characteristics:**

- Zero-trust workload identity
- Short-lived, auto-rotated certificates
- mTLS for transport security
- Defense in depth (mTLS + bearer token)

**OAIAF Implementation:** `AIMSProvider`

```mermaid
sequenceDiagram
    title AIMS/SPIFFE Workload Identity Authentication
    autonumber

    participant agent as OAIAF Agent
    participant spire_agent as SPIRE Agent
    participant auth_server as Authorization Server
    participant resource_server as Resource Server

    rect rgb(240, 240, 240)
    note right of agent: SVID Acquisition
    agent->>spire_agent: FetchX509SVID()<br/>via Workload API socket
    spire_agent->>spire_agent: Workload Attestation<br/>(PID, namespace, container)
    spire_agent-->>agent: X.509 SVID + Trust Bundle<br/>SPIFFE ID: spiffe://example.com/agent/x
    end

    rect rgb(240, 240, 240)
    note right of agent: Token Exchange
    agent->>auth_server: POST /token (mTLS)<br/>grant_type=token-exchange<br/>subject_token_type=svid
    auth_server->>auth_server: Validate SVID<br/>- Verify certificate chain<br/>- Check SPIFFE ID format<br/>- Match registration
    alt svid_valid
        auth_server-->>agent: 200 OK<br/>{access_token, token_type, scope}
    end
    end

    rect rgb(240, 240, 240)
    note right of agent: Resource Access
    agent->>resource_server: GET /api/resource (mTLS)<br/>Authorization: Bearer {token}
    resource_server->>resource_server: Validate mTLS + Token<br/>Match SPIFFE IDs
    resource_server-->>agent: 200 OK<br/>{resource data}
    end
```

**Go Example:**

```go
provider := oaiaf.NewAIMSProvider(agent,
    oaiaf.WithSPIFFEID("spiffe://example.com/agent/my-agent"),
    oaiaf.WithTrustBundle(trustBundle),
    oaiaf.WithSVID(certificate),
)

// Or fetch from SPIRE Workload API
provider := oaiaf.NewAIMSProvider(agent,
    oaiaf.WithWorkloadSocket("/var/run/spiffe/agent.sock"),
)

agent := oaiaf.NewAgent("my-agent",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithProvider(provider),
)
```

## Agent Identity Stack

The complete five-layer identity stack combines all protocols for enterprise-grade agent authorization:

1. **Lifecycle (SCIM)** - Agent provisioning and management
2. **Workload Identity (SPIFFE)** - Infrastructure binding
3. **Agent Authentication (AAuth)** - Agent identity token
4. **Human Delegation (ID-JAG)** - Authority chain
5. **Authorization (AuthZEN)** - Fine-grained access control

```mermaid
sequenceDiagram
    title OAIAF Five-Layer Agent Identity Stack
    autonumber

    participant admin as Enterprise Admin
    participant scim_server as SCIM Server
    participant user as Delegating User
    participant agent as OAIAF Agent
    participant spire as SPIRE
    participant auth_server as Authorization Server
    participant pdp as Policy Decision Point
    participant resource_server as Resource Server

    rect rgb(230, 245, 255)
    note right of admin: Layer 1: Lifecycle Management
    admin->>scim_server: POST /Agents<br/>{displayName, capabilities, owner}
    scim_server-->>admin: 201 Created<br/>{id, spiffeID, status: active}
    end

    rect rgb(230, 255, 230)
    note right of agent: Layer 2: Workload Identity
    agent->>spire: FetchX509SVID()
    spire-->>agent: X.509 SVID<br/>spiffe://example.com/agent/{id}
    end

    rect rgb(255, 245, 230)
    note right of user: Layer 4: Human Delegation
    user->>auth_server: OIDC Login
    auth_server-->>user: ID Token + Session
    end

    rect rgb(255, 230, 230)
    note right of agent: Layer 3: Agent Authentication
    agent->>auth_server: POST /token (mTLS)<br/>ID-JAG assertion<br/>delegator={user_id}
    auth_server->>auth_server: Validate:<br/>- SVID (workload)<br/>- Assertion (agent)<br/>- Delegation (user)
    auth_server-->>agent: Access Token<br/>{sub: agent, act: {sub: user}, scope}
    end

    rect rgb(240, 230, 255)
    note right of agent: Layer 5: Authorization
    agent->>resource_server: POST /api/sensitive-action<br/>Authorization: Bearer {token}
    resource_server->>pdp: POST /access/v1/evaluation<br/>{subject: {agent, user}, action, resource}
    pdp->>pdp: Evaluate Cedar/OpenFGA Policy<br/>- Agent capabilities<br/>- User permissions<br/>- Resource constraints
    alt allowed
        pdp-->>resource_server: {decision: PERMIT}
    end
    resource_server-->>agent: 200 OK<br/>{result}
    end
```

## A2A Agent Delegation

A2A (Agent-to-Agent) protocol enables agents to discover and delegate tasks to other agents while maintaining accountability.

```mermaid
sequenceDiagram
    title A2A Agent-to-Agent Delegation
    autonumber

    participant user as Human User
    participant orchestrator as Orchestrator Agent
    participant specialist as Specialist Agent
    participant auth_server as Authorization Server
    participant resource_server as Resource Server

    rect rgb(240, 240, 240)
    note right of user: Agent Discovery
    user->>orchestrator: Review PR #123 for security issues
    orchestrator->>specialist: GET /.well-known/agent.json
    specialist-->>orchestrator: Agent Card<br/>{capabilities: [security-scan],<br/>endpoints, auth}
    end

    rect rgb(240, 240, 240)
    note right of orchestrator: Task Delegation
    orchestrator->>auth_server: POST /token<br/>grant_type=delegation<br/>delegate_to={specialist_id}<br/>scope=security-scan:pr-123
    auth_server->>auth_server: Validate:<br/>- Orchestrator can delegate<br/>- Scope is subset<br/>- Specialist registered
    auth_server-->>orchestrator: Delegation Token<br/>{sub: specialist, act: [{orchestrator}, {user}]}
    end

    rect rgb(240, 240, 240)
    note right of specialist: Task Execution
    orchestrator->>specialist: POST /invoke<br/>{task: security-scan, target: pr-123}<br/>Authorization: Bearer {delegation_token}
    specialist->>resource_server: GET /repos/acme/backend/pulls/123/files<br/>Authorization: Bearer {delegation_token}
    resource_server->>resource_server: Validate delegation chain<br/>Log: user -> orchestrator -> specialist
    resource_server-->>specialist: 200 OK<br/>{files: [...]}
    specialist->>specialist: Security Analysis
    end

    rect rgb(240, 240, 240)
    note right of user: Task Completion
    specialist-->>orchestrator: 200 OK<br/>{findings: [...], risk_level: medium}
    orchestrator-->>user: Security Review Complete<br/>2 vulnerabilities found
    end
```

## MCP with OAIAF Authorization

MCP (Model Context Protocol) tool invocation integrated with OAIAF authorization.

```mermaid
sequenceDiagram
    title MCP Tool Invocation with OAIAF Authorization
    autonumber

    participant agent as OAIAF Agent
    participant mcp_server as MCP Server
    participant auth_server as Authorization Server
    participant external_api as External API

    rect rgb(240, 240, 240)
    note right of agent: MCP Initialization
    agent->>mcp_server: initialize<br/>{protocolVersion, capabilities}
    mcp_server-->>agent: initialize response<br/>{serverInfo, capabilities}
    agent->>mcp_server: tools/list
    mcp_server-->>agent: tools/list response<br/>[{name: search_code, inputSchema, auth_required: true}]
    end

    rect rgb(240, 240, 240)
    note right of agent: Tool Authentication
    agent->>auth_server: POST /token<br/>grant_type=client_credentials<br/>scope=mcp:search_code
    auth_server-->>agent: 200 OK<br/>{access_token, scope: mcp:search_code}
    end

    rect rgb(240, 240, 240)
    note right of agent: Tool Invocation
    agent->>mcp_server: tools/call<br/>{name: search_code, arguments: {query: "sql injection"}}<br/>Authorization: Bearer {token}
    mcp_server->>mcp_server: Validate Token<br/>Check scope: mcp:search_code
    end

    rect rgb(240, 240, 240)
    note right of mcp_server: Tool Execution
    mcp_server->>external_api: GET /api/search<br/>?q=sql+injection
    external_api-->>mcp_server: 200 OK<br/>{results: [...]}
    mcp_server-->>agent: tools/call response<br/>{content: [{type: text, text: "Found 3 matches..."}]}
    end
```

## PIDL Source Files

The diagrams in this document are generated from PIDL source files located in `docs/diagrams/pidl/`:

| File | Protocol |
|------|----------|
| [`idjag_token_exchange.pidl.json`](diagrams/pidl/idjag_token_exchange.pidl.json) | ID-JAG Token Exchange |
| [`aauth_consent_flow.pidl.json`](diagrams/pidl/aauth_consent_flow.pidl.json) | AAuth Consent Flow |
| [`aims_spiffe_auth.pidl.json`](diagrams/pidl/aims_spiffe_auth.pidl.json) | AIMS/SPIFFE Authentication |
| [`agent_identity_stack.pidl.json`](diagrams/pidl/agent_identity_stack.pidl.json) | Five-Layer Identity Stack |
| [`a2a_delegation.pidl.json`](diagrams/pidl/a2a_delegation.pidl.json) | A2A Agent Delegation |
| [`mcp_with_auth.pidl.json`](diagrams/pidl/mcp_with_auth.pidl.json) | MCP with OAIAF Auth |

To regenerate diagrams:

```bash
# Install PIDL
go install github.com/grokify/pidl/cmd/pidl@latest

# Generate Mermaid
pidl generate -f mermaid docs/diagrams/pidl/idjag_token_exchange.pidl.json

# Generate PlantUML
pidl generate -f plantuml docs/diagrams/pidl/idjag_token_exchange.pidl.json
```

## References

- [ID-JAG Specification](https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/)
- [AAuth Protocol](https://datatracker.ietf.org/doc/draft-hardt-oauth-aauth-protocol/)
- [SPIFFE](https://spiffe.io/)
- [A2A Protocol](https://github.com/a2a-protocol/a2a)
- [Model Context Protocol](https://spec.modelcontextprotocol.io/)
- [AuthZEN](https://openid.net/specs/openid-authzen-authorization-api-1_0.html)
- [PIDL](https://github.com/grokify/pidl)
