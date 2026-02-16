# 02 - Functional and Non-Functional Requirements

## Functional Requirements

### FR-1 Interface Layer

- Provide CLI interface implemented with `urfave/cli/v3`
- Provide `go-mcp` interface adapter for agent-callable operations
- Keep orchestration logic independent of interface transport
- Define stable application service layer for future interfaces (TUI/API/Web)

### FR-1E go-mcp Interface Contract

- Expose core runtime operations as MCP tools so compatible agents can call the system directly
- Maintain behavioral parity between CLI and MCP for shared operations
- Support MCP tools for board generation, run start/resume, runtime diagnostics, and control-plane actions
- Enforce same authz/safety/audit rules for MCP calls as CLI calls
- MCP interface failures must not compromise core orchestration state consistency

### FR-1D Autonomous Walk-Away Mode

- System default mode must support end-to-end autonomous execution with little to zero human intervention
- Operator should be able to generate board, start run, and allow system to complete full board cycle automatically
- Human intervention is reserved for exceptional/terminal policy escalations only

### FR-1A Durable Execution Core

- Use Asynq as the primary execution engine for asynchronous durable tasks
- Define typed task kinds for each lifecycle phase (prepare, execute, validate, PR, merge, cleanup)
- Ensure all worker handlers are idempotent and checkpoint-aware
- Persist retry metadata, error class, and last checkpoint per run

### FR-1B External Dependency and Auth Preflight Gate

- Before accepting or enqueueing any task, run a mandatory preflight validation phase
- Validate required tool availability and minimum versions (`git`, `gh`)
- Validate `gh` authentication status and required scopes for PR operations
- Validate runtime model configuration policy and supported model allowlist
- Validate schema compatibility for all persisted artifacts before task admission
- Fail fast in control plane if validation fails; do not accept tasks into execution queues
- Re-run preflight on startup, after credential rotation events, and on operator-demanded health checks

### FR-1C Agent Model Selection Policy

- Default agent model must be `gpt-5.3-codex`
- Support per-task model override via PRD metadata
- Per-task override must be validated before task admission/execution
- If override is unsupported, automatically fallback to default model `gpt-5.3-codex`
- Emit explicit event/annotation when fallback is applied

### FR-2 Task and Kanban Management

- Support parallel task lists
- Support task dependencies (DAG, no cycles)
- Track task states: `backlog`, `ready`, `in_progress`, `blocked`, `review`, `done`, `failed`
- Enforce dependency completion before scheduling
- Support task priority, labels, and ownership
- Represent tasks as dependency-aware tree/DAG structures
- Support PRD attachment per task for execution context
- Support PRD metadata for runtime model override policy
- Materialize runnable DAG nodes into Asynq jobs only when dependencies are satisfied

### FR-2A Persistence Abstraction

- Define backend-agnostic persistence interfaces for board/tasks/PRDs/events
- Core orchestration must depend only on interfaces, never concrete storage implementations
- Support runtime backend selection via configuration
- First concrete backend is file-based JSON PRD + board storage
- Keep task metadata persistence separate from Asynq queue persistence responsibilities
- All persisted schemas must include explicit version fields
- New schema versions must preserve backward-read compatibility with supported prior versions
- Schema evolution must define deterministic migration/up-conversion rules

### FR-2B CLI Seeding

- CLI must seed required template files for first-time initialization
- Seed output must include minimal board file and PRD template files
- Seeding operation must be idempotent and safe to rerun

### FR-2C Realtime Watch

- Support realtime watch of task/board/PRD changes
- Fan out change notifications to all active task threads/workers
- Guarantee ordered delivery per source file and monotonic event timestamps
- Include queue lifecycle events (`enqueued`, `started`, `retry`, `archived`, `succeeded`)

### FR-2D Scope Ingestion and Board Generation

- Provide a dedicated ingestion path that accepts a scope/task/feature input file
- Analyze current project state before planning tasks
- Decompose scope into dependency-safe task DAG/tree with full metadata
- Assign decomposed tasks into parallelizable lanes/lists
- Materialize internal planning model and export JSON seed artifacts in one operation
- Keep ingestion output schema-compatible with future relational table backends
- Ingestion must read supported older schema versions and normalize to current canonical model

### FR-3 Worktree Lifecycle

