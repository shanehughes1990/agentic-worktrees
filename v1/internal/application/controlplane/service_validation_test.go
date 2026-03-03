package controlplane

import "testing"

func validProjectSetupRequest() UpsertProjectSetupRequest {
	return UpsertProjectSetupRequest{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		SCMs: []ProjectSCM{
			{
				SCMID:       "scm-1",
				SCMProvider: "github",
				SCMToken:    "token",
			},
		},
		Repositories: []ProjectRepository{
			{
				RepositoryID:  "repo-1",
				SCMID:         "scm-1",
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

func TestUpsertProjectSetupValidateRequiresRepositorySCMReference(t *testing.T) {
	request := validProjectSetupRequest()
	request.Repositories[0].SCMID = ""

	err := request.Validate()
	if err == nil || err.Error() != "repositories[0].scm_id is required" {
		t.Fatalf("expected repository scm_id validation error, got %v", err)
	}
}

func TestUpsertProjectSetupValidateAllowsZeroBoards(t *testing.T) {
	request := validProjectSetupRequest()
	request.Boards = nil

	err := request.Validate()
	if err != nil {
		t.Fatalf("expected no board-count validation error, got %v", err)
	}
}

func TestUpsertProjectSetupValidateRejectsMoreThanOneBoard(t *testing.T) {
	request := validProjectSetupRequest()
	request.Boards = append(request.Boards, ProjectBoard{
		BoardID:                  "board-2",
		TrackerProvider:          "internal",
		TaskboardName:            "Main Board",
		AppliesToAllRepositories: true,
	})

	err := request.Validate()
	if err == nil || err.Error() != "at most one board is supported" {
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

func TestUpsertProjectSetupValidateAllowsBlankSCMToken(t *testing.T) {
	request := validProjectSetupRequest()
	request.SCMs[0].SCMToken = ""

	err := request.Validate()
	if err != nil {
		t.Fatalf("expected blank scm token to pass request validation, got %v", err)
	}
}
