package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newProjectSetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func newProjectSetupTestCrypto(t *testing.T, db *gorm.DB) *SCMTokenCrypto {
	t.Helper()
	crypto, err := NewSCMTokenCrypto(db)
	if err != nil {
		t.Fatalf("new scm token crypto: %v", err)
	}
	return crypto
}

func sampleProjectSetup() applicationcontrolplane.ProjectSetup {
	return applicationcontrolplane.ProjectSetup{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		SCMs: []applicationcontrolplane.ProjectSCM{{
			SCMID:       "scm-1",
			SCMProvider: "github",
			SCMToken:    "token",
		}},
		Repositories: []applicationcontrolplane.ProjectRepository{{
			RepositoryID:  "repo-1",
			SCMID:         "scm-1",
			RepositoryURL: "https://github.com/octo/repo",
			IsPrimary:     true,
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
	crypto := newProjectSetupTestCrypto(t, db)
	repo, err := NewProjectSetupRepository(db, crypto)
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
	crypto := newProjectSetupTestCrypto(t, db)
	repo, err := NewProjectSetupRepository(db, crypto)
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

func TestProjectSetupRepositoryEncryptsSCMTokenAtRest(t *testing.T) {
	db := newProjectSetupTestDB(t)
	crypto := newProjectSetupTestCrypto(t, db)
	repo, err := NewProjectSetupRepository(db, crypto)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	setup := sampleProjectSetup()
	setup.SCMs[0].SCMToken = "ghp_super_secret_token"
	if _, err := repo.UpsertProjectSetup(context.Background(), setup); err != nil {
		t.Fatalf("upsert setup: %v", err)
	}

	var storedSCM projectSCMRecord
	if err := db.Where("project_id = ? AND scm_id = ?", "project-1", "scm-1").Take(&storedSCM).Error; err != nil {
		t.Fatalf("load stored scm row: %v", err)
	}
	if storedSCM.SCMToken == "ghp_super_secret_token" {
		t.Fatalf("expected scm token to be encrypted at rest")
	}

	loadedSetup, err := repo.GetProjectSetup(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("load project setup: %v", err)
	}
	if loadedSetup == nil || len(loadedSetup.SCMs) == 0 {
		t.Fatalf("expected loaded project setup scm entry")
	}
	if loadedSetup.SCMs[0].SCMToken != "ghp_super_secret_token" {
		t.Fatalf("expected decrypted scm token for internal use, got %q", loadedSetup.SCMs[0].SCMToken)
	}
}

func TestProjectSetupRepositoryWithoutBoardsDoesNotSeedSnapshot(t *testing.T) {
	db := newProjectSetupTestDB(t)
	crypto := newProjectSetupTestCrypto(t, db)
	repo, err := NewProjectSetupRepository(db, crypto)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	setup := sampleProjectSetup()
	setup.Boards = nil
	if _, err := repo.UpsertProjectSetup(context.Background(), setup); err != nil {
		t.Fatalf("upsert setup: %v", err)
	}

	var snapshotCount int64
	if err := db.Model(&trackerBoardSnapshotRecord{}).
		Where("run_id = ?", "project-1").
		Count(&snapshotCount).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	if snapshotCount != 0 {
		t.Fatalf("expected no tracker snapshots for setup without boards, got %d", snapshotCount)
	}
}

func TestProjectSetupRepositoryPreservesStoredSCMTokenWhenBlankOnUpdate(t *testing.T) {
	db := newProjectSetupTestDB(t)
	crypto := newProjectSetupTestCrypto(t, db)
	repo, err := NewProjectSetupRepository(db, crypto)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	setup := sampleProjectSetup()
	setup.SCMs[0].SCMToken = "ghp_initial_token"
	if _, err := repo.UpsertProjectSetup(context.Background(), setup); err != nil {
		t.Fatalf("initial upsert setup: %v", err)
	}

	update := sampleProjectSetup()
	update.ProjectName = "Project One Updated"
	update.SCMs[0].SCMToken = ""
	if _, err := repo.UpsertProjectSetup(context.Background(), update); err != nil {
		t.Fatalf("update setup with blank token: %v", err)
	}

	loadedSetup, err := repo.GetProjectSetup(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("load project setup: %v", err)
	}
	if loadedSetup == nil || len(loadedSetup.SCMs) == 0 {
		t.Fatalf("expected loaded project setup scm entry")
	}
	if loadedSetup.SCMs[0].SCMToken != "ghp_initial_token" {
		t.Fatalf("expected preserved scm token, got %q", loadedSetup.SCMs[0].SCMToken)
	}
}
