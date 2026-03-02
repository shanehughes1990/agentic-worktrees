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
				SCMToken:      "token",
				RepositoryURL: "https://github.com/acme/repo",
				IsPrimary:     true,
			},
		},
		Boards: []ProjectBoard{
			{
				BoardID:                  "board-1",
				TrackerProvider:          "internal",
				TaskboardName:            "Acme Repo Board",
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
		TrackerProvider:          "internal",
		TaskboardName:            "Main Board",
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
	if err == nil || err.Error() != "boards[0].tracker_provider must be internal" {
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
