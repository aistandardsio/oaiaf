# OAIAF Architecture

This document describes the architecture of the OpenAI Agent Framework (OAIAF) and how identity, authorization, interoperability, and observability standards fit together for enterprise AI agent management.

## Executive Summary

OAIAF provides a unified Go framework for building AI agents with enterprise-grade identity and authorization. It implements a layered architecture that addresses the fundamental questions enterprises face when deploying autonomous AI agents:

- **Who is this agent?** (Identity)
- **What workload is hosting it?** (Workload Identity)
- **Who authorized it to act?** (Human Delegation)
- **What is it allowed to do?** (Authorization)
- **How do we observe its behavior?** (Observability)

OAIAF achieves this by integrating emerging standards from IETF OAuth Working Group, SPIFFE/SPIRE, Linux Foundation, and OpenTelemetry communities into a cohesive framework.

For detailed sequence diagrams of each protocol flow, see [Protocol Flows](flows.md).

## The Agent Identity Stack

Enterprise AI agents require a multi-layered identity model. OAIAF implements a five-layer stack where each layer builds on the ones below:

```
┌─────────────────────────────────────────────────────────────────┐
│                       AUTHORIZATION                              │
│            AuthZEN API │ Cedar │ OpenFGA                        │
│         "What can this agent+human+workload do?"                │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                    HUMAN DELEGATION                              │
│                   OAuth 2.x + ID-JAG                            │
│        "Which user delegated authority to this agent?"          │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                  AGENT AUTHENTICATION                            │
│                        AAuth                                     │
│             "Which autonomous agent is this?"                   │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                   WORKLOAD IDENTITY                              │
│                   WIMSE / SPIFFE                                │
│       "Which workload/service is hosting this agent?"           │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                  LIFECYCLE MANAGEMENT                            │
│                  SCIM Agent Resource                            │
│      "What agents exist and what are their capabilities?"       │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Interactions

Each layer serves a distinct purpose and interacts with adjacent layers:

| Layer | Question Answered | Input From Below | Output To Above |
|-------|-------------------|------------------|-----------------|
| Authorization | What actions are permitted? | Agent + Human + Workload context | Access decision |
| Human Delegation | Who authorized the agent? | Authenticated agent | Delegation claims |
| Agent Authentication | Which agent is this? | Workload identity | Agent identity token |
| Workload Identity | Which service hosts this? | Lifecycle registration | SVID/workload token |
| Lifecycle | What agents exist? | Enterprise provisioning | Agent metadata |

## Layer 1: Lifecycle Management (SCIM Agent Resource)

The foundation layer tracks agent existence, capabilities, and metadata using SCIM (System for Cross-domain Identity Management) extended with agent-specific attributes.

### Purpose

- Provision and deprovision agents across enterprise systems
- Store agent capabilities, owners, and organizational metadata
- Enable discovery of registered agents
- Support compliance and audit requirements

### Specification

[draft-wzdk-scim-agent-resource](https://datatracker.ietf.org/doc/draft-wzdk-scim-agent-resource/) - Extends SCIM with an Agent resource type.

### Agent Resource Schema

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:Agent"],
  "id": "agent-123",
  "displayName": "Code Review Agent",
  "agentType": "autonomous",
  "capabilities": ["code-review", "security-scan"],
  "owner": {
    "value": "user-456",
    "$ref": "https://scim.example.com/Users/user-456"
  },
  "trustDomain": "example.com",
  "registeredAt": "2024-01-15T10:30:00Z",
  "status": "active"
}
```

### OAIAF Integration

OAIAF agents can be provisioned via SCIM and retrieve their configuration:

```go
// Future SCIM client integration
agent := oaiaf.NewAgent("agent-123",
    oaiaf.WithSCIMEndpoint("https://scim.example.com"),
    oaiaf.WithSCIMToken(scimToken),
)

// Agent fetches its own configuration from SCIM
if err := agent.FetchSCIMConfig(ctx); err != nil {
    log.Fatal(err)
}
```

## Layer 2: Workload Identity (WIMSE/SPIFFE)

This layer establishes the identity of the computing workload (container, VM, or process) that hosts the agent.

### Purpose

- Bind agent identity to infrastructure identity
- Enable zero-trust networking with cryptographic attestation
- Support workload-to-workload authentication
- Provide short-lived, automatically rotated credentials

### Specifications

