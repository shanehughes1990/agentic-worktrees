# Slice 08 — Container-First Deployment

## Objective

Make containerized runtime the default deployment contract for API and worker components.

## Task Checklist

- [ ] Add first-class Dockerfiles for API and worker runtimes.
- [ ] Define compose topology for local integration parity.
- [ ] Externalize env/secrets contract for both services.
- [ ] Add health/startup/shutdown checks for orchestrated runtime.
- [ ] Add containerized integration checks for key execution paths.
- [ ] Define versioned image artifact/release process.

## Deliverables

- First-class container images for API/control plane and worker runtime.
- Compose topology for local integration parity.
- Externalized env/secrets runtime contract.
- Health/lifecycle wiring for orchestrated startup and shutdown.
- Versioned image artifact flow for release.

## In Scope

- Runtime packaging and orchestration baseline.
- Local/remote environment parity through configuration-only differences.
- Containerized validation for key execution paths.

## Out of Scope

- Environment-specific infrastructure provisioning templates beyond baseline requirements.
- Non-container primary deployment targets.

## Acceptance Criteria

- Local integration uses compose as default path.
- Same images run across local and remote contexts with config changes only.
- Worker + API startup/shutdown and health checks behave consistently in containers.

## Dependencies

- Slices 04 and 06.
- Observability and health platform integration.

## Exit Check

This slice is complete when container artifacts are the primary release vehicle and validated runtime path.
