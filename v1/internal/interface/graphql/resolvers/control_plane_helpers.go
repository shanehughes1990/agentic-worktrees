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

func streamSubscription(ctx context.Context, streamService *applicationstream.Service, correlation models.SupervisorCorrelationInput, fromOffset *int32, eventFilter func(domainstream.EventType) bool) (<-chan models.StreamEventResult, error) {
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
	go func() {
		defer cancel()
		defer close(output)

		for _, event := range replay {
			if !matchesCorrelation(event, correlation) || !eventFilter(event.EventType) {
				continue
			}
			select {
			case output <- models.StreamEventSuccess{Event: mapStreamEvent(event)}:
			case <-ctx.Done():
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, open := <-live:
				if !open {
					return
				}
				if !matchesCorrelation(event, correlation) || !eventFilter(event.EventType) {
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

func matchesCorrelation(event domainstream.Event, correlation models.SupervisorCorrelationInput) bool {
	switch event.EventType {
	case domainstream.EventWorkerHeartbeat, domainstream.EventWorkerShutdown, domainstream.EventWorkerDeregistered, domainstream.EventWorkerRogueDetected:
		return true
	}
	if strings.TrimSpace(correlation.RunID) != "" && strings.TrimSpace(event.CorrelationIDs.RunID) != strings.TrimSpace(correlation.RunID) {
		return false
	}
	if strings.TrimSpace(correlation.TaskID) != "" && strings.TrimSpace(event.CorrelationIDs.TaskID) != strings.TrimSpace(correlation.TaskID) {
		return false
	}
	if strings.TrimSpace(correlation.JobID) != "" && strings.TrimSpace(event.CorrelationIDs.JobID) != strings.TrimSpace(correlation.JobID) {
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
		Payload:       string(payload),
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

func toGraphJobKind(value taskengine.JobKind) (models.JobKind, error) {
	switch strings.TrimSpace(string(value)) {
	case string(taskengine.JobKindIngestionAgent):
		return models.JobKindIngestionAgentRun, nil
	case string(taskengine.JobKindAgentWorkflow):
		return models.JobKindAgentWorkflowRun, nil
	case string(taskengine.JobKindSCMWorkflow):
		return models.JobKindScmWorkflowRun, nil
	default:
		return "", fmt.Errorf("unsupported job kind %q", value)
	}
}

func toTrackerSourceKindString(kind models.TrackerSourceKind) string {
	switch kind {
	case models.TrackerSourceKindLocalJSON:
		return "local_json"
	case models.TrackerSourceKindGithubIssues:
		return "github_issues"
	case models.TrackerSourceKindJira:
		return "jira"
	case models.TrackerSourceKindLinear:
		return "linear"
	default:
		return ""
	}
}

func toGraphTrackerSourceKind(kind string) (models.TrackerSourceKind, error) {
	switch strings.TrimSpace(kind) {
	case "local_json":
		return models.TrackerSourceKindLocalJSON, nil
	case "github_issues":
		return models.TrackerSourceKindGithubIssues, nil
	case "jira":
		return models.TrackerSourceKindJira, nil
	case "linear":
		return models.TrackerSourceKindLinear, nil
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

func toGraphProjectSetup(project *applicationcontrolplane.ProjectSetup) (*models.ProjectSetup, error) {
	if project == nil {
		return nil, fmt.Errorf("project setup is required")
	}
	scmProvider, scmErr := toGraphSCMProvider(project.SCMProvider)
	if scmErr != nil {
		return nil, scmErr
	}
	trackerProvider, trackerErr := toGraphTrackerSourceKind(project.TrackerProvider)
	if trackerErr != nil {
		return nil, trackerErr
	}
	return &models.ProjectSetup{
		ProjectID:       project.ProjectID,
		ProjectName:     project.ProjectName,
		ScmProvider:     scmProvider,
		RepositoryURL:   project.RepositoryURL,
		TrackerProvider: trackerProvider,
		TrackerLocation: nilIfEmpty(project.TrackerLocation),
		TrackerBoardID:  nilIfEmpty(project.TrackerBoardID),
		CreatedAt:       project.CreatedAt.UTC(),
		UpdatedAt:       project.UpdatedAt.UTC(),
	}, nil
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
