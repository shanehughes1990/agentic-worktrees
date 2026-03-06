# Asynq Exit Coverage Matrix

This matrix is the checked-in source of truth for worker action exit-path lifecycle coverage.

Contract enforced: `ingest -> persist (project_session_history/project_sessions) -> emit upstream`.

## Coverage

| Worker Action (`job_kind`)        | Handler                               | Enqueued                       | Started | Retry Started           | Heartbeats                                          | Completed | Failed | Terminated                        | Retry Scheduled                             | Dead Lettered                                     | Persisted-First Emit                             |
| --------------------------------- | ------------------------------------- | ------------------------------ | ------- | ----------------------- | --------------------------------------------------- | --------- | ------ | --------------------------------- | ------------------------------------------- | ------------------------------------------------- | ------------------------------------------------ |
| `ingestion.agent.run`             | `IngestionAgentHandler`               | Yes (admission ledger + queue) | Yes     | Yes (`retry_count > 0`) | Yes (`heartbeat`, layered runtime/process/activity) | Yes       | Yes    | Yes (`context canceled/deadline`) | Yes (retryable failure + attempt remaining) | Yes (terminal failure class or retries exhausted) | Yes (lifecycle append publishes persisted event) |
| `agent.workflow.run`              | `AgentWorkflowHandler`                | Yes (admission ledger + queue) | Yes     | Yes (`retry_count > 0`) | Yes (`heartbeat`, layered runtime/process/activity) | Yes       | Yes    | Yes (`context canceled/deadline`) | Yes (retryable failure + attempt remaining) | Yes (terminal failure class or retries exhausted) | Yes (lifecycle append publishes persisted event) |
| `scm.workflow.run`                | `SCMWorkflowHandler`                  | Yes (admission ledger + queue) | Yes     | Yes (`retry_count > 0`) | Yes (`heartbeat`, layered runtime/process/activity) | Yes       | Yes    | Yes (`context canceled/deadline`) | Yes (retryable failure + attempt remaining) | Yes (terminal failure class or retries exhausted) | Yes (lifecycle append publishes persisted event) |
| `project.document.upload.prepare` | `ProjectDocumentPrepareUploadHandler` | Yes (admission ledger + queue) | Yes     | Yes (`retry_count > 0`) | Yes (`heartbeat`, layered runtime/process/activity) | Yes       | Yes    | Yes (`context canceled/deadline`) | Yes (retryable failure + attempt remaining) | Yes (terminal failure class or retries exhausted) | Yes (lifecycle append publishes persisted event) |
| `project.document.delete`         | `ProjectDocumentDeleteHandler`        | Yes (admission ledger + queue) | Yes     | Yes (`retry_count > 0`) | Yes (`heartbeat`, layered runtime/process/activity) | Yes       | Yes    | Yes (`context canceled/deadline`) | Yes (retryable failure + attempt remaining) | Yes (terminal failure class or retries exhausted) | Yes (lifecycle append publishes persisted event) |
| `prompt.refinement.agent.run`     | `PromptRefinementHandler`             | Yes (admission ledger + queue) | Yes     | Yes (`retry_count > 0`) | Yes (`heartbeat`, layered runtime/process/activity) | Yes       | Yes    | Yes (`context canceled/deadline`) | Yes (retryable failure + attempt remaining) | Yes (terminal failure class or retries exhausted) | Yes (lifecycle append publishes persisted event) |

## Notes

- Terminal lifecycle events are now attempt-bound and explicit: `completed`, `failed`, or `terminated`.
- Attempt disposition events are explicit and persisted:
  - `retry_scheduled` when retryable and retries remain.
  - `dead_lettered` when failure class is terminal or retries are exhausted.
- Queue policy mapping uses typed failure class:
  - terminal class -> `asynq.SkipRetry`
  - transient/unknown class -> retry policy continues.
- Session snapshot `ended_at` is driven by terminal lifecycle event types (`completed`, `failed`, `terminated`, `dead_lettered`).
