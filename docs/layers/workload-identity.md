# Layer 2: Workload Identity (SPIFFE)

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#layer-2-workload-identity-wimsespiffe) for current documentation.

## Overview

This layer establishes the identity of the computing workload (container, VM, or process) that hosts the agent.

## Specifications

- [WIMSE Architecture](https://datatracker.ietf.org/doc/draft-ietf-wimse-architecture/)
- [SPIFFE](https://spiffe.io/)

## Implementation

See [agent-protocols/aims](https://github.com/aistandardsio/agent-protocols/tree/main/aims) for the Go implementation.
