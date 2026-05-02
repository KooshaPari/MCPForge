# MCPForge — Workspace Operating Notes

## Project Overview

MCPForge provides an MCP (Model Context Protocol) language server that gives MCP-enabled clients access to semantic tools like get definition, references, rename, and diagnostics.

- **Stack**: Go
- **CLI**: `go run ./cmd/mcp-language-server/...`
- **Integration tests**: `go test ./integrationtests/...`

## Build & Test

```bash
go build ./...        # Build all packages
go test ./...         # Run all tests
task lint             # Run linters (if Taskfile available)
```

## Code Quality

- Run `go vet ./...` before commit
- `golangci-lint run` if configured
- No new `unsafe` without explicit safety comments

## Governance

- Reference: `~/.claude/CLAUDE.md` (global baseline)
- Reference: `AgilePlus/CLAUDE.md` (Phenotype org governance)
- CLAUDE.md required for all Phenotype org repos
- Fork-aware: MCPForge extends upstream MCP protocol; contribute upstream where applicable
