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

```go
package main

import (
 "context"

 "agentic-orchestrator/internal/infrastructure/observability"
)

func bootstrapObservability(ctx context.Context) (*observability.Platform, error) {
 platform, err := observability.Bootstrap(ctx, observability.Config{
  LogFormat:    observability.LogFormatText,
  LogLevel:     observability.LogLevelInfo,
  OTLPEndpoint: "", // empty => in-memory telemetry
 })
 if err != nil {
  return nil, err
 }
 return platform, nil
}
```

### 2) Service-level logging

Use for startup/health/infra events that are not tied to one operation.

```go
func logServiceLifecycle(platform *observability.Platform) {
 platform.ServiceEntry().Info("service starting")
 platform.ServiceEntry().WithField("component", "http").Info("listener ready")
}
```

### 3) Entry-level contextual logging

Use when a context already contains correlation IDs and/or trace state.

```go
func logWithContext(platform *observability.Platform, ctx context.Context) {
 ctx = observability.WithCorrelationIDs(ctx, observability.CorrelationIDs{
  RunID:  "run-123",
  TaskID: "task-456",
  JobID:  "job-789",
 })

 platform.Entry(ctx).
  WithField("stage", "dispatch").
  Info("task submitted")
}
```

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
- logging: `LogFormat`, `LogLevel`, `PrettyPrintJSON`
- telemetry transport: `OTLPEndpoint`, `OTLPHeaders`

Behavior:

- identity defaults when omitted:
  - `ServiceName`: `unknown`
  - `Environment`: `local`
  - `Version`: `development`
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
