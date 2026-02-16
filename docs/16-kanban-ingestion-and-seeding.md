# 16 - Kanban Ingestion and Board Seeding Pipeline

## Purpose

Define the application path that ingests a scope/task/feature input file, inspects the current project state, derives dependency-safe task trees, and seeds the mini kanban board for parallel execution.

Primary intended flow: generate board, start autonomous run, and allow system to complete the board cycle with minimal operator involvement.

## Objectives

- provide internal domain objects used by planner/workflow logic
- export all required JSON artifacts for the file-based board backend
- preserve compatibility with future relational database table storage
- produce deterministic, repeatable outputs from the same inputs and project revision

## Inputs

- scope file path (feature/scope/task specification)
- repository path
- target branch (defaults to current main repository branch)
- optional planning constraints:
  - max parallel lanes
  - priority policy
  - labeling policy
  - risk policy

## Required Ingestion Path

`scope file -> repository analysis -> task decomposition -> dependency DAG/tree -> parallel lane assignment -> board seed artifacts`

## Pipeline Stages

### Stage 1: Load and Validate Scope

- parse scope file format (versioned schema)
- validate required fields and acceptance criteria presence
- validate scope schema version is supported or migratable
- normalize language into canonical planning model
- up-convert older supported schema versions into current canonical model

Output:

- `ScopeSpec` (canonical representation)

### Stage 2: Project State Discovery

- inspect repository tree and relevant files
- detect existing implementation surface and prior task artifacts
- collect baseline metadata at current branch HEAD commit

Output:

- `ProjectSnapshot` with commit SHA, components, and detected capabilities

### Stage 3: Task Decomposition

- break scope into implementable units
- generate explicit task metadata:
  - title, description, labels, priority
  - acceptance criteria references
  - target paths/components
  - estimated complexity and risk class
  - optional runtime model override in PRD metadata

Output:

- candidate task set

### Stage 4: Dependency Graph Construction

- build DAG from task prerequisites
- allow parent/child tree edges only if acyclic
- run cycle/orphan checks and reject invalid graph

Output:

- validated dependency graph + tree projections

### Stage 5: Parallel Lane Planning

- compute executable layers by dependency depth
- assign tasks into parallel lanes subject to constraints
- annotate lane ordering and expected concurrency windows

Output:

- lane plan + runnable frontier set

### Stage 6: Board Materialization

- create internal board domain objects for application use
- create exported JSON artifacts for board backend seeding
- include deterministic IDs and revision metadata
- emit current canonical `schema_version` on all generated artifacts

Output:

- in-memory `BoardModel`
- on-disk seed files under `tasks/`

## JSON Seed Artifacts (Current Backend)

Minimum files to generate:

- `tasks/board.json`
- `tasks/prd/<task-id>.json` for each task
- `tasks/ingestion-report.json`

### `board.json` Required Sections

- schema version
- source scope file fingerprint
- repository commit SHA used for planning
- lists/lanes
- tasks with dependencies and metadata
- planning revision and generation timestamp
- model policy reference (default `gpt-5.3-codex`)

### `ingestion-report.json` Required Sections

- input summary
- discovered project context summary
- decomposition stats (task count, dependency count)
- unresolved ambiguities or assumptions
- warnings and blocked items
- schema compatibility section (`input_version`, `effective_version`, `migration_applied`)

## Internal vs Export Model Boundary

The ingest operation must produce both:

- internal models for runtime orchestration (planner + workflow)
- exported persistence models for external storage adapters

Mapping must be explicit and versioned to allow future migration to database tables.

Model selection metadata mapping must include:

- default model policy (`gpt-5.3-codex`)
- per-task override field
- validation result and fallback indicator

## Future Database Mapping Contract

The same canonical entities must map cleanly to future relational tables:

- `boards`
- `tasks`
- `task_dependencies`
- `task_labels`
- `task_prd_documents`
- `ingestion_runs`
- `ingestion_warnings`

Primary keys and stable IDs generated during ingestion must be backend-agnostic.

## Determinism and Idempotency Rules

- same scope file + same repository commit + same config => same task graph IDs and lane assignments
- rerunning ingestion writes a new planning revision but preserves stable task identity where unchanged
- no destructive overwrite without explicit `--force` behavior
- schema migrations must be deterministic and produce stable canonical output

## CLI Contract (Documentation-Level)

The CLI must expose a dedicated ingestion path (example shape):

- `kanban ingest --scope <file> --repo <path> [--branch <name>] [--out tasks/]`

Autonomous execution contract (example shape):

- `kanban run --from-board tasks/board.json --autonomous`

Expected behavior:

1. validate scope and project readiness
2. run staged pipeline
3. emit internal planning summary
4. write seed artifacts
5. print deterministic report with counts and warnings

For autonomous run:

- capture origin target branch at run start
- execute full board until completion or terminal escalation
- merge successful task PRs into captured origin target branch

Model-specific behavior:

- if task PRD metadata includes model override, validate against allowlist
- if unsupported, record warning and set effective model to `gpt-5.3-codex`

## Failure Modes Specific to Ingestion

- invalid scope schema
- unsupported schema version outside compatibility window
- repository not in clean/readable state
- dependency cycle generated during decomposition
- ambiguous decomposition requiring operator clarification
- seed write failure due to permissions or partial I/O

Required handling:

- classify transient vs terminal
- preserve diagnostic report for failed ingestion run
- do not publish partial board as active revision

## Realtime Integration

When ingestion creates or updates board artifacts:

- emit canonical events for board/prd/report changes
- include ingestion run ID for traceability
- notify all active subscribers so planner/workers can refresh safely

## Security and Safety Notes

- never execute arbitrary scope-file code
- sanitize all extracted path references
- enforce repository path boundaries during analysis
- redact sensitive data from ingestion reports
