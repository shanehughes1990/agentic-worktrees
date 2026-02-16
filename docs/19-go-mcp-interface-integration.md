# 19 - go-mcp Interface Integration

## Purpose

Define the `go-mcp` interface layer that allows compatible agents to call this system directly through MCP tools.

## Scope

- provide a first-class MCP server adapter in the application
- expose core orchestration and diagnostics capabilities as MCP tools
- preserve safety, auditability, and state consistency under MCP invocation

## Design Goals

- parity with CLI capabilities for shared operations
- stable, versioned MCP tool naming
- strict policy enforcement identical to CLI controls
- deterministic request/response contracts

## Required MCP Tool Categories

### Board and Planning Tools

- board ingestion from scope file
- board status/listing
- dependency and blocked-task insight

### Execution Tools

- start autonomous run
- pause/resume queues
- resume task/run from checkpoint
- retry/cancel task/run

### Diagnostics Tools

- queue health and worker status
- agent health and stream insight
- worktree status and per-worktree log tail
- audit sink status and audit tail

### Safety and Control Tools

- preflight status and recheck
- schema compatibility status
- watchdog status and alerts

## MCP Parity Rules

- if capability exists in CLI runtime surface, equivalent MCP tool should exist unless explicitly restricted
- MCP-restricted operations must be documented with policy reason
- CLI and MCP must produce equivalent state changes and audit entries for the same operation

## Authentication and Authorization

- authenticate MCP client identity
- authorize tool execution by policy scope
- deny and audit unauthorized MCP calls

## Audit Requirements for MCP Calls

Each MCP call must emit audit records containing:

- request ID
- caller identity
- tool name and input summary
- policy decision (allow/deny)
- resulting action/status
- linked `run_id`/`task_id`/`job_id` if applicable

## Error and Resilience Semantics

- tool failures must return typed errors (`transient`, `terminal`, `policy_denied`)
- MCP transport failures must not mutate orchestration state without durable intent record
- retries from MCP callers must be idempotency-safe for mutating operations

## Versioning

- MCP tool schema and names must be versioned
- breaking changes require version bump and compatibility window policy
- prefer additive evolution for existing tool schemas

## Security Boundaries

- enforce same command/path restrictions as runtime policy engine
- prevent MCP clients from bypassing admission gates, watchdog controls, or audit requirements

## Non-MVP Expansion

Future enhancements may include:

- richer subscription streams via MCP
- typed dashboards over MCP endpoints
- multi-tenant MCP access controls
