# 07 - Copilot ADK Integration

## Objective

Integrate GitHub Copilot ADK as the first agent runtime while keeping runtime adapters swappable.

## Integration Model

- Define `AgentAdapter` interface in core application layer
- Implement `CopilotADKAdapter` as first concrete runtime
- Invoke adapter from Asynq lifecycle handlers (primarily `task.execute_agent`)
- Use adapter contract for run lifecycle:
  - `StartRun`
  - `StreamEvents`
  - `StopRun`
  - `GetRunResult`

## Model Selection Policy

- global default model: `gpt-5.3-codex`
- per-task override source: PRD metadata
- override validation occurs in preflight/admission path and again at runtime boundary
- unsupported override values must fallback to `gpt-5.3-codex`
- runtime event stream must include selected model and fallback reason (if applied)

## Required Runtime Capabilities

- Start agent with task context and repository/worktree path
- Provide tool access and policy-constrained environment
- Stream token/events/logs in near realtime
- Return structured result with artifacts
- Return error classes suitable for queue retry policy (`transient`, `terminal`)

## Context Injection

Agent should receive:

- task definition and acceptance criteria
- dependency and branch context
- changed-file constraints and policy hints
- retry count and run history summary
- selected runtime model (`gpt-5.3-codex` default or validated override)

## Safety Guardrails

- enforce allowed command/tool policy
- block destructive repository actions outside policy
- capture full transcript/event stream for audit

## Failure Handling

- heartbeat timeout triggers graceful stop then hard kill
- partial output must still be persisted
- adapter errors are typed: transient vs terminal
- transient errors map to Asynq retry path
- terminal errors map to dead-letter/archive with diagnostic payload

## Future Runtime Portability

To support non-Copilot runtimes later, keep:

- runtime-neutral event schema
- runtime-neutral run result schema
- strict separation of adapter-specific config