- [WIMSE Architecture](https://datatracker.ietf.org/doc/draft-ietf-wimse-architecture/) - IETF Workload Identity in Multi-System Environments
- [SPIFFE](https://spiffe.io/) - Secure Production Identity Framework for Everyone

### SPIFFE ID Format

```
spiffe://trust-domain/path/to/workload
spiffe://example.com/agent/code-review
spiffe://prod.acme.io/ns/ai-agents/sa/reviewer
```

### X.509 SVID Structure

SPIFFE Verifiable Identity Documents (SVIDs) embed the SPIFFE ID in the certificate's URI SAN:

```
Subject: CN=code-review-agent
URI SAN: spiffe://example.com/agent/code-review
Issuer: CN=SPIRE Server CA
Validity: 1 hour (auto-rotated)
```

### OAIAF Integration

OAIAF's `AIMSProvider` implements SPIFFE-based workload identity:

```go
// Option 1: Provide SVID directly
provider := oaiaf.NewAIMSProvider(agent,
    oaiaf.WithSPIFFEID("spiffe://example.com/agent/code-review"),
    oaiaf.WithTrustBundle(trustBundle),
    oaiaf.WithSVID(certificate),
)

// Option 2: Fetch from SPIRE Workload API
provider := oaiaf.NewAIMSProvider(agent,
    oaiaf.WithWorkloadSocket("/var/run/spiffe/agent.sock"),
)
err := provider.FetchSVIDFromWorkloadAPI(ctx)

// Use with agent
agent := oaiaf.NewAgent("code-review",
    oaiaf.WithProvider(provider),
)
```

### Kubernetes Integration

In Kubernetes environments, SPIFFE IDs can be derived from pod identity:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: code-review-agent
  labels:
    spiffe.io/spiffe-id: "spiffe://cluster.local/ns/ai-agents/sa/reviewer"
spec:
  serviceAccountName: reviewer
  containers:
    - name: agent
      volumeMounts:
        - name: spiffe-workload-api
          mountPath: /var/run/spiffe
  volumes:
    - name: spiffe-workload-api
      csi:
        driver: csi.spiffe.io
```

## Layer 3: Agent Authentication (AAuth)

This layer authenticates the agent as an autonomous entity, distinct from both the workload and the human user.

### Purpose

- Authenticate agents independently from users
- Support mission-based authorization scopes
- Enable human-in-the-loop consent flows
- Track agent actions for audit and compliance

### Specification

[draft-hardt-oauth-aauth-protocol](https://datatracker.ietf.org/doc/draft-hardt-oauth-aauth-protocol/) - Agent Authorization Protocol

### AAuth Flow

```
┌─────────┐                              ┌─────────────┐
│  Agent  │                              │  Auth Server │
└────┬────┘                              └──────┬──────┘
     │                                          │
     │  1. Authorization Request                │
     │  POST /authorize                         │
     │  {agent_token, scope, user_id}           │
     │─────────────────────────────────────────▶│
     │                                          │
     │  2a. Immediate Approval (policy-based)   │
     │  {access_token, token_type, expires_in}  │
     │◀─────────────────────────────────────────│
     │                                          │
     │  2b. Consent Required                    │
     │  {consent_uri, status_uri, mission_id}   │
     │◀─────────────────────────────────────────│
     │                                          │
     │  3. Poll Status (if consent required)    │
     │  GET {status_uri}                        │
     │─────────────────────────────────────────▶│
     │                                          │
     │  4. Token (after human approval)         │
     │  {access_token, scope}                   │
     │◀─────────────────────────────────────────│
```

### Mission Scopes

AAuth introduces "missions" - bounded task scopes that define what an agent can do:

```
mission:code-review:pr-123        # Review a specific PR
mission:deploy:staging            # Deploy to staging environment
mission:data-analysis:q4-report   # Analyze Q4 data
```

### OAIAF Integration

OAIAF's `AAuthProvider` implements the full consent flow:

```go
// Create AAuth provider with consent handler
provider := oaiaf.NewAAuthProvider(agent)
provider.ConsentHandler = func(consentURI string) (bool, error) {
    // Display consent URI to user (e.g., open browser, send notification)
    fmt.Printf("Please approve at: %s\n", consentURI)

    // Wait for user action (simplified)
    return waitForUserApproval()
}
provider.Timeout = 10 * time.Minute

agent := oaiaf.NewAgent("code-review",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithProvider(provider),
)

// Request with mission scope
req, _ := http.NewRequestWithContext(ctx, "POST",
    "https://api.example.com/reviews", body)
resp, err := agent.AuthorizedRequest(ctx, "mission:code-review:pr-123", req)
```

## Layer 4: Human Delegation (OAuth + ID-JAG)

This layer connects agent actions to human authority through delegation assertions.

### Purpose

- Establish chain of authority from human to agent
- Enable policy-based automated authorization
- Support compliance requirements for human oversight
- Allow agents to act on behalf of users

### Specifications

- [OAuth 2.1](https://datatracker.ietf.org/doc/draft-ietf-oauth-v2-1/) - Core authorization framework
- [ID-JAG](https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/) - Identity Assertion Authorization Grant

### ID-JAG Token Exchange

ID-JAG enables agents to exchange signed assertions for access tokens:

```
┌─────────┐                              ┌─────────────┐
│  Agent  │                              │  Auth Server │
└────┬────┘                              └──────┬──────┘
     │                                          │
     │  1. Token Exchange Request               │
     │  POST /token                             │
     │  grant_type=token-exchange               │
     │  subject_token={signed_jwt}              │
     │  subject_token_type=jwt                  │
     │  scope=read:email                        │
     │─────────────────────────────────────────▶│
     │                                          │
     │  2. Validate JWT signature               │
     │     Check agent registration             │
     │     Evaluate delegation policies         │
     │                                          │
     │  3. Access Token                         │
     │  {access_token, token_type, scope}       │
     │◀─────────────────────────────────────────│
```

### Assertion JWT Structure

```json
{
  "iss": "agent-code-review",
  "sub": "agent-code-review",
  "aud": "https://auth.example.com",
  "iat": 1704067200,
  "exp": 1704067500,
  "jti": "unique-request-id",
  "scope": "read:email write:pr-comments",
  "delegator": "user-456",
  "delegation_context": {
    "purpose": "code-review",
    "constraints": ["read-only", "single-repo"]
  }
}
```

### OAIAF Integration

OAIAF's `IDJAGProvider` handles JWT assertion creation and token exchange:

```go
// Generate or load agent credentials
privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
keyID := "agent-key-2024"

// Create agent with ID-JAG protocol
agent := oaiaf.NewAgent("code-review",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithCredentials(privateKey, keyID),
    oaiaf.WithProtocol(oaiaf.ProtocolIDJAG),
)

// Make authorized request - ID-JAG assertion created automatically
req, _ := http.NewRequestWithContext(ctx, "GET",
    "https://api.example.com/user/email", nil)
resp, err := agent.AuthorizedRequest(ctx, "read:email", req)
```

### Supported Key Types

OAIAF supports multiple signing algorithms:

| Key Type | JWT Algorithm | Use Case |
|----------|---------------|----------|
| ECDSA P-256 | ES256 | Default, good performance |
| RSA 2048+ | RS256 | Legacy compatibility |
| Ed25519 | EdDSA | Modern, compact signatures |

## Layer 5: Authorization (AuthZEN, Cedar, OpenFGA)

The top layer makes access control decisions combining agent, human, and workload identity with resource policies.

### Purpose

- Centralized policy decision point (PDP)
- Fine-grained attribute-based access control (ABAC)
- Relationship-based access control (ReBAC)
- Audit trail for all access decisions

### Specifications and Implementations

| Standard | Type | Description |
|----------|------|-------------|
| [AuthZEN](https://openid.net/specs/openid-authzen-authorization-api-1_0.html) | API Standard | PEP-PDP communication protocol |
| [Cedar](https://www.cedarpolicy.com/) | Policy Language | AWS-backed ABAC policy language |
| [OpenFGA](https://openfga.dev/) | Service | Google Zanzibar-inspired ReBAC |

### AuthZEN API

AuthZEN standardizes how Policy Enforcement Points (PEPs) query Policy Decision Points (PDPs):

```http
POST /access/v1/evaluation HTTP/1.1
Content-Type: application/json

{
  "subject": {
    "type": "agent",
    "id": "code-review",
    "properties": {
      "workload_id": "spiffe://example.com/agent/code-review",
      "delegator": "user-456"
    }
  },
  "resource": {
    "type": "repository",
    "id": "acme/backend",
    "properties": {
      "visibility": "private"
    }
  },
  "action": {
    "name": "comment",
    "properties": {
      "pr_number": 123
    }
  },
  "context": {
    "time": "2024-01-15T10:30:00Z",
    "mission": "code-review:pr-123"
  }
}
```

### Cedar Policy Example

```cedar
// Allow agents to comment on PRs when delegated by repo maintainer
permit(
    principal is Agent,
    action == Action::"comment",
    resource is PullRequest
) when {
    // Agent must have valid delegation
    principal.delegator in resource.repository.maintainers &&
    // Mission must match the PR
    context.mission == "code-review:pr-" + resource.pr_number &&
    // Agent must have code-review capability
    "code-review" in principal.capabilities
};

// Deny agents from approving PRs (human-only)
forbid(
    principal is Agent,
    action == Action::"approve",
    resource is PullRequest
);
```

### OpenFGA Model Example

```yaml
model
  schema 1.1

type user

type agent
  relations
    define delegator: [user]
    define capability: [capability]

type capability

type repository
  relations
    define maintainer: [user]
    define reader: [user, agent]
    define commenter: [agent] and delegator from maintainer

type pull_request
  relations
    define parent: [repository]
    define can_comment: commenter from parent
```

### OAIAF Integration (Future)

```go
// Future AuthZEN integration
authzClient := authzen.NewClient("https://pdp.example.com")

agent := oaiaf.NewAgent("code-review",
    oaiaf.WithAuthZENClient(authzClient),
)

// Authorization check included in request flow
resp, err := agent.AuthorizedRequest(ctx, "comment:pr-123", req)
// Internally calls AuthZEN before making request
```

## Interoperability Protocols

Beyond identity and authorization, agents need protocols to discover and communicate with each other and external tools.

### Protocol Landscape

```
┌───────────────────────────────────────────────────────────────┐
│                      AGENT ECOSYSTEM                           │
│                                                                │
│  ┌──────────────┐                        ┌──────────────┐     │
│  │   Agent A    │◀──── A2A Protocol ────▶│   Agent B    │     │
│  │ (Claude Code)│      (discovery,       │  (CrewAI)    │     │
│  └──────┬───────┘       delegation)      └──────┬───────┘     │
│         │                                       │              │
│         │ MCP                                   │ MCP          │
│         ▼                                       ▼              │
│  ┌──────────────┐                        ┌──────────────┐     │
│  │  MCP Server  │                        │  MCP Server  │     │
│  │   (Tools)    │                        │ (Resources)  │     │
│  └──────────────┘                        └──────────────┘     │
└───────────────────────────────────────────────────────────────┘
```

### A2A (Agent-to-Agent Protocol)

**Governance**: Linux Foundation

**Purpose**: Enables agents to discover, authenticate, and delegate to other agents.

**Key Features**:

- Agent discovery via well-known endpoints
- Capability negotiation
- Task delegation with accountability
- Multi-agent orchestration

**Agent Card** (Discovery Document):

```json
{
  "id": "code-review-agent",
  "name": "Code Review Agent",
  "description": "Automated code review with security scanning",
  "version": "1.0.0",
  "capabilities": [
    {
      "id": "review-pr",
      "description": "Review pull request for issues",
      "input_schema": {...},
      "output_schema": {...}
    }
  ],
  "authentication": {
    "type": "bearer",
    "token_endpoint": "https://auth.example.com/token"
  },
  "endpoints": {
    "invoke": "https://agent.example.com/invoke",
    "status": "https://agent.example.com/status/{task_id}"
  }
}
```

**Discovery**:

```http
GET /.well-known/agent.json HTTP/1.1
Host: agent.example.com
```

### MCP (Model Context Protocol)

**Governance**: Agentic AI Foundation (OpenAI, Anthropic, Microsoft, Google)

**Purpose**: Standardizes how agents interact with tools and resources.

**Key Features**:

- Tool discovery and invocation
- Resource access (files, APIs, databases)
- Prompt templates
- Sampling (LLM access for tools)

**MCP Architecture**:

```
┌────────────┐         ┌────────────┐         ┌────────────┐
│   Agent    │◀──────▶│ MCP Client │◀──────▶│ MCP Server │
│  (Host)    │  JSON   │            │  JSON   │  (Tools)   │
└────────────┘  RPC    └────────────┘  RPC    └────────────┘
```

**Tool Definition**:

```json
{
  "name": "search_code",
  "description": "Search codebase for patterns",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {"type": "string"},
      "file_pattern": {"type": "string"}
    },
    "required": ["query"]
  }
}
```

**OAIAF + MCP Integration**:

```go
// Agent with MCP server for tool access
agent := oaiaf.NewAgent("code-review",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithCredentials(privateKey, keyID),
)

