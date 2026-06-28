# OAIAF Roadmap

This roadmap outlines planned work across the AI Standards ecosystem repositories.

## Repository Structure

| Repository | Purpose |
|------------|---------|
| [agent-protocols](https://github.com/aistandardsio/agent-protocols) | Protocol implementations (Go libraries) |
| [agentauth](https://github.com/plexusone/agentauth) | Protocol orchestration (combining protocols) |
| [oaiaf](https://github.com/aistandardsio/oaiaf) | Reference architecture documentation |

## Phase 0: Protocol Role Mapping & Deployment Architecture

**Target:** Document protocol roles, trust relationships, and logical deployment components to provide practical implementation guidance.

**Dependency:** Requires PIDL v0.7.0 enhancements (protocol roles, deployment components, trust relationships).

### Protocol Role Analysis

For each OAIAF protocol, document:

| Protocol | Roles to Document |
|----------|-------------------|
| **OAuth 2.x** | Resource Owner, Client, Authorization Server, Resource Server |
| **SCIM** | SCIM Client (provisioner), SCIM Service Provider |
| **SPIFFE/WIMSE** | Workload, SPIRE Agent, SPIRE Server, Trust Domain |
| **AAuth** | Agent, Person Server, Access Server, Agent Provider |
| **ID-JAG** | Assertion Issuer, Token Endpoint, Relying Party |
| **AuthZEN** | PEP (Policy Enforcement Point), PDP (Policy Decision Point), PAP (Policy Admin Point), PIP (Policy Info Point) |
| **MCP** | Host, Client, Server |
| **A2A** | Agent (initiator), Agent (responder), Registry |

| Task | Status |
|------|--------|
| Define protocol roles for all 8 protocols | Planned |
| Document client vs server role for each | Planned |
| Create PIDL files with `protocol_roles` annotations | Planned |

### Logical Deployment Components

Group protocol roles into real-world deployment components:

| Component | Example Products | Implements Roles |
|-----------|------------------|------------------|
| **Identity Provider (IdP)** | Okta, Entra ID, Auth0 | OAuth AS, OIDC Provider, SCIM SP |
| **Identity Governance (IGA)** | SailPoint, Saviynt | SCIM Client, Audit, Lifecycle |
| **Agent Provider** | Custom, agentauth | AAuth Agent Provider, JWKS |
| **Person Server** | Custom, agentauth | AAuth Person Server, Consent UI |
| **Access Server** | Custom, agentauth | AAuth Access Server, Token Endpoint |
| **Policy Decision Point** | Cedar, OpenFGA, OPA | AuthZEN PDP, PAP |
| **Access Gateway** | Kong, Envoy, Istio | AuthZEN PEP, OAuth RS, mTLS termination |
| **MCP Client** | Claude Code, Cursor | MCP Client, AAuth Agent, ID-JAG Client |
| **MCP Server** | Tool providers | MCP Server, OAuth RS |
| **Resource API** | Backend services | OAuth RS, AuthZEN PEP, SPIFFE Workload |
| **SPIRE Infrastructure** | SPIRE | SPIRE Server, SPIRE Agent |

| Task | Status |
|------|--------|
| Define logical component taxonomy | Planned |
| Map each component to protocol roles | Planned |
| Create PIDL files with `deployment_components` | Planned |
| Generate component diagrams via PIDL | Planned |

### Trust Relationships

Document trust relationships between components:

| Relationship Type | Example |
|-------------------|---------|
| **authenticates** | IdP authenticates User |
| **validates** | PDP validates Agent claims |
| **delegates** | User delegates to Agent via ID-JAG |
| **authorizes** | PDP authorizes Agent action |
| **issues** | Agent Provider issues aa-agent+jwt |
| **trusts** | Resource API trusts IdP (via JWKS) |

| Task | Status |
|------|--------|
| Define trust relationship taxonomy | Planned |
| Document credential flows between components | Planned |
| Create PIDL files with `trust_relations` | Planned |
| Generate trust relationship diagrams via PIDL | Planned |

### Deliverables

| Deliverable | Format | Description |
|-------------|--------|-------------|
| `docs/roles/protocol-roles.md` | Markdown | Protocol role reference |
| `docs/roles/deployment-components.md` | Markdown | Component architecture guide |
| `docs/roles/trust-relationships.md` | Markdown | Trust model documentation |
| `pidl/oaiaf-roles.json` | PIDL | Machine-readable role definitions |
| `pidl/oaiaf-components.json` | PIDL | Component topology |
| `pidl/oaiaf-trust.json` | PIDL | Trust relationship model |
| Generated diagrams | SVG | Component and trust diagrams |

---

## Phase 1: Documentation Infrastructure

**Target:** Complete MkDocs setup and GitHub Pages deployment for oaiaf.

| Task | Repository | Status |
|------|------------|--------|
| Create mkdocs.yml with Mermaid support | oaiaf | Planned |
| Add GitHub Pages deployment workflow | oaiaf | Planned |
| Create navigation structure | oaiaf | Planned |
| Add search and theme configuration | oaiaf | Planned |

## Phase 2: Protocol Implementations

**Target:** Complete protocol coverage in agent-protocols.

### AuthZEN Client (Layer 5: Authorization)

| Task | Repository | Status |
|------|------------|--------|
| Create `authzen/` package structure | agent-protocols | Planned |
| Implement AuthZEN evaluation API client | agent-protocols | Planned |
| Add Cedar policy evaluation support | agent-protocols | Planned |
| Add OpenFGA client integration | agent-protocols | Planned |
| Create test fixtures and mocks | agent-protocols | Planned |

**Proposed structure:**

```
agent-protocols/authzen/
├── authzen.go           # Core AuthZEN types
├── client.go            # PDP client implementation
├── client_test.go       # Client tests
├── cedar/               # Cedar policy adapter
│   ├── cedar.go
│   └── cedar_test.go
├── openfga/             # OpenFGA adapter
│   ├── openfga.go
│   └── openfga_test.go
└── examples/
    └── basic/
        └── main.go
```

### A2A Protocol (Agent-to-Agent)

| Task | Repository | Status |
|------|------------|--------|
| Create `a2a/` package structure | agent-protocols | Planned |
| Implement agent card parsing | agent-protocols | Planned |
| Implement discovery client | agent-protocols | Planned |
| Implement delegation token exchange | agent-protocols | Planned |
| Add task invocation client | agent-protocols | Planned |

**Proposed structure:**

```
agent-protocols/a2a/
├── a2a.go               # Core A2A types (AgentCard, etc.)
├── discovery.go         # Agent discovery client
├── delegation.go        # Delegation token handling
├── invoke.go            # Task invocation
├── server.go            # A2A server helpers
└── examples/
    ├── discovery/
    └── delegation/
```

### MCP Integration (Model Context Protocol)

| Task | Repository | Status |
|------|------------|--------|
| Create `mcp/` package structure | agent-protocols | Planned |
| Implement MCP client with auth middleware | agent-protocols | Planned |
| Add tool invocation with token injection | agent-protocols | Planned |
| Add resource access with authorization | agent-protocols | Planned |

**Proposed structure:**

```
agent-protocols/mcp/
├── mcp.go               # Core MCP types
├── client.go            # MCP client
├── auth.go              # Authorization middleware
├── tools.go             # Tool invocation helpers
└── examples/
    └── authorized_tools/
```

## Phase 3: Orchestration Integration

**Target:** Integrate new protocol implementations into agentauth.

| Task | Repository | Status |
|------|------------|--------|
| Add AuthZEN provider to orchestration | agentauth | Planned |
| Add A2A client integration | agentauth | Planned |
| Add MCP authorization middleware | agentauth | Planned |
| Create hybrid provider with all protocols | agentauth | Planned |

## Phase 4: Observability

**Target:** Add OpenTelemetry integration across all repositories.

| Task | Repository | Status |
|------|------------|--------|
| Add OpenTelemetry tracing to agent-protocols | agent-protocols | Planned |
| Add gen_ai.* semantic conventions | agent-protocols | Planned |
| Add AgentOps integration | agentauth | Planned |
| Document observability patterns | oaiaf | Planned |

## Phase 5: Quality & Testing

**Target:** Improve code quality and test coverage.

| Task | Repository | Status |
|------|------------|--------|
| Fix golangci-lint warnings | oaiaf | Planned |
| Add integration tests for token flows | agent-protocols | Planned |
| Add go-spiffe real Workload API support | agent-protocols | Planned |
| Add end-to-end tests | agentauth | Planned |
| Add benchmarks for critical paths | agent-protocols | Planned |

## Phase 6: Examples & Demos

**Target:** Comprehensive examples for all protocols.

| Task | Repository | Status |
|------|------------|--------|
| Multi-protocol example (ID-JAG + AAuth + AIMS) | agentauth | Planned |
| A2A delegation chain example | agent-protocols | Planned |
| MCP with authorization example | agent-protocols | Planned |
| Kubernetes deployment example | oaiaf | Planned |
| Service mesh (Istio) example | oaiaf | Planned |

## Completed Work

### oaiaf

| Task | Status | Date |
|------|--------|------|
| Architecture documentation (1100+ lines) | Done | 2024-06 |
| PIDL flow diagrams (6 flows) | Done | 2024-06 |
| Protocol flows documentation | Done | 2024-06 |

### agent-protocols

| Task | Status |
|------|--------|
| AAuth protocol implementation | Done |
| AIMS/SPIFFE implementation | Done |
| ID-JAG implementation | Done |
| SCIM Agent Resource extension | Done |
| Protocol bridge | Done |

### agentauth

| Task | Status |
|------|--------|
| AAuth provider | Done |
| ID-JAG provider | Done |
| Hybrid provider | Done |
| Policy evaluation | Done |
| Token verification | Done |

## Contributing

To contribute to any of these items:

1. Check the repository's issue tracker for related issues
2. Open an issue if one doesn't exist
3. Reference this roadmap in your PR description
4. Follow the repository's contribution guidelines

## References

- [OAIAF Architecture](../architecture.md)
- [Protocol Flows](../flows.md)
- [agent-protocols README](https://github.com/aistandardsio/agent-protocols)
- [agentauth README](https://github.com/plexusone/agentauth)
