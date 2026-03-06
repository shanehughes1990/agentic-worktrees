package controlplane

import (
	"context"
	"testing"
)

type projectArtifactsQueryRepoStub struct{}

func (repository *projectArtifactsQueryRepoStub) ListSessions(ctx context.Context, limit int) ([]SessionSummary, error) {
	return nil, nil
}

func (repository *projectArtifactsQueryRepoStub) GetSession(ctx context.Context, runID string) (*SessionSummary, error) {
	return nil, nil
}

func (repository *projectArtifactsQueryRepoStub) ListWorkflowJobs(ctx context.Context, runID string, taskID string, limit int) ([]WorkflowJob, error) {
	return nil, nil
}

func (repository *projectArtifactsQueryRepoStub) ListExecutionHistory(ctx context.Context, filter CorrelationFilter, limit int) ([]ExecutionHistoryRecord, error) {
	return nil, nil
}

func (repository *projectArtifactsQueryRepoStub) ListDeadLetterHistory(ctx context.Context, queue string, limit int) ([]DeadLetterHistoryRecord, error) {
	return nil, nil
}

func (repository *projectArtifactsQueryRepoStub) ListLifecycleSessionSnapshots(ctx context.Context, projectID string, pipelineType string, limit int) ([]LifecycleSessionSnapshot, error) {
	return nil, nil
}

func (repository *projectArtifactsQueryRepoStub) ListLifecycleSessionHistory(ctx context.Context, projectID string, sessionID string, fromEventSeq int64, limit int) ([]LifecycleHistoryEvent, error) {
	return nil, nil
}

func (repository *projectArtifactsQueryRepoStub) ListLifecycleTreeNodes(ctx context.Context, filter LifecycleTreeFilter, limit int) ([]LifecycleTreeNode, error) {
	return nil, nil
}

type projectArtifactsRepoStub struct {
	setup        *ProjectSetup
	upserted     *ProjectSetup
	deleteCalled bool
}

func (repository *projectArtifactsRepoStub) ListProjectSetups(ctx context.Context, limit int) ([]ProjectSetup, error) {
	if repository.setup == nil {
		return nil, nil
	}
	return []ProjectSetup{*repository.setup}, nil
}

func (repository *projectArtifactsRepoStub) GetProjectSetup(ctx context.Context, projectID string) (*ProjectSetup, error) {
	if repository.setup == nil {
		return nil, nil
	}
	copyValue := *repository.setup
	return &copyValue, nil
}

func (repository *projectArtifactsRepoStub) UpsertProjectSetup(ctx context.Context, setup ProjectSetup) (*ProjectSetup, error) {
	copyValue := setup
	repository.upserted = &copyValue
	repository.setup = &copyValue
	return &copyValue, nil
}

func (repository *projectArtifactsRepoStub) DeleteProjectSetup(ctx context.Context, projectID string) error {
	repository.deleteCalled = true
	repository.setup = nil
	return nil
}

type captureProjectRepositoryArtifactManager struct {
	called   int
	previous *ProjectSetup
	current  ProjectSetup
}

func (manager *captureProjectRepositoryArtifactManager) ReconcileProjectRepositories(ctx context.Context, previous *ProjectSetup, current ProjectSetup) error {
	manager.called++
	if previous != nil {
		copyPrevious := *previous
		manager.previous = &copyPrevious
	} else {
		manager.previous = nil
	}
	manager.current = current
	return nil
}

func TestUpsertProjectSetupReconcilesRepositoryArtifacts(t *testing.T) {
	repository := &projectArtifactsRepoStub{setup: &ProjectSetup{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		SCMs: []ProjectSCM{{
			SCMID:       "scm-1",
			SCMProvider: "github",
			SCMToken:    "token",
		}},
		Repositories: []ProjectRepository{{
			RepositoryID:  "repo-old",
			SCMID:         "scm-1",
			RepositoryURL: "https://github.com/acme/repo-old",
			IsPrimary:     true,
		}},
	}}
	service, err := NewService(nil, &projectArtifactsQueryRepoStub{}, repository, nil)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	artifactManager := &captureProjectRepositoryArtifactManager{}
	service.SetProjectRepositoryArtifactManager(artifactManager)

	_, err = service.UpsertProjectSetup(context.Background(), UpsertProjectSetupRequest{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		SCMs: []ProjectSCM{{
			SCMID:       "scm-1",
			SCMProvider: "github",
			SCMToken:    "token",
		}},
		Repositories: []ProjectRepository{{
			RepositoryID:  "repo-new",
			SCMID:         "scm-1",
			RepositoryURL: "https://github.com/acme/repo-new",
			IsPrimary:     true,
		}},
	})
	if err != nil {
		t.Fatalf("upsert project setup: %v", err)
	}
	if artifactManager.called != 1 {
		t.Fatalf("expected one artifact reconcile call, got %d", artifactManager.called)
	}
	if artifactManager.previous == nil || artifactManager.previous.ProjectID != "project-1" {
		t.Fatalf("expected previous setup snapshot for project-1, got %+v", artifactManager.previous)
	}
	if len(artifactManager.current.Repositories) != 1 || artifactManager.current.Repositories[0].RepositoryID != "repo-new" {
		t.Fatalf("expected current repositories to include repo-new, got %+v", artifactManager.current.Repositories)
	}
}

func TestDeleteProjectSetupReconcilesRepositoryArtifacts(t *testing.T) {
	repository := &projectArtifactsRepoStub{setup: &ProjectSetup{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		SCMs: []ProjectSCM{{
			SCMID:       "scm-1",
			SCMProvider: "github",
			SCMToken:    "token",
		}},
		Repositories: []ProjectRepository{{
			RepositoryID:  "repo-1",
			SCMID:         "scm-1",
			RepositoryURL: "https://github.com/acme/repo-1",
			IsPrimary:     true,
		}},
	}}
	service, err := NewService(nil, &projectArtifactsQueryRepoStub{}, repository, nil)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	artifactManager := &captureProjectRepositoryArtifactManager{}
	service.SetProjectRepositoryArtifactManager(artifactManager)

	if err := service.DeleteProjectSetup(context.Background(), "project-1"); err != nil {
		t.Fatalf("delete project setup: %v", err)
	}
	if artifactManager.called != 1 {
		t.Fatalf("expected one artifact reconcile call, got %d", artifactManager.called)
	}
	if artifactManager.previous == nil || artifactManager.previous.ProjectID != "project-1" {
		t.Fatalf("expected previous setup snapshot for project-1, got %+v", artifactManager.previous)
	}
	if artifactManager.current.ProjectID != "project-1" {
		t.Fatalf("expected current reconcile setup scoped to project-1, got %+v", artifactManager.current)
	}
	if len(artifactManager.current.Repositories) != 0 {
		t.Fatalf("expected empty repositories in delete reconcile, got %+v", artifactManager.current.Repositories)
	}
}