// MCP client uses agent's authorization
mcpClient := mcp.NewClient(
    mcp.WithTransport(mcp.StdioTransport()),
    mcp.WithAuthProvider(agent), // Reuse OAIAF auth
)

// Tool invocation includes authorization
result, err := mcpClient.CallTool(ctx, "search_code", map[string]any{
    "query": "sql injection",
})
```

## Observability

Enterprise AI agents require comprehensive observability for debugging, compliance, and optimization.

### OpenTelemetry GenAI Semantic Conventions

OpenTelemetry defines standard attributes for AI/ML workloads:

**Trace Attributes** (`gen_ai.*`):

| Attribute | Type | Description |
|-----------|------|-------------|
| `gen_ai.system` | string | AI system (e.g., "openai", "anthropic") |
| `gen_ai.request.model` | string | Model identifier |
| `gen_ai.request.max_tokens` | int | Token limit |
| `gen_ai.response.finish_reasons` | string[] | Completion reasons |
| `gen_ai.usage.input_tokens` | int | Prompt tokens |
| `gen_ai.usage.output_tokens` | int | Completion tokens |

**Agent-Specific Attributes**:

| Attribute | Type | Description |
|-----------|------|-------------|
| `agent.id` | string | OAIAF agent identifier |
| `agent.mission` | string | Current mission scope |
| `agent.delegator` | string | Human who delegated |
| `agent.workload_id` | string | SPIFFE ID |

**Example Trace**:

```go
ctx, span := tracer.Start(ctx, "agent.authorized_request",
    trace.WithAttributes(
        attribute.String("agent.id", agent.ID()),
        attribute.String("agent.mission", scope),
        attribute.String("agent.protocol", string(protocol)),
        attribute.String("gen_ai.system", "anthropic"),
        attribute.String("gen_ai.request.model", "claude-3-opus"),
    ),
)
defer span.End()
```

### AgentOps

AgentOps extends OpenTelemetry with agent-specific monitoring:

**Key Metrics**:

- **Session replay**: Full conversation and action history
- **Tool usage**: Which tools, how often, success rate
- **Authorization events**: Token acquisitions, consent flows
- **Error rates**: By operation type, resource, mission

**Integration**:

```go
// Future AgentOps integration
agent := oaiaf.NewAgent("code-review",
    oaiaf.WithAgentOps(agentops.Config{
        APIKey:     os.Getenv("AGENTOPS_API_KEY"),
        ProjectID:  "my-project",
        SessionTags: map[string]string{"env": "production"},
    }),
)
```

## Runtime Infrastructure

### Kubernetes Workload Identity

Kubernetes 1.35+ includes native workload identity with SPIFFE:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: code-review-agent
  annotations:
    # Projected SPIFFE ID
    spiffe.io/identity: "spiffe://cluster.local/ns/ai-agents/sa/code-review"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: code-review-agent
spec:
  template:
    spec:
      serviceAccountName: code-review-agent
      containers:
        - name: agent
          image: acme/code-review-agent:v1
          env:
            - name: SPIFFE_ENDPOINT_SOCKET
              value: "unix:///var/run/spiffe/agent.sock"
          volumeMounts:
            - name: spiffe-workload-api
              mountPath: /var/run/spiffe
              readOnly: true
      volumes:
        - name: spiffe-workload-api
          projected:
            sources:
              - serviceAccountToken:
                  path: token
                  expirationSeconds: 3600
                  audience: spiffe://cluster.local
```

