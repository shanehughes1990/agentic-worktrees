package taskengine

import "context"

// CheckpointStore persists and retrieves step checkpoints so retries can resume
// from the last completed step without the orchestrator re-embedding them in payloads.
type CheckpointStore interface {
	Save(ctx context.Context, idempotencyKey string, checkpoint Checkpoint) error
	Load(ctx context.Context, idempotencyKey string) (*Checkpoint, error)
}
