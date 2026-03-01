# Slice 01 — SCM Core (GitHub First, Agents Second)

## Objective

Deliver Slice 01 in two SCM-specific parts, in this order:

1. **Part 01A — SCM for GitHub (Full Vertical Slice)**
2. **Part 01B — SCM for Agents**

This ordering ensures complete SCM management capability exists first (admission, execution, auth, provider integration), then agent execution composes it through strict DDD ports.

## Part 01A — SCM for GitHub (Full Vertical Slice)

### Part 01A Scope

Implement a complete GitHub-first SCM vertical slice across API, queue, worker, and infrastructure:

- Source-state operations
- Worktree lifecycle operations (`create`, `sync/update`, `cleanup`)
- Branch lifecycle operations
- Pull request lifecycle operations
- Review operations
- Merge-intent/merge-readiness operations

This includes API control-plane admission, taskengine dispatch policy, worker execution, worker-only SCM auth execution, and GitHub adapter behavior under typed failure semantics.

### Part 01A Task Checklist

- [x] Define `internal/domain/scm` contracts and invariants for source/worktree/branch/PR/review/merge-intent.
- [x] Implement `internal/application/scm` use-cases for full SCM management flows.
- [x] Add SCM API admission contract (GraphQL/control-plane) for required SCM management operations.
- [x] Add SCM job kind/policy wiring in taskengine for queue, idempotency, timeout, and retry defaults.
- [x] Implement worker SCM handler(s) that execute SCM flows through application services.
- [x] Implement worker-only SCM authentication contract and runtime execution path.
- [x] Add first concrete `internal/infrastructure/scm` GitHub adapter for full Part 01A scope.
- [x] Add typed failure classification (`transient` vs `terminal`) and retry-safe mapping.
- [x] Add checkpoint/resume boundaries for long-running SCM worker operations.
- [x] Add integration tests for API admission -> queue -> worker -> SCM adapter path.

### Part 01A Deliverables

- GitHub-first SCM contracts covering `source`, `worktree`, `branch`, `pull request`, `review`, and merge-intent checks.
- SCM application orchestration services using ports only.
- SCM API admission + worker execution path wired end-to-end.
- Worker-owned SCM authentication implementation for provider access.
- Initial GitHub infrastructure adapter implementation.
- Retry-safe typed failure semantics and correlation/idempotency propagation.
- Integration coverage proving full vertical slice behavior.

### Part 01A Acceptance Criteria

- SCM use-cases call only SCM ports; no provider calls from application layer.
- API layer validates/maps/admits; it does not execute SCM auth/provider logic.
- Worker is the only runtime that executes SCM auth logic and provider operations.
- GitHub baseline source/worktree/branch/PR/review/merge-intent primitives are functional through contracts.
- End-to-end SCM flow executes through API -> taskengine -> worker -> SCM adapter with correlation IDs and idempotency.
- DDD boundaries remain strict (`interface -> application -> domain`; infrastructure implements ports).

## Part 01B — SCM for Agents

### Part 01B Scope

Implement agent-facing SCM orchestration that consumes Part 01A SCM contracts for execution workflows.

### Part 01B Task Checklist

- [x] Define `internal/domain/agent` contracts that reference SCM capabilities through ports.
- [x] Implement `internal/application/agent` orchestration for SCM-driven execution/session flows.
- [x] Add correlation and idempotency contract alignment between agent workflows and SCM operations.
- [x] Add checkpoint/resume boundaries for long-running agent+SCM orchestration.
- [x] Add worker handler integration points for agent tasks that invoke SCM application services.
- [x] Add integration tests for agent orchestration through SCM ports (without bypassing application layer).

### Part 01B Deliverables

- Agent contract for SCM-aware execution and session introspection.
- Agent application orchestration that composes SCM services via ports.
- Shared execution semantics (correlation IDs, idempotency, typed failures) across agent+SCM.
- Test coverage for agent-to-SCM orchestration path.

### Part 01B Acceptance Criteria

- Agent layer never calls SCM provider adapters directly.
- SCM interactions in agent workflows occur only through application/domain contracts.
- Retry, checkpoint, and correlation behavior is deterministic and test-covered.
- No Copilot/SCM SDK execution occurs outside worker handler paths.