### SPIRE

SPIRE (SPIFFE Runtime Environment) provides the workload identity infrastructure:

**Architecture**:

```
┌─────────────────────────────────────────────────────────────┐
│                    SPIRE Server                              │
│  - Certificate Authority                                     │
│  - Registration entries                                      │
│  - Trust bundle management                                   │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
┌────────▼────────┐    ┌────────▼────────┐
│   SPIRE Agent   │    │   SPIRE Agent   │
│   (Node 1)      │    │   (Node 2)      │
│                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │  Workload   │ │    │ │  Workload   │ │
│ │  (Agent)    │ │    │ │  (Agent)    │ │
│ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘
```

**Registration Entry**:

```bash
spire-server entry create \
    -spiffeID spiffe://example.com/agent/code-review \
    -parentID spiffe://example.com/node/k8s-node-1 \
    -selector k8s:ns:ai-agents \
    -selector k8s:sa:code-review
```

### Service Mesh (Istio)

Istio can enforce authorization policies based on SPIFFE identity:

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-code-review-agent
  namespace: backend
spec:
  action: ALLOW
  rules:
    - from:
        - source:
            principals:
              - "spiffe://cluster.local/ns/ai-agents/sa/code-review"
      to:
        - operation:
            methods: ["GET", "POST"]
            paths: ["/api/reviews/*", "/api/comments/*"]
    - from:
        - source:
            principals:
              - "spiffe://cluster.local/ns/ai-agents/sa/code-review"
      to:
        - operation:
            methods: ["GET"]
            paths: ["/api/repos/*/files/*"]
