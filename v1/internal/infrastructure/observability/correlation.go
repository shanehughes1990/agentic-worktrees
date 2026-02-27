package observability

import "context"

// CorrelationIDs holds workflow-level IDs used to join logs, traces, and metrics.
//
// This structure is a living expansion point for the project's correlation
// centralization model and is intentionally unstable. Fields are expected to evolve
// as orchestration scope changes, and consumers should treat this contract as
// adjustable over time.
type CorrelationIDs struct {
	RunID  string
	TaskID string
	JobID  string
}

type correlationContextKey struct{}

// WithCorrelationIDs stores correlation IDs on the context used by observability
// entry and telemetry emission paths.
func WithCorrelationIDs(ctx context.Context, ids CorrelationIDs) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, correlationContextKey{}, ids)
}

// CorrelationIDsFromContext reads correlation IDs from context.
//
// It returns an empty CorrelationIDs value when IDs are not present.
func CorrelationIDsFromContext(ctx context.Context) CorrelationIDs {
	if ctx == nil {
		return CorrelationIDs{}
	}
	ids, ok := ctx.Value(correlationContextKey{}).(CorrelationIDs)
	if !ok {
		return CorrelationIDs{}
	}
	return ids
}
