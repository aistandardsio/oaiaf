# OpenTelemetry Integration

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#opentelemetry-genai-semantic-conventions) for current documentation.

## Overview

OpenTelemetry defines standard attributes for AI/ML workloads.

## GenAI Semantic Conventions

### Trace Attributes (`gen_ai.*`)

| Attribute | Type | Description |
|-----------|------|-------------|
| `gen_ai.system` | string | AI system (e.g., "openai", "anthropic") |
| `gen_ai.request.model` | string | Model identifier |
| `gen_ai.request.max_tokens` | int | Token limit |
| `gen_ai.response.finish_reasons` | string[] | Completion reasons |
| `gen_ai.usage.input_tokens` | int | Prompt tokens |
| `gen_ai.usage.output_tokens` | int | Completion tokens |

### Agent-Specific Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `agent.id` | string | OAIAF agent identifier |
| `agent.mission` | string | Current mission scope |
| `agent.delegator` | string | Human who delegated |
| `agent.workload_id` | string | SPIFFE ID |

## Reference

- [OpenTelemetry GenAI Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/)