```

## Agent Type Reference

Different agent deployment patterns have different identity and authorization requirements.

### Reference Matrix

| Agent Type | Examples | Identity | Auth | Delegation | Interop | Runtime |
|------------|----------|----------|------|------------|---------|---------|
| Coding Assistants | Claude Code, Codex CLI | SPIFFE | AAuth | ID-JAG | MCP | Local/Container |
| Hosted Platforms | OpenClaw, OmniAgent | SCIM+SPIFFE | AAuth+OAuth | ID-JAG | A2A+MCP | Kubernetes |
| Single-Agent Orchestration | LangChain | SPIFFE | AAuth | ID-JAG | MCP | Various |
| Multi-Agent Orchestration | CrewAI, Google ADK | SCIM+SPIFFE | AAuth | ID-JAG | A2A+MCP | Kubernetes |
| Enterprise Frameworks | Microsoft Agent Framework | SCIM+SPIFFE | AAuth+OAuth | ID-JAG | A2A+MCP | Azure/K8s |

### Coding Assistants

**Examples**: Claude Code, OpenAI Codex CLI, GitHub Copilot CLI

**Characteristics**:

- Run locally on developer machines or in containers
- Interactive sessions with human developers
- Need access to local files, git, and APIs
- Session-scoped authorization

**Identity Requirements**:

- Workload identity from local SPIRE agent or container runtime
- Human delegation via interactive OAuth flow
- MCP for tool integration

**OAIAF Example**:

```go
// Coding assistant with interactive consent
agent := oaiaf.NewAgent("claude-code",
    oaiaf.WithAuthServer("https://auth.anthropic.com"),
    oaiaf.WithProtocol(oaiaf.ProtocolAAuth),
)

