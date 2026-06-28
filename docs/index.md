# Open Agent Internet Architecture Framework (OAIAF)

**An Open Standards Reference Architecture for Enterprise AI Agents**

OAIAF provides a reference architecture for enterprise AI agent identity and authorization. It documents how emerging standards fit together to address the fundamental questions enterprises face when deploying autonomous AI agents.

## The Five-Layer Agent Identity Stack

```
┌────────────────────────────────────────────────────────────────────────────┐
│  Layer 5: AUTHORIZATION                                                    │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐                   │
│  │   AuthZEN     │  │    Cedar      │  │   OpenFGA     │                   │
│  │   (API)       │  │   (ABAC)      │  │   (ReBAC)     │                   │
│  └───────────────┘  └───────────────┘  └───────────────┘                   │
│  "What can this agent do?" → Policy-based access control decisions         │
├────────────────────────────────────────────────────────────────────────────┤
│  Layer 4: HUMAN DELEGATION                                                 │
│  ┌───────────────────────────┐  ┌──────────────────────────┐               │
│  │      OAuth 2.x            │  │        ID-JAG            │               │
│  │   (Authorization)         │  │  (Identity Assertion)    │               │
│  └───────────────────────────┘  └──────────────────────────┘               │
│  "Who delegated authority?" → Chain of authority from human to agent       │
├────────────────────────────────────────────────────────────────────────────┤
│  Layer 3: AGENT AUTHENTICATION                                             │
│  ┌─────────────────────────────────────────────────────────┐               │
│  │                        AAuth                            │               │
│  │            (HTTP Signatures + Mission Scope)            │               │
│  └─────────────────────────────────────────────────────────┘               │
│  "Which autonomous agent is this?" → Cryptographic agent identity          │
├────────────────────────────────────────────────────────────────────────────┤
│  Layer 2: WORKLOAD IDENTITY                                                │
│  ┌───────────────────────────┐  ┌──────────────────────────┐               │
│  │         WIMSE             │  │        SPIFFE            │               │
│  │    (Workload Identity)    │  │    (X.509 SVIDs)         │               │
│  └───────────────────────────┘  └──────────────────────────┘               │
│  "Which workload hosts this agent?" → Infrastructure-level identity        │
├────────────────────────────────────────────────────────────────────────────┤
│  Layer 1: LIFECYCLE MANAGEMENT                                             │
│  ┌─────────────────────────────────────────────────────────┐               │
│  │                  SCIM Agent Resource                    │               │
│  │          (Provisioning, Capabilities, Metadata)         │               │
│  └─────────────────────────────────────────────────────────┘               │
│  "What agents exist?" → Agent registration and capability declaration      │
└────────────────────────────────────────────────────────────────────────────┘

Cross-Cutting Concerns:
┌─────────────────────────┐  ┌─────────────────────────┐  ┌─────────────────┐
│  A2A (Agent-to-Agent)   │  │  MCP (Model Context)    │  │  OpenTelemetry  │
│  Discovery & Delegation │  │  Tool Integration       │  │  Observability  │
└─────────────────────────┘  └─────────────────────────┘  └─────────────────┘
```

| Layer | Standards | Question Answered |
|-------|-----------|-------------------|
| 5. Authorization | AuthZEN, Cedar, OpenFGA | What can this agent do? |
| 4. Human Delegation | OAuth 2.x, ID-JAG | Who delegated authority to this agent? |
| 3. Agent Authentication | AAuth | Which autonomous agent is this? |
| 2. Workload Identity | WIMSE, SPIFFE | Which workload/service hosts this agent? |
| 1. Lifecycle | SCIM Agent Resource | What agents exist and what are their capabilities? |

## About the Name

Each word in **Open Agent Internet Architecture Framework** was chosen deliberately:

| Term | Meaning |
|------|---------|
| **Open** | Emphasizes open standards, vendor neutrality, and interoperability—not necessarily open source |
| **Agent** | Clearly defines the domain as AI agents |
| **Internet** | Reflects that the framework is grounded in Internet standards from IETF, OpenID Foundation, W3C, Linux Foundation, and related communities |
| **Architecture** | Distinguishes it from AI governance, ethics, or policy-only frameworks by making it clear this is a technical reference architecture |
| **Framework** | Positions it alongside mature architecture frameworks like TOGAF and SABSA rather than as a single specification |

!!! note "Definition of Open"
    **Open** refers to the use of open Internet standards and interoperable architectures developed by standards organizations and open industry communities. It does not imply that every implementation must be open source.

## Ecosystem Position

OAIAF sits within a broader ecosystem of standards and tooling:

```
Standards Catalog Framework (SCF)
        │
        ▼
Agent Standards Catalog (ASC)
        │
        ▼
Open Agent Internet Architecture Framework (OAIAF)
        │
        ▼
agent-protocols
        │
        ▼
Generated protocol artifacts
(SCIM, AAuth, A2A, MCP, AuthZEN, etc.)
```

## Quick Links

<div class="grid cards" markdown>

- :material-sitemap:{ .lg .middle } **Architecture**

    ---

    Comprehensive architecture covering the five-layer identity stack

    [:octicons-arrow-right-24: Architecture Guide](architecture.md)

- :material-chart-timeline:{ .lg .middle } **Protocol Flows**

    ---

    Detailed sequence diagrams for ID-JAG, AAuth, AIMS, and more

    [:octicons-arrow-right-24: Protocol Flows](flows.md)

- :material-road:{ .lg .middle } **Roadmap**

    ---

    Planned work across the AI Standards ecosystem

    [:octicons-arrow-right-24: Roadmap](specs/ROADMAP.md)

</div>

## Related Projects

| Repository | Purpose |
|------------|---------|
| [agent-protocols](https://aistandards.io/agent-protocols/) | Go implementations of AAuth, ID-JAG, AIMS, SCIM Agent Resource |
| [agentauth](https://github.com/plexusone/agentauth) | Protocol orchestration and hybrid providers |
| [PIDL](https://github.com/grokify/pidl) | Protocol Interaction Description Language for diagrams |

## Supported Protocols

### Identity & Authentication

- **[ID-JAG](https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/)** - Identity Assertion Authorization Grant for automated agent authorization
- **[AAuth](https://datatracker.ietf.org/doc/draft-hardt-oauth-aauth-protocol/)** - Agent Authorization Protocol for human-in-the-loop consent
- **[AIMS/SPIFFE](https://spiffe.io/)** - Workload identity with X.509 SVIDs
- **[SCIM Agent Resource](https://datatracker.ietf.org/doc/draft-wzdk-scim-agent-resource/)** - Agent lifecycle management

### Authorization

- **[AuthZEN](https://openid.net/specs/openid-authzen-authorization-api-1_0.html)** - PEP-PDP communication API
- **[Cedar](https://www.cedarpolicy.com/)** - ABAC policy language
- **[OpenFGA](https://openfga.dev/)** - ReBAC authorization service

### Interoperability

- **[A2A](https://google.github.io/A2A/)** - Agent-to-Agent discovery and delegation
- **[MCP](https://spec.modelcontextprotocol.io/)** - Model Context Protocol for tool integration

## Getting Started

OAIAF is primarily a documentation project. For code implementations, see:

```bash
# Protocol implementations
go get github.com/aistandardsio/agent-protocols

# Orchestration library
go get github.com/plexusone/agentauth
```

## License

MIT License - see [LICENSE](https://github.com/aistandardsio/oaiaf/blob/main/LICENSE) for details.
