package resolvers

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationstream "agentic-orchestrator/internal/application/stream"
	"agentic-orchestrator/internal/application/taskengine"
	domainstream "agentic-orchestrator/internal/domain/stream"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

func streamSubscription(ctx context.Context, streamService *applicationstream.Service, correlation models.CorrelationInput, fromOffset *int32, eventFilter func(domainstream.Event) bool) (<-chan models.StreamEventResult, error) {
	if streamService == nil {
		return singleEventStream(models.GraphError{Code: models.GraphErrorCodeUnavailable, Message: "stream service is not configured"}), nil
	}
	offset := int32ToInt(fromOffset)
	if offset < 0 {
		offset = 0
	}
	replay, err := streamService.ReplayFromOffset(ctx, uint64(offset), 500)
	if err != nil {
		return singleEventStream(graphErrorFromError(fmt.Errorf("replay stream events: %w", err))), nil
	}
	output := make(chan models.StreamEventResult, 64)
	_, live, cancel := streamService.Subscribe(256)
	_, changes, cancelChanges := streamService.SubscribeChanges(64)
	go func() {
		defer cancel()
		defer cancelChanges()
		defer close(output)

		lastOffset := uint64(offset)
		emitBatch := func(events []domainstream.Event) bool {
			for _, event := range events {
				if event.StreamOffset > lastOffset {
					lastOffset = event.StreamOffset
				}
				if !matchesCorrelation(event, correlation) || !eventFilter(event) {
					continue
				}
				select {
				case output <- models.StreamEventSuccess{Event: mapStreamEvent(event)}:
				case <-ctx.Done():
					return false
				}
			}
			return true
		}

		if !emitBatch(replay) {
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, open := <-live:
				if !open {
					return
				}
				if !emitBatch([]domainstream.Event{event}) {
					return
				}
			case <-changes:
				events, replayErr := streamService.ReplayFromOffset(ctx, lastOffset, 500)
				if replayErr != nil {
					select {
					case output <- graphErrorFromError(fmt.Errorf("replay stream events: %w", replayErr)):
					case <-ctx.Done():
					}
					continue
				}
				if !emitBatch(events) {
					return
				}
			}
		}
	}()
	return output, nil
}