// Set up interactive consent handler
provider := oaiaf.NewAAuthProvider(agent)
provider.ConsentHandler = func(consentURI string) (bool, error) {
    // Open browser for user approval
    return openBrowserAndWait(consentURI)
}

// Request with session scope
resp, err := agent.AuthorizedRequest(ctx, "session:code-edit", req)
```

### Hosted Platforms

**Examples**: OpenClaw, OmniAgent, Fixie

**Characteristics**:

- Multi-tenant cloud platforms
- Manage many agents for many organizations
- Require strong isolation and audit
- API-driven agent creation

**Identity Requirements**:

- SCIM for agent provisioning and lifecycle
- SPIFFE for workload identity within platform
- OAuth for platform-level authentication
- ID-JAG for delegated agent actions

**OAIAF Example**:

```go
// Platform-managed agent with SCIM provisioning
agent := oaiaf.NewAgent("tenant-123-agent-456",
    oaiaf.WithAuthServer("https://platform.example.com/oauth"),
    oaiaf.WithProtocol(oaiaf.ProtocolIDJAG),
    oaiaf.WithCredentials(platformKey, keyID),
)

// Platform adds tenant context
agent.SetMetadata("tenant_id", "tenant-123")
agent.SetMetadata("organization", "Acme Corp")
```

### Multi-Agent Orchestration

**Examples**: CrewAI, Google ADK, Microsoft AutoGen

**Characteristics**:

- Multiple agents collaborating on tasks
- Agents delegate to other agents
- Complex authorization chains
- Need discovery and capability negotiation

**Identity Requirements**:

- A2A for agent-to-agent communication
- SCIM for agent registration and discovery
- Per-agent SPIFFE identity
- Delegation chains tracked via ID-JAG

**OAIAF Example**:

```go
// Orchestrator agent that delegates to specialists
orchestrator := oaiaf.NewAgent("orchestrator",
    oaiaf.WithAuthServer("https://auth.example.com"),
    oaiaf.WithProtocol(oaiaf.ProtocolAAuth),
)

