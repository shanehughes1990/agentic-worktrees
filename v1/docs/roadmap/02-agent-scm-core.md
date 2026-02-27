# Slice 02 — Agent + SCM Core

## Objective

Define and implement the `agent` and `scm` slots as first-class V1 contracts with initial providers.

## Scope

- `agent` slot contract for execution and session introspection.
- `scm` slot contract for branch context, PR lifecycle, and review state.
- SCM bootstrap flow primitives used by local and remote execution.
- Initial provider adapter(s) for current target SCM.

## Acceptance Criteria

- Agent calls are mediated by slot interfaces, not concrete SDKs in core flow.
- SCM operations support source branch context and commit/PR linkage.
- Remote worker bootstrap prerequisites are available via SCM contract.

## Dependencies

- Worker execution-plane contract.
- Orchestrator lifecycle policy.
