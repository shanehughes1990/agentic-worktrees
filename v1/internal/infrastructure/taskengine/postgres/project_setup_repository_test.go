package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"encoding/json"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newProjectSetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func sampleProjectSetup() applicationcontrolplane.ProjectSetup {
	return applicationcontrolplane.ProjectSetup{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		Repositories: []applicationcontrolplane.ProjectRepository{{
			RepositoryID: "repo-1",
			SCMProvider:  "github",
			RepositoryURL: "https://github.com/octo/repo",
			IsPrimary:    true,
		}},
		Boards: []applicationcontrolplane.ProjectBoard{{
			BoardID:                  "project_1_board",
			TrackerProvider:          "internal",
			TaskboardName:            "Project 1 Board",
			AppliesToAllRepositories: true,
		}},
	}
}

func TestProjectSetupRepositorySeedsInitialInternalSnapshot(t *testing.T) {
	db := newProjectSetupTestDB(t)
	repo, err := NewProjectSetupRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	if _, err := repo.UpsertProjectSetup(context.Background(), sampleProjectSetup()); err != nil {
		t.Fatalf("upsert setup: %v", err)
	}

	var snapshot trackerBoardSnapshotRecord
	if err := db.Where("run_id = ? AND board_id = ?", "project-1", "project_1_board").Take(&snapshot).Error; err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(snapshot.Payload, &payload); err != nil {
		t.Fatalf("decode snapshot payload: %v", err)
	}
	if payload["status"] != "not-started" {
		t.Fatalf("expected seeded status not-started, got %v", payload["status"])
	}
}

func TestProjectSetupRepositoryDoesNotOverwriteExistingInternalSnapshot(t *testing.T) {
	db := newProjectSetupTestDB(t)
	repo, err := NewProjectSetupRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	if _, err := repo.UpsertProjectSetup(context.Background(), sampleProjectSetup()); err != nil {
		t.Fatalf("upsert setup: %v", err)
	}

	persistedBoard := map[string]any{
		"board_id": "project_1_board",
		"run_id":   "project-1",
		"title":    "Project 1 Board",
		"status":   "in-progress",
		"epics": []map[string]any{{
			"id":       "epic-1",
			"board_id": "project_1_board",
			"title":    "Existing Epic",
			"status":   "in-progress",
			"tasks": []map[string]any{{
				"id":       "task-1",
				"board_id": "project_1_board",
				"title":    "Existing Task",
				"status":   "in-progress",
			}},
		}},
	}
	persistedPayload, err := json.Marshal(persistedBoard)
	if err != nil {
		t.Fatalf("encode persisted payload: %v", err)
	}
	if err := db.Model(&trackerBoardSnapshotRecord{}).
		Where("run_id = ? AND board_id = ?", "project-1", "project_1_board").
		Update("payload", persistedPayload).Error; err != nil {
		t.Fatalf("seed persisted payload: %v", err)
	}

	if _, err := repo.UpsertProjectSetup(context.Background(), sampleProjectSetup()); err != nil {
		t.Fatalf("second upsert setup: %v", err)
	}

	var snapshot trackerBoardSnapshotRecord
	if err := db.Where("run_id = ? AND board_id = ?", "project-1", "project_1_board").Take(&snapshot).Error; err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(snapshot.Payload, &payload); err != nil {
		t.Fatalf("decode snapshot payload: %v", err)
	}
	if payload["status"] != "in-progress" {
		t.Fatalf("expected existing board payload to be preserved, got status %v", payload["status"])
	}
}
