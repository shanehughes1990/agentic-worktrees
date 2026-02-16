# 10 - Failure Modes and Recovery

## Reliability Target

Design for graceful degradation and deterministic recovery rather than assuming ideal conditions.

## Failure Handling Objectives

- preserve correctness over throughput
- fail closed on ambiguous high-risk actions
- guarantee resumability from durable checkpoints
- isolate blast radius to run/task/worktree scope
- produce audit-quality evidence for every terminal failure

## Severity Levels

- `SEV-0` data integrity or unsafe merge risk
- `SEV-1` automation blocked but recoverable with operator action
- `SEV-2` transient degradation with automatic recovery expected
- `SEV-3` cosmetic or non-blocking telemetry/UI issues

## Universal Detection Signals

- lifecycle heartbeat timeout
- repeated retries crossing policy threshold
- queue latency and backlog growth beyond policy threshold
- revision/cursor mismatch in event streams
- write/read checksum mismatch for persisted records
- state transition attempted from invalid prior state
- policy gate failure on rebase/merge action

## End-to-End Failure Point Catalog

### 1) Initialization and Bootstrapping

Failure points:

- missing or invalid config
- unsupported backend selected
- seed template files malformed or absent
- secrets unavailable at startup
- `git` version below minimum supported
- `gh` version below minimum supported
- `gh` authentication missing/expired/insufficient for PR operations
- default model policy missing or invalid (must resolve to `gpt-5.3-codex`)
- PRD model override references unsupported model identifier
- persisted artifact schema version outside supported compatibility window

Detection:

- startup validation phase
- schema/version checks

Recovery:

- fail startup before queue activation
- emit actionable diagnostics with exact missing keys/files
- support `init/seed` rerun after correction
- block task intake when dependency/auth preflight is unhealthy
- support explicit operator-triggered recheck after remediation
- fallback unsupported PRD model override to `gpt-5.3-codex` and continue with warning
- run schema migration/normalization before re-enabling task intake

Escalate to terminal when:

- config is irreconcilable for chosen runtime mode

### 2) Task Intake and Dependency Graph

Failure points:

- invalid task IDs or duplicate IDs
- dependency cycles
- orphan dependency references
- parent/child graph conflicts with DAG constraints

Detection:

- graph validation on every mutation
- periodic full graph audit

Recovery:

- reject invalid write atomically
- keep last known-good graph revision active
- quarantine invalid mutation event for operator review

Escalate to terminal when:

- persisted graph cannot be parsed into a consistent DAG

### 3) Planner and Enqueue Control

Failure points:

- duplicate dispatch due to race
- starvation of low-priority tasks
- planner lock contention
- stale readiness view from lagging store reads

Detection:

- unique run lease checks
- queue age and fairness metrics
- lock acquisition latency alarms

Recovery:

- lease-based single ownership with TTL
- planner rebalance pass
- force refresh from latest board revision before dispatch

Escalate to terminal when:

- repeated duplicate dispatch indicates broken exclusivity guarantees

### 3A) Asynq Queue and Worker Core

Failure points:

- Redis unavailable or unstable
- job stuck in pending due to worker starvation
- retry storm from misclassified terminal errors
- poisoned job repeatedly failing fast
- duplicate side effects due to non-idempotent handler behavior

Detection:

- queue depth/latency SLO monitors
- worker heartbeat and throughput drop alerts
- per-job retry histogram and anomaly detection

Recovery:

- hold enqueueing for affected queues during Redis outage
- autoscale workers or rebalance queue weights under backlog pressure
- classify errors correctly (`transient` vs `terminal`) and tune retry limits
- quarantine poisoned jobs to dead-letter/archive with payload fingerprint
- enforce idempotency keys and checkpoint guards in handlers

Escalate to terminal when:

- Redis durability guarantees cannot be restored within recovery window
- queue semantics cannot guarantee safe continuation of job processing

### 4) Persistence Backend (JSON-first, portable model)

Failure points:

- partial/torn file write
- concurrent write overwrite
- corrupted JSON record
- filesystem permission denial
- storage medium full or read-only remount

Detection:

- atomic write protocol verification
- optimistic revision/version checks
- checksum/hash validation for critical records

Recovery:

- write using tmp + fsync + atomic rename
- on revision conflict: reload, re-apply mutation, retry bounded times
- quarantine corrupted record, load last valid snapshot, replay valid events
- backoff and alert on permission/device errors

Escalate to terminal when:

- no valid snapshot/event chain remains reconstructible

### 5) Realtime Watch and Fan-Out

Failure points:

- file watcher drops events
- subscriber backlog overflow
- out-of-order event observation across sources
- broadcaster deadlock from slow consumers

Detection:

- monotonic cursor validation per source
- queue depth and lag metrics per subscriber
- watchdog for stalled broadcaster loop

Recovery:

- auto re-subscribe with jittered backoff
- degrade slow subscribers into snapshot+replay resync mode
- non-blocking fan-out with bounded channels and drop counters
- enforce per-source revision ordering and gap recovery

Escalate to terminal when:

- canonical stream cannot provide consistent replay semantics

### 6) Worktree Lifecycle

Failure points:

- path collision for worktree directory
- detached HEAD or wrong branch mapping
- stale/orphaned worktrees after crash
- cleanup failure due to active file handles

Detection:

- preflight path and branch identity checks
- reconciliation sweep on startup

Recovery:

- deterministic naming with collision-resistant suffixes
- branch/worktree remap from durable run metadata
- adopt-or-clean policy for orphaned worktrees
- deferred cleanup retry with backoff

Escalate to terminal when:

- repository/worktree mapping is irrecoverably inconsistent

### 7) Git Fetch/Rebase/Commit/Push

Failure points:

- fetch fails (network/auth)
- rebase conflict or interrupted rebase state
- non-fast-forward push rejection
- commit cannot be created due to index/worktree inconsistency

Detection:

- explicit exit-code classification
- repository health checks before each git step

Recovery:

- transient failures: retry with bounded exponential backoff
- interrupted rebase: abort, restore checkpointed branch state, retry
- non-fast-forward: fetch + rebase + push retry loop (bounded)
- integrity failure: reset to checkpointed revision and re-run guarded steps

Escalate to terminal when:

- repeated rebase failure exceeds policy budget
- git state remains invalid after recovery actions

### 8) Pull Request and Merge Automation

Failure points:

- PR API unavailable or rate-limited
- mergeability unknown for prolonged duration
- required checks never report terminal status
- merge attempt rejected by policy/rules

Detection:

- API status and retry headers
- check-state timeout monitors

Recovery:

- idempotent PR create-or-update semantics
- adaptive poll/backoff for check status
- auto rebase and re-evaluate mergeability once per iteration
- route unresolved policy/rule failures to manual review queue

Escalate to terminal when:

- PR cannot be merged within bounded retries and SLA window

### 9) Agent Runtime and Execution Control

Failure points:

- startup timeout
- heartbeat loss mid-run
- malformed artifact payload
- command/tool policy violation
- runaway resource usage (CPU/memory)

Detection:

- startup and heartbeat timers
- artifact schema validation
- policy engine violation events
- runtime resource budget monitors

Recovery:

- graceful stop then hard kill on timeout ladder
- restart from last safe checkpoint when retry budget remains
- persist partial outputs before teardown
- quarantine run on policy violation; require explicit requeue

Escalate to terminal when:

- repeated runtime crashes exceed task retry budget

### 10) Container and Host Runtime

Failure points:

- container restart loop
- read-only filesystem where write path expected
- volume detachment/corruption
- clock skew affecting ordering assumptions

Detection:

- liveness/readiness failures
- filesystem and volume health probes
- timestamp skew checks

Recovery:

- hold queue in safe mode until readiness restored
- remount/rebind writable volumes
- restore from latest durable snapshot and event ledger
- normalize time references to UTC and monotonic counters

Escalate to terminal when:

- durable state volume is unavailable beyond recovery timeout

### 11) Security and Identity

Failure points:

- expired/revoked credentials
- secret rotation drift
- unauthorized action attempt
- audit log write failure

Detection:

- auth error classification
- credential expiry preflight checks
- policy denial telemetry

Recovery:

- pause privileged operations, continue read-only diagnostics
- rotate credentials and revalidate scopes
- block unsafe actions by default
- buffer audit events locally and flush on recovery

Escalate to terminal when:

- audit durability cannot be guaranteed for critical actions

### 12) Control Plane and Operator Actions

Failure points:

- duplicate cancel/retry commands
- conflicting operator intents
- partial application of pause/drain commands

Detection:

- command idempotency keys
- control-plane state reconciliation loop

Recovery:

- idempotent command handling
- serialize control actions through authoritative coordinator
- explicit ack/nack for every command

Escalate to terminal when:

- control-plane state diverges from worker state with no safe reconciliation path

## Recovery Strategy Ladder

1. retry transiently (bounded)
2. resync from authoritative snapshot/event stream
3. restart component scope (worker/run)
4. quarantine task/run to dead-letter with diagnostics
5. escalate to operator with exact remediation steps

## Dead-Letter Requirements

Each dead-letter entry must include:

- task/run IDs and last successful checkpoint
- failure class, severity, and terminal reason
- attempted recovery actions and counts
- pointers to relevant logs/events/artifacts
- recommended next action (retry manually, edit task graph, resolve conflict, rotate credential)

## Checkpoint and Resume Rules

- never re-run side effects without idempotency key validation
- checkpoint before and after every external side-effect boundary
- on restart, reconcile desired state vs observed state before resuming
- orphaned resources are adopted or cleaned according to policy

## Data Integrity and Consistency Guards

- optimistic concurrency using revision fields
- append-only event ledger for reconstruction
- schema versioning and migration gates
- periodic snapshot plus replay verification drills

## Observability for Failure Management

Required telemetry for every failure event:

- component and operation name
- error class (`transient`, `recoverable`, `terminal`)
- retry attempt/limit
- checkpoint reference
- recovery action chosen and outcome

## Failure Drills (Expanded)

- kill orchestrator during each workflow phase and verify deterministic resume
- inject malformed JSON into board/PRD/event files and verify quarantine+rebuild path
- simulate watcher event loss and verify resync logic
- induce rebase conflicts repeatedly until escalation threshold
- simulate API rate limiting and validate adaptive backoff behavior
- force disk saturation on logs/worktrees and verify safe degradation
- revoke credentials mid-run and verify privileged action freeze
- break control-plane connectivity and verify command reconciliation

## Runbook Summary

1. classify failure and severity
2. validate integrity of checkpoint + revision state
3. execute strategy ladder starting from least invasive recovery
4. quarantine and dead-letter when policy threshold is reached
5. emit post-incident record and prevention follow-up action