// Discover specialist agent via A2A
specialist, err := a2a.DiscoverAgent("https://specialist.example.com")
if err != nil {
    log.Fatal(err)
}

// Delegate with mission scope
delegationToken, err := orchestrator.DelegateToAgent(ctx, specialist.ID,
    oaiaf.WithMission("code-review:pr-123"),
    oaiaf.WithConstraints("read-only"),
)

// Specialist invokes with delegation
resp, err := specialist.InvokeWithDelegation(ctx, delegationToken, task)
```

### Enterprise Frameworks

**Examples**: Microsoft Agent Framework, Salesforce Agentforce

**Characteristics**:

- Deep integration with enterprise systems (AD, SAP, Salesforce)
- Strict compliance requirements (SOX, HIPAA, GDPR)
- Complex approval workflows
- Audit logging requirements

**Identity Requirements**:

- All five layers fully implemented
- Integration with enterprise IdP (Entra ID, Okta)
- Fine-grained authorization (Cedar/OpenFGA)
- Comprehensive observability

**OAIAF Example**:

```go
// Enterprise agent with full identity stack
agent := oaiaf.NewAgent("enterprise-agent",
    // Layer 1: SCIM lifecycle
    oaiaf.WithSCIMEndpoint("https://scim.enterprise.com"),

    // Layer 2: Workload identity
    oaiaf.WithWorkloadSocket("/var/run/spiffe/agent.sock"),

    // Layer 3-4: Agent auth + delegation
    oaiaf.WithAuthServer("https://auth.enterprise.com"),
    oaiaf.WithCredentials(key, keyID),
    oaiaf.WithProtocol(oaiaf.ProtocolIDJAG),

    // Layer 5: Authorization
    oaiaf.WithAuthZENClient(authzClient),

    // Observability
    oaiaf.WithTracer(tracer),
    oaiaf.WithAgentOps(agentopsConfig),
)

// All requests include full context
resp, err := agent.AuthorizedRequest(ctx,
    "mission:expense-approval:req-789", req)
```

## OAIAF Implementation Mapping

The following table maps OAIAF components to standards:

| OAIAF Component | Standard | Description |
|-----------------|----------|-------------|
| `Agent` | - | Core agent abstraction |
| `IDJAGProvider` | ID-JAG | Token exchange with JWT assertions |
| `AAuthProvider` | AAuth | Human-in-the-loop consent |
| `AIMSProvider` | SPIFFE/WIMSE | Workload identity with mTLS |
| `Provider` interface | - | Extensible provider pattern |
| `TokenResponse` | OAuth 2.1 | Standard token structure |

### Protocol Selection Guide

| Scenario | Recommended Protocol | Rationale |
|----------|---------------------|-----------|
| Automated pipelines | ID-JAG | No human interaction needed |
| Sensitive operations | AAuth | Human approval required |
| Container/K8s workloads | AIMS | Workload identity built-in |
| Interactive sessions | AAuth + ID-JAG | Initial consent, then automated |

### Token Flow Example

```
User ──▶ OIDC Login ──▶ ID-JAG Assertion ──▶ AAuth Token ──▶ Resource
  │                          │                    │              │
  │ 1. Authenticate          │                    │              │
  │    via IdP               │                    │              │
  │                          │                    │              │
  │         2. ID-JAG assertion                   │              │
  │            (human identity delegation)        │              │
  │                                               │              │
  │                          3. Token exchange    │              │
  │                             (agent auth +     │              │
  │                              mission scope)   │              │
  │                                               │              │
  │                                    4. Authorized request     │
  │                                       (SPIFFE mTLS + Bearer) │