## In Scope (Slice 01)

- Part 01A full GitHub-first SCM vertical slice.
- Part 01B agent orchestration that consumes Part 01A SCM capabilities.
- API admission/configuration and worker execution paths needed to support both parts.

## Out of Scope (Slice 01)

- Multi-provider parity beyond GitHub baseline.
- UI/client product flows beyond required control-plane admission.
- Non-worker authentication execution paths.

## Dependencies

- Slice 00 complete.
- Existing V1 observability and bootstrap scaffolding available.

## Exit Check

Slice 01 is complete when:

1. GitHub-first SCM primitives (`source`, `worktree`, `branch`, `PR`, `review`, merge-intent) are contract-defined, application-orchestrated, API-admitted, worker-executed, and infrastructure-backed.
2. Agent execution flows consume SCM through ports with typed failures, correlation/idempotency, and checkpoint/resume semantics under strict DDD boundaries.

## Implementation Extension — `.docs/agent-orchestrator` (Reference Mapping)

This section extends Slice 01 with a concrete mapping of how the `.docs/agent-orchestrator` project implements agent+SCM orchestration end-to-end. It is a reference for behavior and contracts, not a direct source migration plan.

### A) Where Agent Orchestrator Reads From

#### 1. Config + project definition reads

- `packages/core/src/config.ts`
  - Reads `agent-orchestrator.yaml` via `readFileSync` + YAML parse + Zod validation.
  - Expands project paths, applies defaults, infers tracker/SCM plugins, validates uniqueness.
  - Exposes fields that directly affect prompt and execution:
    - `projects[*].path`, `repo`, `defaultBranch`, `sessionPrefix`
    - `agentRules`, `agentRulesFile`, `orchestratorRules`
    - `tracker`, `scm`, `reactions`, `agentConfig`

#### 2. Session metadata reads/writes

- `packages/core/src/metadata.ts`
  - Reads/writes flat key-value metadata files in:
    - `~/.agent-orchestrator/{hash}-{projectId}/sessions/{sessionName}`
  - Keys include `status`, `branch`, `pr`, `issue`, `worktree`, `runtimeHandle`, `agent`.
- `packages/core/src/session-manager.ts`
  - Reads metadata on list/get/send/kill/restore.
  - Uses metadata as canonical session control state.

#### 3. Tracker issue/context reads

- `packages/plugins/tracker-github/src/index.ts`
  - Reads issue via `gh issue view` (`title`, `body`, labels, assignee, state).
  - Builds issue prompt context via `generatePrompt(identifier, project)`.

#### 4. Runtime/agent state reads

- `packages/plugins/runtime-tmux/src/index.ts`
  - Reads tmux liveness and pane output (`tmux has-session`, `capture-pane`).
- `packages/plugins/agent-claude-code/src/index.ts`
  - Reads Claude JSONL session files from `~/.claude/projects/{encoded-path}/...jsonl`.
  - Extracts activity state, summary, and cost from JSONL tail.
- `packages/plugins/agent-codex/src/index.ts` / `agent-opencode`
  - Process liveness checks via `ps`/TTY; limited session-introspection support.

#### 5. SCM + PR/CI/review reads

- `packages/plugins/scm-github/src/index.ts`
  - Reads PR state, checks, reviews, unresolved comments, mergeability via `gh` API/GraphQL.
  - Detects PR from branch when metadata has no PR URL (`detectPR`).

---

### B) Prompt Construction + Injection (How Prompts Are Implemented)

#### 1. Worker-agent prompt composition

- `packages/core/src/prompt-builder.ts`
  - `buildPrompt()` composes **3 layers**:
    1. `BASE_AGENT_PROMPT` (session lifecycle, git/PR expectations)
    2. Config-derived context (project/repo/default branch/tracker/issue/reactions)
    3. User rules (`agentRules` + `agentRulesFile` content)
  - Optional explicit user prompt is appended as highest-priority instructions.
  - `agentRulesFile` is read from disk using `readFileSync(resolve(project.path, file))`.

