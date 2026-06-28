# Agent Types Overview

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#agent-type-reference) for current documentation.

## Reference Matrix

| Agent Type | Examples | Identity | Auth | Delegation | Interop | Runtime |
|------------|----------|----------|------|------------|---------|---------|
| Coding Assistants | Claude Code, Codex CLI | SPIFFE | AAuth | ID-JAG | MCP | Local/Container |
| Hosted Platforms | OpenClaw, OmniAgent | SCIM+SPIFFE | AAuth+OAuth | ID-JAG | A2A+MCP | Kubernetes |
| Single-Agent Orchestration | LangChain | SPIFFE | AAuth | ID-JAG | MCP | Various |
| Multi-Agent Orchestration | CrewAI, Google ADK | SCIM+SPIFFE | AAuth | ID-JAG | A2A+MCP | Kubernetes |
| Enterprise Frameworks | Microsoft Agent Framework | SCIM+SPIFFE | AAuth+OAuth | ID-JAG | A2A+MCP | Azure/K8s |

## Detailed Pages

- [Coding Assistants](coding-assistants.md)
- [Hosted Platforms](hosted-platforms.md)
- [Multi-Agent Orchestration](multi-agent.md)
- [Enterprise Frameworks](enterprise.md)
