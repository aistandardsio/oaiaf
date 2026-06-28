# Hosted Platforms

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#hosted-platforms) for current documentation.

## Examples

- OpenClaw
- OmniAgent
- Fixie

## Characteristics

- Multi-tenant cloud platforms
- Manage many agents for many organizations
- Require strong isolation and audit
- API-driven agent creation

## Identity Requirements

- SCIM for agent provisioning and lifecycle
- SPIFFE for workload identity within platform
- OAuth for platform-level authentication
- ID-JAG for delegated agent actions