func streamLiveSubscription(ctx context.Context, streamService *applicationstream.Service, correlation models.CorrelationInput, fromOffset *int32, eventFilter func(domainstream.Event) bool) (<-chan models.StreamEventResult, error) {
	if streamService == nil {
		return singleEventStream(models.GraphError{Code: models.GraphErrorCodeUnavailable, Message: "stream service is not configured"}), nil
	}
	offset := int32ToInt(fromOffset)
	if offset < 0 {
		offset = 0
	}
	output := make(chan models.StreamEventResult, 64)
	_, live, cancel := streamService.Subscribe(256)
	go func() {
		defer cancel()
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case event, open := <-live:
				if !open {
					return
				}
				// Runtime live signals may not have a persisted stream offset yet.
				if event.StreamOffset > 0 && event.StreamOffset <= uint64(offset) {
					continue
				}
				if !matchesCorrelation(event, correlation) || !eventFilter(event) {
					continue
				}
				select {
				case output <- models.StreamEventSuccess{Event: mapStreamEvent(event)}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return output, nil
}

func streamQuery(ctx context.Context, streamService *applicationstream.Service, correlation models.CorrelationInput, fromOffset *int32, limit *int32, eventFilter func(domainstream.Event) bool) models.StreamEventsResult {
	if streamService == nil {
		return models.GraphError{Code: models.GraphErrorCodeUnavailable, Message: "stream service is not configured"}
	}
	offset := int32ToInt(fromOffset)
	if offset < 0 {
		offset = 0
	}
	resolvedLimit := int32ToInt(limit)
	if resolvedLimit <= 0 {
		resolvedLimit = 100
	}
	if resolvedLimit > 500 {
		resolvedLimit = 500
	}
	events, err := streamService.ReplayFromOffset(ctx, uint64(offset), resolvedLimit)
	if err != nil {
		return graphErrorFromError(fmt.Errorf("replay stream events: %w", err))
	}
	items := make([]*models.StreamEvent, 0, len(events))
	nextFromOffset := int32(offset)
	for _, event := range events {
		if !matchesCorrelation(event, correlation) || !eventFilter(event) {
			continue
		}
		if event.StreamOffset > uint64(nextFromOffset) {
			nextFromOffset = uint64ToInt32(event.StreamOffset)
		}
		items = append(items, mapStreamEvent(event))
	}
	return models.StreamEventsSuccess{Events: items, NextFromOffset: nextFromOffset}
}

func requireProjectScopedCorrelation(correlation models.CorrelationInput) error {
	if strings.TrimSpace(derefString(correlation.ProjectID)) == "" {
		return fmt.Errorf("project_id is required")
	}
	return nil
}

func matchesCorrelation(event domainstream.Event, correlation models.CorrelationInput) bool {
	if strings.TrimSpace(derefString(correlation.RunID)) != "" && strings.TrimSpace(event.CorrelationIDs.RunID) != strings.TrimSpace(derefString(correlation.RunID)) {
		return false
	}
	if strings.TrimSpace(derefString(correlation.TaskID)) != "" && strings.TrimSpace(event.CorrelationIDs.TaskID) != strings.TrimSpace(derefString(correlation.TaskID)) {
		return false
	}
	if strings.TrimSpace(derefString(correlation.JobID)) != "" && strings.TrimSpace(event.CorrelationIDs.JobID) != strings.TrimSpace(derefString(correlation.JobID)) {
		return false
	}
	if strings.TrimSpace(derefString(correlation.ProjectID)) != "" && strings.TrimSpace(event.CorrelationIDs.ProjectID) != strings.TrimSpace(derefString(correlation.ProjectID)) {
		return false
	}
	return true
}

func mapStreamEvent(event domainstream.Event) *models.StreamEvent {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		payload = []byte("{}")
	}
	expectedEventSeq, hasExpectedEventSeq := payloadInt32(event.Payload, "expected_event_seq")
	observedEventSeq, hasObservedEventSeq := payloadInt32(event.Payload, "observed_event_seq")
	return &models.StreamEvent{
		EventID:       event.EventID,
		StreamOffset:  uint64ToInt32(event.StreamOffset),
		OccurredAt:    event.OccurredAt.UTC(),
		RunID:         nilIfEmpty(event.CorrelationIDs.RunID),
		TaskID:        nilIfEmpty(event.CorrelationIDs.TaskID),
		JobID:         nilIfEmpty(event.CorrelationIDs.JobID),
		ProjectID:     nilIfEmpty(event.CorrelationIDs.ProjectID),
		SessionID:     nilIfEmpty(event.CorrelationIDs.SessionID),
		CorrelationID: strings.TrimSpace(event.CorrelationIDs.CorrelationID),
		Source:        toGraphStreamEventSource(event.Source),
		EventType:     string(event.EventType),
		GapDetected:   payloadBool(event.Payload, "gap_detected"),
		GapReconciled: payloadBool(event.Payload, "gap_reconciled"),
		ExpectedEventSeq: func() *int32 {
			if !hasExpectedEventSeq {
				return nil
			}
			value := expectedEventSeq
			return &value
		}(),
		ObservedEventSeq: func() *int32 {
			if !hasObservedEventSeq {
				return nil
			}
			value := observedEventSeq
			return &value
		}(),
		Payload: string(payload),
	}
}

func payloadBool(payload map[string]any, key string) bool {
	if payload == nil {
		return false
	}
	raw, ok := payload[key]
	if !ok || raw == nil {
		return false
	}
	value, ok := raw.(bool)
	return ok && value
}

func payloadInt32(payload map[string]any, key string) (int32, bool) {
	if payload == nil {
		return 0, false
	}
	raw, ok := payload[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch value := raw.(type) {
	case int:
		return int64ToInt32(int64(value)), true
	case int32:
		return value, true
	case int64:
		return int64ToInt32(value), true
	case float64:
		return int64ToInt32(int64(value)), true
	default:
		return 0, false
	}
}

func singleEventStream(result models.StreamEventResult) <-chan models.StreamEventResult {
	output := make(chan models.StreamEventResult, 1)
	output <- result
	close(output)
	return output
}

func int32ToInt(value *int32) int {
	if value == nil {
		return 0
	}
	return int(*value)
}

func uint64ToInt32(value uint64) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

func int64ToInt32(value int64) int32 {
	if value < math.MinInt32 {
		return math.MinInt32
	}
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

func toGraphJobKind(value taskengine.JobKind) (models.JobKind, error) {
	switch strings.TrimSpace(string(value)) {
	case string(taskengine.JobKindIngestionAgent):
		return models.JobKindIngestionAgentRun, nil
	case string(taskengine.JobKindAgentWorkflow):
		return models.JobKindAgentWorkflowRun, nil
	case string(taskengine.JobKindSCMWorkflow):
		return models.JobKindScmWorkflowRun, nil
	case string(taskengine.JobKindProjectDocumentUploadPrepare):
		return models.JobKindProjectDocumentUploadPrepare, nil
	case string(taskengine.JobKindProjectDocumentDelete):
		return models.JobKindProjectDocumentDelete, nil
	default:
		return "", fmt.Errorf("unsupported job kind %q", value)
	}
}

func toTrackerSourceKindString(kind models.TrackerSourceKind) string {
	switch kind {
	case models.TrackerSourceKindInternal:
		return "internal"
	default:
		return ""
	}
}

func toGraphTrackerSourceKind(kind string) (models.TrackerSourceKind, error) {
	switch strings.TrimSpace(kind) {
	case "internal":
		return models.TrackerSourceKindInternal, nil
	default:
		return "", fmt.Errorf("unsupported tracker provider %q", kind)
	}
}

func toSCMProviderString(provider models.SCMProvider) string {
	switch provider {
	case models.SCMProviderGithub:
		return "github"
	default:
		return ""
	}
}

func toGraphSCMProvider(provider string) (models.SCMProvider, error) {
	switch strings.TrimSpace(provider) {
	case "github":
		return models.SCMProviderGithub, nil
	default:
		return "", fmt.Errorf("unsupported scm provider %q", provider)
	}
}

func toGraphLifecycleTreeNodeType(value applicationcontrolplane.LifecycleTreeNodeType) (models.LifecycleTreeNodeType, error) {
	switch strings.TrimSpace(string(value)) {
	case string(applicationcontrolplane.LifecycleTreeNodeTypeRun):
		return models.LifecycleTreeNodeTypeRun, nil
	case string(applicationcontrolplane.LifecycleTreeNodeTypeTask):
		return models.LifecycleTreeNodeTypeTask, nil
	case string(applicationcontrolplane.LifecycleTreeNodeTypeJob):
		return models.LifecycleTreeNodeTypeJob, nil
	case string(applicationcontrolplane.LifecycleTreeNodeTypeSession):
		return models.LifecycleTreeNodeTypeSession, nil
	default:
		return "", fmt.Errorf("unsupported lifecycle tree node type %q", value)
	}
}

func toGraphProjectSetup(project *applicationcontrolplane.ProjectSetup) (*models.ProjectSetup, error) {
	if project == nil {
		return nil, fmt.Errorf("project setup is required")
	}
	repositories := make([]*models.ProjectRepository, 0, len(project.Repositories))
	for _, repository := range project.Repositories {
		repositories = append(repositories, &models.ProjectRepository{
			RepositoryID:  repository.RepositoryID,
			ScmID:         repository.SCMID,
			RepositoryURL: repository.RepositoryURL,
			IsPrimary:     repository.IsPrimary,
		})
	}
	scms := make([]*models.ProjectScm, 0, len(project.SCMs))
	for _, scm := range project.SCMs {
		scmProvider, scmErr := toGraphSCMProvider(scm.SCMProvider)
		if scmErr != nil {
			return nil, scmErr
		}
		scms = append(scms, &models.ProjectScm{ScmID: scm.SCMID, ScmProvider: scmProvider})
	}
	boards := make([]*models.ProjectBoard, 0, len(project.Boards))
	for _, board := range project.Boards {
		trackerProvider, trackerErr := toGraphTrackerSourceKind(board.TrackerProvider)
		if trackerErr != nil {
			return nil, trackerErr
		}
		repositoryIDs := make([]string, 0, len(board.RepositoryIDs))
		for _, repositoryID := range board.RepositoryIDs {
			repositoryIDs = append(repositoryIDs, strings.TrimSpace(repositoryID))
		}
		boards = append(boards, &models.ProjectBoard{
			BoardID:                  board.BoardID,
			TrackerProvider:          trackerProvider,
			TaskboardName:            nilIfEmpty(board.TaskboardName),
			AppliesToAllRepositories: board.AppliesToAllRepositories,
			RepositoryIDs:            repositoryIDs,
		})
	}
	return &models.ProjectSetup{
		ProjectID:    project.ProjectID,
		ProjectName:  project.ProjectName,
		Scms:         scms,
		Repositories: repositories,
		Boards:       boards,
		CreatedAt:    project.CreatedAt.UTC(),
		UpdatedAt:    project.UpdatedAt.UTC(),
	}, nil
}

func toGraphProjectDocument(document applicationcontrolplane.ProjectDocument) *models.ProjectDocument {
	return &models.ProjectDocument{
		ProjectID:   document.ProjectID,
		DocumentID:  document.DocumentID,
		FileName:    document.FileName,
		ContentType: document.ContentType,
		ObjectPath:  document.ObjectPath,
		CdnURL:      document.CDNURL,
		Status:      document.Status,
		CreatedAt:   document.CreatedAt.UTC(),
		UpdatedAt:   document.UpdatedAt.UTC(),
	}
}

func toGraphStreamEventSource(source domainstream.Source) models.StreamEventSource {
	switch source {
	case domainstream.SourceACP:
		return models.StreamEventSourceAcp
	case domainstream.SourceSessionFile:
		return models.StreamEventSourceSessionFile
	default:
		return models.StreamEventSourceWorker
	}
}

func mapRepositorySourceBranches(values []*models.RepositorySourceBranchInput) []applicationcontrolplane.RepositorySourceBranch {
	mapped := make([]applicationcontrolplane.RepositorySourceBranch, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		repositoryID := strings.TrimSpace(value.RepositoryID)
		branch := strings.TrimSpace(value.Branch)
		if repositoryID == "" || branch == "" {
			continue
		}
		mapped = append(mapped, applicationcontrolplane.RepositorySourceBranch{RepositoryID: repositoryID, Branch: branch})
	}
	return mapped
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func nilIfEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func strPtr(value string) *string {
	return &value
}
