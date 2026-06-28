# Multi-Agent Orchestration

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#multi-agent-orchestration) for current documentation.

## Examples

- CrewAI
- Google Agent Developer Kit (ADK)
- Microsoft AutoGen

## Characteristics

- Multiple agents collaborating on tasks
- Agents delegate to other agents
- Complex authorization chains
- Need discovery and capability negotiation

## Identity Requirements

- A2A for agent-to-agent communication
- SCIM for agent registration and discovery
- Per-agent SPIFFE identity
- Delegation chains tracked via ID-JAG
