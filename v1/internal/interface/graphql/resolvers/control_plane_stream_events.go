package resolvers

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"context"
	"fmt"
	"strings"
	"time"
)

func publishTaskboardStreamEvent(ctx context.Context, streamService interface {
	AppendAndPublish(ctx context.Context, event domainstream.Event) (domainstream.Event, error)
}, eventType domainstream.EventType, projectID string, payload map[string]any) {
	if streamService == nil {
		return
	}
	cleanProjectID := strings.TrimSpace(projectID)
	if cleanProjectID == "" {
		return
	}
	_, _ = streamService.AppendAndPublish(ctx, domainstream.Event{
		EventID:    fmt.Sprintf("taskboard-%d", time.Now().UTC().UnixNano()),
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceACP,
		EventType:  eventType,
		CorrelationIDs: domainstream.CorrelationIDs{
			ProjectID:     cleanProjectID,
			CorrelationID: fmt.Sprintf("project:%s", cleanProjectID),
		},
		Payload: payload,
	})
}
