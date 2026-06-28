# Kubernetes Workload Identity

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#kubernetes-workload-identity) for current documentation.

## Overview

Kubernetes 1.35+ includes native workload identity with SPIFFE support.

## Example Deployment

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: code-review-agent
  annotations:
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

## Reference

- [Kubernetes ServiceAccount](https://kubernetes.io/docs/concepts/security/service-accounts/)
