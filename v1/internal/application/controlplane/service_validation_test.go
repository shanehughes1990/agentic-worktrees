package controlplane

import "testing"

func validProjectSetupRequest() UpsertProjectSetupRequest {
	return UpsertProjectSetupRequest{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		Repositories: []ProjectRepository{
			{
				RepositoryID:  "repo-1",
				SCMProvider:   "github",
				RepositoryURL: "https://github.com/acme/repo",
				IsPrimary:     true,
			},
		},
		Boards: []ProjectBoard{
			{
				BoardID:                  "board-1",
				TrackerProvider:          "github_issues",
				TrackerLocation:          "acme/repo",
				TrackerBoardID:           "",
				AppliesToAllRepositories: true,
				RepositoryIDs:            nil,
			},
		},
	}
}

func TestUpsertProjectSetupValidateRequiresExactlyOneBoard(t *testing.T) {
	request := validProjectSetupRequest()
	request.Boards = append(request.Boards, ProjectBoard{
		BoardID:                  "board-2",
		TrackerProvider:          "local_json",
		TrackerLocation:          "taskboard.json",
		AppliesToAllRepositories: true,
	})

	err := request.Validate()
	if err == nil || err.Error() != "exactly one board is required" {
		t.Fatalf("expected exact board-count validation error, got %v", err)
	}
}

func TestUpsertProjectSetupValidateRejectsUnsupportedTrackerProvider(t *testing.T) {
	request := validProjectSetupRequest()
	request.Boards[0].TrackerProvider = "jira"

	err := request.Validate()
	if err == nil || err.Error() != "boards[0].tracker_provider must be one of: local_json, github_issues" {
		t.Fatalf("expected tracker-provider validation error, got %v", err)
	}
}

func TestUpsertProjectSetupValidateRejectsRepositoryScopedBoard(t *testing.T) {
	request := validProjectSetupRequest()
	request.Boards[0].AppliesToAllRepositories = false
	request.Boards[0].RepositoryIDs = []string{"repo-1"}

	err := request.Validate()
	if err == nil || err.Error() != "boards[0].applies_to_all_repositories must be true" {
		t.Fatalf("expected project-scoped board validation error, got %v", err)
	}
}

func validIngestionRequest() EnqueueIngestionWorkflowRequest {
	return EnqueueIngestionWorkflowRequest{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "idem-1",
		Prompt:         "ingest",
		ProjectID:      "project-1",
		WorkflowID:     "workflow-1",
		BoardSources: []IngestionBoardSource{
			{
				BoardID:                  "board-1",
				Kind:                     "github_issues",
				Location:                 "acme/repo",
				AppliesToAllRepositories: true,
			},
		},
	}
}

func TestEnqueueIngestionWorkflowValidateRequiresExactlyOneBoardSource(t *testing.T) {
	request := validIngestionRequest()
	request.BoardSources = append(request.BoardSources, IngestionBoardSource{BoardID: "board-2", Kind: "local_json", Location: "taskboard.json", AppliesToAllRepositories: true})

	err := request.Validate()
	if err == nil || err.Error() != "exactly one board_source is required" {
		t.Fatalf("expected exact board_source-count validation error, got %v", err)
	}
}

func TestEnqueueIngestionWorkflowValidateRejectsUnsupportedKind(t *testing.T) {
	request := validIngestionRequest()
	request.BoardSources[0].Kind = "jira"

	err := request.Validate()
	if err == nil {
		t.Fatalf("expected source-kind validation error")
	}
}
