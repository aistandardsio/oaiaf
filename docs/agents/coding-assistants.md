# Coding Assistants

!!! note "Work in Progress"
    This page is a placeholder. See [Architecture](../architecture.md#coding-assistants) for current documentation.

## Examples

- Claude Code
- OpenAI Codex CLI
- GitHub Copilot CLI

## Characteristics

- Run locally on developer machines or in containers
- Interactive sessions with human developers
- Need access to local files, git, and APIs
- Session-scoped authorization

## Identity Requirements

- Workload identity from local SPIRE agent or container runtime
- Human delegation via interactive OAuth flow
- MCP for tool integration
