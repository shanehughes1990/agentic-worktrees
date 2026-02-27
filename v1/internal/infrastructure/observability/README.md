# Observability Platform (V1)

## Responsibility

This package is responsible for one thing: provide a single observability platform for a service runtime.

It owns:

- service-scoped logging
- operation-scoped tracing
- operation metrics
- correlation propagation across runtime boundaries
- unified bootstrap and shutdown lifecycle

It does this through one exposed surface in `observability.go`.

## Use Cases

### 1) Bootstrap service observability once

Use when a process starts (API, worker, scheduler).

- call `Bootstrap(ctx, Config)`
- keep returned `Platform` for runtime usage
- call `Platform.Shutdown(ctx)` on process exit

### 2) Emit service lifecycle logs

Use for startup/health/infra events that are not tied to one operation.

- use `Platform.ServiceEntry()`
- log service-level messages with service identity fields attached

### 3) Emit contextual logs

Use when a context already contains correlation IDs and trace state.

- use `Platform.Entry(ctx)`
- log with automatic `run_id`, `task_id`, `job_id`, `trace_id`, `span_id`

### 4) Instrument a business operation

Use for a unit of work that must be traced and measured.

- call `Platform.StartOperation(ctx, name)`
- use `Operation.Entry()` for operation logs
- use `Operation.Context()` for downstream calls
- finish with `Operation.End(err)`

### 5) Correlate across API/worker boundaries

Use when handing work between runtimes.

- set IDs with `WithCorrelationIDs(ctx, ids)`
- recover IDs with `CorrelationIDsFromContext(ctx)`
- preserve IDs in payload/envelope and restore in the next runtime

## Public Surface

Primary exposed types:

- `Config`
- `Platform`
- `Operation`
- `Entry`
- `CorrelationIDs`

Primary exposed constructor:

- `Bootstrap(ctx, config)`

## Configuration Scope

`Config` controls:

- identity: `ServiceName`, `Environment`, `Version`
- logging: `LogFormat`, `LogLevel`, `TimestampFormat`, `PrettyPrintJSON`
- telemetry transport: `OTLPEndpoint`, `OTLPHeaders`

Behavior:

- if `OTLPEndpoint` is empty, telemetry runs in-memory
- if `OTLPEndpoint` is set, OTLP/HTTP exporters are used

## Non-Responsibilities

This package does not own:

- business logic orchestration
- queue payload schema design
- domain-level error policy
- external monitoring dashboards/alerts configuration

## Implementation Files

- `observability.go`: exposed platform contract
- `logrus.go`: internal logging implementation
- `otel.go`: internal telemetry implementation