#### 2. Orchestrator-agent system prompt

- `packages/core/src/orchestrator-prompt.ts`
  - `generateOrchestratorPrompt()` builds the orchestrator system prompt with command surface + workflows.
- `packages/cli/src/commands/start.ts`
  - Calls `generateOrchestratorPrompt(...)` and passes it to `spawnOrchestrator`.
- `packages/core/src/session-manager.ts`
  - For orchestrator sessions, writes prompt to file (`orchestrator-prompt.md`) and passes `systemPromptFile` to agent launch config.

#### 3. Agent plugin launch mapping

- `agent-claude-code` plugin:
  - Uses `--append-system-prompt` + `-p` for task prompt.
  - Supports `systemPromptFile` to avoid shell/tmux truncation.
- `agent-codex` plugin:
  - Uses Codex config overrides (`model_instructions_file` or `developer_instructions`) + prompt arg.
- `agent-opencode` plugin:
  - Uses `opencode run <prompt>`.

---

### C) End-to-End Execution Flow (Through and Through)

#### 1. Spawn path

- `packages/cli/src/commands/spawn.ts` → `SessionManager.spawn(...)`
- `packages/core/src/session-manager.ts::spawn`
  1. Resolve plugins (runtime/agent/workspace/tracker/scm)
  2. Validate issue existence via tracker (when configured)
  3. Reserve session ID + derive branch
  4. Create workspace/worktree via workspace plugin
  5. Build issue context via tracker `generatePrompt`
  6. Compose final prompt via `buildPrompt`
  7. Build agent launch command/env from agent plugin
  8. Create runtime session (tmux/process)
  9. Persist metadata (`status=spawning`, branch/worktree/runtimeHandle/etc.)
  10. Run `postLaunchSetup` hooks (agent-specific)

#### 2. Orchestrator start path

- `packages/cli/src/commands/start.ts`
  - Starts dashboard and orchestrator session.
  - Generates orchestrator system prompt and calls `spawnOrchestrator`.
- `packages/core/src/session-manager.ts::spawnOrchestrator`
  - Writes system prompt file, launches runtime, writes metadata, applies post-launch setup.

#### 3. Polling + transitions + automated actions

- `packages/core/src/lifecycle-manager.ts`
  - Polls sessions, computes status transitions, writes updated status to metadata.
  - Auto-detects PR if missing, then checks CI/review/merge readiness.
  - Executes configured reactions (`send-to-agent`, `notify`, `auto-merge` behavior path).
  - Uses retries/escalation before human notifications where configured.

---

### D) How Agent Code Is Handled (Operational Mechanisms)

#### 1. Metadata mutation from agent-side commands

- `agent-codex`
  - Installs PATH wrappers for `gh` and `git` to auto-update metadata on:
    - branch creation
    - PR create
    - PR merge
  - Appends AO section into `AGENTS.md` for reinforcement/fallback.
- `agent-claude-code`
  - Installs `.claude/metadata-updater.sh` and `settings.json` PostToolUse hooks.
  - Hook parses command invocations and updates metadata (`branch`, `pr`, `status`).

#### 2. Message delivery to agents

- `SessionManager.send(...)` resolves runtime handle and delegates to runtime plugin.
- `runtime-tmux.sendMessage(...)` injects text reliably (buffer paste for long/multiline, then `Enter`).

#### 3. Recovery behavior

- `SessionManager.restore(...)` rebuilds session from active/archive metadata and runtime/workspace restoration path.
- Lifecycle state recomputation prevents stale metadata from becoming authoritative forever.

---

### E) Slice 01 (Part 01B) Alignment Requirements Derived from AO Reference

To align V1 Part 01B with this proven operational model:

1. Preserve strict separation of concerns:
   - admission + mapping in interface layer,
   - orchestration + checkpoint logic in application layer,
   - SCM/provider execution in worker/infrastructure.
2. Keep prompt construction deterministic and layered:
   - base instructions,
   - config/session/task context,
   - file-based and inline rule overlays.
