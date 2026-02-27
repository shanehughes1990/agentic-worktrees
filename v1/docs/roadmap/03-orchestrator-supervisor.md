# Slice 03 — Orchestrator Supervisor

## Objective

Implement the central orchestration supervisor responsible for policy decisions, global lifecycle tracking, and escalation triggers.

## Task Checklist

- [ ] Define supervisor domain model (states, transitions, invariants).
- [ ] Define supervisor event taxonomy with typed reason codes.
- [ ] Implement application-layer policy engine for attention-zone transitions.
- [ ] Implement correlation-ID propagation across supervisor decisions.
- [ ] Add deterministic transition tests over real worker/scm signal fixtures.
- [ ] Expose supervisor decision history for query/subscription layers.

## Deliverables

- Supervisor state model for sessions/workflows.
- Decision policy model for attention zones and action transitions.
- Event taxonomy for supervisor decisions and lifecycle transitions.
- Correlation-ID propagation contract across supervisor flows.

## In Scope

- Domain/application contracts for supervisor state and decisions.
- Deterministic transition handling and error classification.
- Event emission contracts consumed by stream/control-plane layers.
- Policy tests over real SCM/worker execution context (not synthetic-only assumptions).

## Out of Scope

- Full provider adapter implementations.
- Final UI rendering logic.

## Acceptance Criteria

- Transition rules are deterministic and testable.
- Supervisor decisions can be queried/traced by correlation IDs.
- Policies are provider-agnostic and do not depend on concrete SDKs.
- Supervisor behavior is validated against SCM-backed execution signals.

## Dependencies

- Slice 00 complete.
- Slice 01 (`agent` + `scm` core) complete.
- Slice 02 (worker execution plane) baseline complete.
- Stream contract alignment from slice 05.

## Exit Check

This slice is complete when supervisor decisions can be explained and replayed from emitted lifecycle events in real execution context.
