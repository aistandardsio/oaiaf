# Service Mesh (Istio)

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#service-mesh-istio) for current documentation.

## Overview

Istio can enforce authorization policies based on SPIFFE identity.

## Authorization Policy Example

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

## Reference

- [Istio Authorization](https://istio.io/latest/docs/concepts/security/)
