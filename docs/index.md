# OAIAF - OpenAI Agent Framework

OAIAF provides a reference architecture for enterprise AI agent identity and authorization. It documents how emerging standards fit together to address the fundamental questions enterprises face when deploying autonomous AI agents.

## The Five Questions

| Question | Layer | Standards |
|----------|-------|-----------|
| What agents exist? | Lifecycle | SCIM Agent Resource |
| Which workload hosts it? | Workload Identity | WIMSE, SPIFFE |
| Which agent is this? | Agent Auth | AAuth |
| Who delegated authority? | Human Delegation | OAuth 2.x, ID-JAG |
| What can it do? | Authorization | AuthZEN, Cedar, OpenFGA |

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

- **[A2A](https://github.com/a2a-protocol/a2a)** - Agent-to-Agent discovery and delegation
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
