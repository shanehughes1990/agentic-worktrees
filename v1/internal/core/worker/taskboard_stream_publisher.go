package worker

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"context"
	"fmt"
	"strings"
	"time"
)

type taskboardStreamPublisher struct {
	streamService interface {
		AppendAndPublish(ctx context.Context, event domainstream.Event) (domainstream.Event, error)
	}
}

func newTaskboardStreamPublisher(streamService interface {
	AppendAndPublish(ctx context.Context, event domainstream.Event) (domainstream.Event, error)
}) *taskboardStreamPublisher {
	if streamService == nil {
		return nil
	}
	return &taskboardStreamPublisher{streamService: streamService}
}

func (publisher *taskboardStreamPublisher) PublishTaskboardUpdated(ctx context.Context, projectID string, boardID string, runID string) error {
	if publisher == nil || publisher.streamService == nil {
		return nil
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanProjectID == "" || cleanBoardID == "" {
		return nil
	}
	occurredAt := time.Now().UTC()
	_, err := publisher.streamService.AppendAndPublish(ctx, domainstream.Event{
		EventID:    fmt.Sprintf("taskboard-%d", occurredAt.UnixNano()),
		OccurredAt: occurredAt,
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventTaskboardUpdated,
		CorrelationIDs: domainstream.CorrelationIDs{
			RunID:         strings.TrimSpace(runID),
			ProjectID:     cleanProjectID,
			CorrelationID: fmt.Sprintf("project:%s", cleanProjectID),
		},
		Payload: map[string]any{
			"project_id": cleanProjectID,
			"board_id":   cleanBoardID,
		},
	})
	if err != nil {
		return fmt.Errorf("append and publish taskboard updated stream event: %w", err)
	}
	return nil
}