- Create one worktree per active task/agent execution
- Ensure each iteration fetches and rebases against latest `origin/<target-branch>`
- Preserve isolated working directory and branch mapping
- Clean up worktrees on success and policy-defined failure conditions
- Ensure worktree operations are safe under Asynq retries and at-least-once execution

### FR-4 Agent Execution and Supervision

- Launch and supervise agent sessions through interface-agnostic runtime controls
- Stream stdout/stderr/events in realtime
- Support cancellation, retry, and graceful stop
- Persist execution logs and lifecycle events
- Map Asynq cancellation and timeout semantics to runtime stop semantics

### FR-5 Git + PR Automation

- Automate branch creation and synchronization
- Push commits and create/update pull requests
- Rebase against origin before merge attempts
- Detect and classify merge conflicts
- Apply policy-driven conflict handling (auto/manual/escalate)
- Merge using configured strategy after all gates pass
- Capture target origin branch at run start and use it as canonical merge target for the full board cycle

### FR-6 Reliability and Recovery

- All operations must be idempotent
- Persist orchestrator state for restart recovery
- Resume unfinished workflow from last durable checkpoint
- Support dead letter queue for repeatedly failing tasks
- Use Asynq retry policies with explicit max retry and backoff rules per task type
- Store and expose dead-letter/archive reasons for operator triage
- Treat preflight failures as admission-control failures (block enqueue), not worker retries
- Treat unsupported schema versions as admission-control failures until migration/remediation is complete
- Preserve half-completed work by default and resume from last durable checkpoint when possible
- Discard/reset in-progress work only under explicitly classified extreme/terminal conditions
- Resume logic must prefer same task branch/worktree context before considering rebuild paths

### FR-7 Observability

- Realtime per-agent status and event stream
- Structured logs and metrics for each state transition
- Traceable run IDs across task, worktree, and PR lifecycle
- Queue-level visibility: depth, latency, retries, in-flight workers, dead-letter volume

### FR-7A Mandatory File Audit Trail

- Application must write full end-to-end audit logs to file by default
- Audit entries must include lifecycle step, command/action, relevant paths, and resulting outputs/status
- Audit logging must be enabled before task acceptance and remain active for all runs
- Audit records must be append-only and correlated by run/task/job IDs
- If audit file sink is unavailable or unwritable, task intake/enqueueing must be blocked
- Each worktree agent thread must write to its own dedicated log file named `<worktree-name>.log`
- Primary application logs and per-worktree thread logs must remain separated by default

### FR-7B Runtime Diagnostics Command Surface

- Provide runtime commands for board status, queue status, worker/agent health, worktree state, git action state, and stream inspection
- Support in-flight visibility while runs/jobs are actively executing
- Return correlation metadata (`request_id`, `run_id`, `task_id`, `job_id`) where applicable

### FR-7C Watchdogs and Healthchecks

- Implement mandatory watchdog loops for heartbeat, queue progress, audit sink health, preflight gate health, worktree orphan detection, and stream lag
- Expose watchdog status through runtime health command and service-level health endpoints
- Trigger fail-closed task intake behavior for safety-critical watchdog failures

## Non-Functional Requirements

### NFR-1 Resilience

- System tolerates process restarts without orphaning work
- Explicit handling for network/API/git transient failures
- Controlled exponential backoff with jitter
- Durable queue restart behavior validated through crash-recovery drills

### NFR-2 Performance

- Efficient parallelism without oversubscribing host resources
- Predictable scheduling under configurable concurrency limits
- Queue priorities and weights tuned to prevent starvation of critical tasks

### NFR-3 Safety

- Safe defaults for merge operations
- Policy checks before destructive operations
- Immutable audit trail for task decisions and conflict actions

### NFR-4 Portability

- Runs locally and in containerized environments
- Minimal host assumptions beyond Git and container runtime
- Requires Redis-compatible runtime for Asynq queue durability

### NFR-5 Extensibility

- Interface layer pluggable without core workflow rewrite
- Agent adapters abstracted (Copilot ADK first, others later)
- Persistence backend swappable without planner/workflow rewrites

## Explicit Constraints

- Language: Go
- First UX: CLI (`urfave/cli/v3`)
- Core execution engine: Asynq
- Queue durability backend: Redis
- tmux integration is deferred until a later, non-MVP interface phase
- Must be container-ready
- Must support fully automated repetitive cycle across latest origin state
