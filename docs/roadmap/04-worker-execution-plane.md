# Slice 04 — Worker Execution Plane

## Objective

Deliver local and remote worker parity with SCM-backed bootstrap and resumable execution.

## Scope

- Worker registration and capability advertisement.
- Dispatch + lease model with resumable checkpoints.
- Local worker adapter.
- Remote worker adapter contract and execution handshake.
- SCM-backed remote bootstrap sequence:
  - authenticate with SCM credentials
  - clone/pull from origin
  - checkout/pull source branch
  - execute from fetched source state

## Acceptance Criteria

- Jobs can execute on local or remote workers through one dispatch flow.
- Remote workers do not rely on pre-existing host checkout.
- Checkpoints support resume after transient failures.

## Dependencies

- Agent and SCM contracts.
- Container runtime profiles.
