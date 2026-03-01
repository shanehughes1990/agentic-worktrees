package stream

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"context"
)

const DefaultReplayLimit = 200

type EventStore interface {
	Append(ctx context.Context, event domainstream.Event) (domainstream.Event, error)
	ListFromOffset(ctx context.Context, offset uint64, limit int) ([]domainstream.Event, error)
}