3. Ensure every agent+SCM step is resumable with checkpoint token semantics.
4. Persist canonical session metadata that can be read/written by both orchestrator and worker flows.
5. Keep PR/CI/review reaction handling explicit, typed, and retry/escalation-aware.
6. Keep agent integration adapter-based (Codex/Claude/etc.) while preserving common contracts.

This extension is now the concrete behavioral reference for Slice 01 agent+SCM implementation fidelity.

## Reliability + Stability Hardening (Agent/SCM Core)

This section captures non-supervisor reliability improvements identified by reviewing current V1 implementation against `.docs/agent-orchestrator` operational patterns.

### Current Strengths (V1)

- Strong contract validation and typed failure semantics across domain/application/interface boundaries.
- Queue-level idempotency, uniqueness, retry, and timeout defaults are in place.
- Per-repository concurrency protection exists for SCM worktree orchestration.

### Gaps to Close for Production Stability

1. **Checkpoint persistence is not fully wired into execution flow**
   - V1 has a checkpoint store contract and Redis implementation, but runtime handler/service orchestration still relies mainly on payload-carried checkpoint data.
   - Reliability impact: resume behavior is less durable across worker/runtime boundaries than it should be.

2. **Agent execution step-machine depth is minimal**
   - Current agent execution path covers initial source-state checkpoint behavior but not full stepwise orchestration persistence.
   - Reliability impact: partial-progress recovery is shallow when interruption happens mid-flow.

3. **No durable execution journal comparable to AO metadata lifecycle**
   - AO persists session-level state transitions and operational facts for robust restarts and human debugging.
   - Reliability impact: V1 introspection and replay decisions after crashes are harder and less deterministic.

4. **Dead-letter operational path not codified**
   - Retries are configured, but explicit failure handling/runbook interfaces for dead-letter inspection/requeue are not yet formalized in the same way.
   - Reliability impact: incident handling and recovery speed degrade under repeated failures.

### Priority Hardening Plan (Non-Supervisor Scope)

#### Priority 1 — Persistent checkpoint orchestration

- Wire `CheckpointStore` into worker/application execution boundaries:
  - Load persisted checkpoint at job start by `idempotency_key`.
  - Save checkpoint after each completed step.
  - Use persisted checkpoint as source-of-truth for retry resume.
- Keep retry semantics deterministic via `step + token` matching.

#### Priority 2 — Expand agent execution into explicit resumable steps

- Move from single-step source check to explicit, validated step machine for agent+SCM path (e.g. source/worktree/branch/PR readiness segments as applicable to slice scope).
- Persist checkpoint at each step boundary to support interruption-safe resume.

#### Priority 3 — Add execution-state journal for diagnostics + replay safety

- Introduce a minimal durable state record keyed by `run_id/task_id/job_id` with current step, last update time, and terminal classification.
- Use this to make restart/replay behavior observable and deterministic.

#### Priority 4 — Codify dead-letter and failure triage path

- Add task-engine-facing interface and infra implementation for dead-letter listing/requeue operations.
- Add a concise operational runbook to keep failure recovery repeatable.

### Slice 01 Exit Reinforcement

For Slice 01 to be considered reliability-complete (agent+SCM core), execution must be:

- deterministic under retries,
- resumable from persisted checkpoints (not only payload memory),
- observable through durable execution state,
- recoverable through explicit dead-letter/triage operations,
- compliant with strict DDD boundaries while keeping worker as the only SCM execution runtime.

### Reliability Hardening Implementation Status (Non-Supervisor)

- [x] Persisted checkpoint orchestration wired into worker handlers and bootstrap-backed Redis checkpoint store.
- [x] Deterministic retry resume now loads persisted checkpoint by `idempotency_key` before operation execution.
- [x] Durable execution journal contract added and wired (`running`/`succeeded`/`failed`/`skipped`) for agent and SCM handlers.
- [x] Redis execution journal infrastructure implementation added.
- [x] Dead-letter triage contract added (`list` + `requeue`) with asynq adapter implementation.
- [x] Reliability-focused unit/integration coverage added/updated for handler checkpoint+journal behavior.

This completes the current non-supervisor reliability scope for Slice 01 agent+SCM core.
