# Slice 08 — Container-First Deployment

## Objective

Make containerized runtime the default deployment contract for API and worker components, including first-class Postgres persistence operations.

## Task Checklist

- [ ] Add first-class Dockerfiles for API and worker runtimes.
- [ ] Define compose topology including Postgres and Redis (Redis for queue transport only).
- [ ] Externalize env/secrets contract for both services, including `DATABASE_DSN`.
- [ ] Add migration/init step for Postgres schema management.
- [ ] Add health/startup/shutdown checks for orchestrated runtime (API, worker, Postgres dependencies).
- [ ] Add containerized integration checks for key execution and persistence paths.
- [ ] Define versioned image artifact/release process.

## Deliverables

- First-class container images for API/control plane and worker runtime.
- Compose topology for local integration parity with Postgres + Redis.
- Externalized env/secrets runtime contract.
- Postgres migration/init operational contract.
- Health/lifecycle wiring for orchestrated startup and shutdown.

## In Scope

- Runtime packaging and orchestration baseline.
- Local/remote environment parity through configuration-only differences.
- Containerized validation for key execution and persistence paths.

## Out of Scope

- Environment-specific infrastructure provisioning templates beyond baseline requirements.
- Non-container primary deployment targets.

## Acceptance Criteria

- Local integration uses compose as default path.
- Same images run across local and remote contexts with config changes only.
- Worker + API startup/shutdown and health checks behave consistently in containers.
- Durable operational state remains in Postgres across container restarts.

## Dependencies

- Slices 04 and 06.
- Observability and health platform integration.

## Exit Check

This slice is complete when container artifacts are the primary release vehicle and Postgres-backed runtime behavior is validated end-to-end.