```

## Standards Reference Tables

### Identity Standards

| Standard | Specification | Layer | Status |
|----------|---------------|-------|--------|
| SCIM Agent Resource | [draft-wzdk-scim-agent-resource](https://datatracker.ietf.org/doc/draft-wzdk-scim-agent-resource/) | Lifecycle | Active Draft |
| WIMSE Architecture | [draft-ietf-wimse-architecture](https://datatracker.ietf.org/doc/draft-ietf-wimse-architecture/) | Workload | WG Draft |
| SPIFFE | [spiffe.io](https://spiffe.io/) | Workload | Production |
| AAuth | [draft-hardt-oauth-aauth-protocol](https://datatracker.ietf.org/doc/draft-hardt-oauth-aauth-protocol/) | Agent | Active Draft |
| OAuth 2.1 | [draft-ietf-oauth-v2-1](https://datatracker.ietf.org/doc/draft-ietf-oauth-v2-1/) | Delegation | WG Draft |
| ID-JAG | [draft-ietf-oauth-identity-assertion-authz-grant](https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/) | Delegation | WG Draft |

### Authorization Standards

| Standard | Type | Specification | Purpose |
|----------|------|---------------|---------|
| AuthZEN | API | [OpenID AuthZEN](https://openid.net/specs/openid-authzen-authorization-api-1_0.html) | PEP-PDP communication |
| Cedar | Language | [cedarpolicy.com](https://www.cedarpolicy.com/) | ABAC policy evaluation |
| OpenFGA | Service | [openfga.dev](https://openfga.dev/) | ReBAC authorization |

### Interoperability Standards

| Standard | Governance | Specification | Purpose |
|----------|------------|---------------|---------|
| A2A | Linux Foundation | [A2A Protocol](https://github.com/a2a-protocol/a2a) | Agent discovery & delegation |
| MCP | Agentic AI Foundation | [MCP Spec](https://spec.modelcontextprotocol.io/) | Tool/resource integration |

### Observability Standards

| Standard | Conventions | Documentation | Purpose |
|----------|-------------|---------------|---------|
| OpenTelemetry | gen_ai.* | [GenAI Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/) | Traces, metrics, logs |
| AgentOps | Agent extensions | [agentops.ai](https://www.agentops.ai/) | Session replay, monitoring |

### Runtime Infrastructure

| Component | Standard | Documentation | Purpose |
|-----------|----------|---------------|---------|
| SPIRE | SPIFFE | [spiffe.io/spire](https://spiffe.io/docs/latest/spire-about/) | Workload identity runtime |
| Kubernetes | Workload Identity | [K8s ServiceAccount](https://kubernetes.io/docs/concepts/security/service-accounts/) | Container orchestration |
| Istio | AuthorizationPolicy | [Istio Authorization](https://istio.io/latest/docs/concepts/security/) | Service mesh authorization |

## References

### IETF Specifications

- [draft-wzdk-scim-agent-resource](https://datatracker.ietf.org/doc/draft-wzdk-scim-agent-resource/) - SCIM Agent Resource Extension
- [draft-ietf-wimse-architecture](https://datatracker.ietf.org/doc/draft-ietf-wimse-architecture/) - WIMSE Architecture
- [draft-hardt-oauth-aauth-protocol](https://datatracker.ietf.org/doc/draft-hardt-oauth-aauth-protocol/) - Agent Authorization Protocol
- [draft-ietf-oauth-identity-assertion-authz-grant](https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/) - ID-JAG
- [draft-ietf-oauth-v2-1](https://datatracker.ietf.org/doc/draft-ietf-oauth-v2-1/) - OAuth 2.1
- [RFC 6749](https://www.rfc-editor.org/rfc/rfc6749) - OAuth 2.0 Authorization Framework
- [RFC 7644](https://www.rfc-editor.org/rfc/rfc7644) - SCIM Protocol

### Industry Standards

- [SPIFFE](https://spiffe.io/) - Secure Production Identity Framework for Everyone
- [OpenFGA](https://openfga.dev/) - Fine-Grained Authorization
- [Cedar](https://www.cedarpolicy.com/) - Policy Language
- [AuthZEN](https://openid.net/specs/openid-authzen-authorization-api-1_0.html) - Authorization API

### Agent Protocols

- [A2A Protocol](https://github.com/a2a-protocol/a2a) - Agent-to-Agent Protocol
- [Model Context Protocol](https://spec.modelcontextprotocol.io/) - MCP Specification

### Observability

- [OpenTelemetry GenAI](https://opentelemetry.io/docs/specs/semconv/gen-ai/) - GenAI Semantic Conventions
- [AgentOps](https://www.agentops.ai/) - Agent Observability Platform

### Agent Frameworks

- [Claude Code](https://docs.anthropic.com/claude/docs/claude-code) - Anthropic's coding assistant
- [OpenAI Codex](https://platform.openai.com/docs/guides/code) - OpenAI code generation
- [CrewAI](https://www.crewai.com/) - Multi-agent orchestration
- [Google Agent Developer Kit](https://cloud.google.com/vertex-ai/docs/generative-ai/agent-builder/overview) - Google ADK
- [Microsoft Agent Framework](https://learn.microsoft.com/en-us/azure/ai-services/agents/) - Microsoft agents
- [LangChain](https://www.langchain.com/) - LLM application framework
