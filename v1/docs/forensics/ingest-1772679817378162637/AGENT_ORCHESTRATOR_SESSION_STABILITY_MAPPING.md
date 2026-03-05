# Agent-Orchestrator Session Stability Mapping

## Purpose
This document summarizes what `.docs/agent-orchestrator` uses for session/task stability, and maps that behavior to the forensic artifacts dumped for ingestion cycle `ingest-1772679817378162637`.

## Forensic Inputs (This Ingestion Bundle)
From this bundle, the key session-state artifacts are:

- `session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/events.jsonl`
- `session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/workspace.yaml`
- `session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/session.db`
- `session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/checkpoints/index.md`
- `session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/plan.md`

## What Agent-Orchestrator Actually Uses for Stability

### 1. Runtime liveness and state reconciliation
Session stability is primarily derived from runtime/process liveness checks and enrichment:

- Mark session `exited` when terminal state is known.
- Mark session `killed` when runtime handle says not alive.
- Reconcile this in `list/get/restore` paths.

References:

- `.docs/agent-orchestrator/packages/core/src/session-manager.ts:257`
- `.docs/agent-orchestrator/packages/core/src/session-manager.ts:276`
- `.docs/agent-orchestrator/packages/core/src/session-manager.ts:682`
- `.docs/agent-orchestrator/packages/core/src/session-manager.ts:720`

### 2. Flat metadata files as persistence contract
They persist canonical session control data in metadata files (not Copilot session-state files):

- `status`, `runtimeHandle`, `branch`, `worktree`, etc.
- Metadata archive on delete for later restoration.
- Restore can read archived metadata and recreate active metadata.

References:

- `.docs/agent-orchestrator/packages/core/src/metadata.ts:112`
- `.docs/agent-orchestrator/packages/core/src/metadata.ts:160`
- `.docs/agent-orchestrator/packages/core/src/metadata.ts:191`
- `.docs/agent-orchestrator/packages/core/src/metadata.ts:211`
- `.docs/agent-orchestrator/packages/core/src/session-manager.ts:944`

### 3. Restore eligibility guardrails
Restore behavior is constrained by explicit terminal/restorable rules:

- Terminal status/activity checks.
- Non-restorable status set (e.g., merged).

References:

- `.docs/agent-orchestrator/packages/core/src/types.ts:90`
- `.docs/agent-orchestrator/packages/core/src/types.ts:101`
- `.docs/agent-orchestrator/packages/core/src/types.ts:116`
- `.docs/agent-orchestrator/packages/core/src/session-manager.ts:985`

### 4. Agent plugin activity source differs by agent

- `agent-codex`: no per-session JSONL introspection currently; activity resolution returns `null` after process-running check.
- `agent-claude-code`: uses JSONL tail reads and file mtime to classify activity (`active`, `ready`, `idle`, `waiting_input`, `blocked`).

References:

- `.docs/agent-orchestrator/packages/plugins/agent-codex/src/index.ts:345`
- `.docs/agent-orchestrator/packages/plugins/agent-codex/src/index.ts:357`
- `.docs/agent-orchestrator/packages/plugins/agent-codex/src/index.ts:413`
- `.docs/agent-orchestrator/packages/plugins/agent-claude-code/src/index.ts:205`
- `.docs/agent-orchestrator/packages/plugins/agent-claude-code/src/index.ts:603`
- `.docs/agent-orchestrator/packages/core/src/utils.ts:87`

## Mapping to Our Dumped Session-State Files

### Directly used by AO stability today

- `events.jsonl`: Not used in AO Codex stability path.
- `workspace.yaml`: Not used.
- `session.db`: Not used.
- `checkpoints/index.md`: Not used.
- `plan.md`: Not used.

### Indirectly analogous behavior

- AO Claude plugin does use JSONL log tails for activity, but this is Claude-specific under `~/.claude/projects/...`, not Copilot session-state JSONL.
- AO core always falls back to runtime liveness + metadata state as the stability backbone.

## Practical Implication for Our Forensics
For this ingestion cycle, the dumped Copilot session-state artifacts are high-value for forensic traceability and post-incident analysis, but they do not mirror AO's current Codex stability mechanism.

If we choose to adopt AO-like stability patterns in `v1`, the closest parity baseline is:

1. Treat runtime liveness + persisted metadata as authoritative control-plane state.
2. Use session-state files (`events.jsonl`, etc.) as observability/enrichment signals, not sole truth, unless we explicitly build deterministic parsers and fallback rules.
3. Define strict restore eligibility rules (terminal vs non-restorable) before enabling automated recovery.
