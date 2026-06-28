# SPIRE

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#spire) for current documentation.

## Overview

SPIRE (SPIFFE Runtime Environment) provides the workload identity infrastructure.

## Architecture

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

## Registration Entry Example

```bash
spire-server entry create \
    -spiffeID spiffe://example.com/agent/code-review \
    -parentID spiffe://example.com/node/k8s-node-1 \
    -selector k8s:ns:ai-agents \
    -selector k8s:sa:code-review
```

## Reference

- [SPIRE Documentation](https://spiffe.io/docs/latest/spire-about/)
