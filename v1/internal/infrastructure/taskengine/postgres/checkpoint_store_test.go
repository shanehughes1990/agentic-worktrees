package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func TestPostgresCheckpointStoreSaveAndLoad(t *testing.T) {
	store, err := NewPostgresCheckpointStore(newTestDB(t))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := store.Save(context.Background(), "id-1", taskengine.Checkpoint{Step: "source_state", Token: "id-1"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	checkpoint, err := store.Load(context.Background(), "id-1")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if checkpoint == nil || checkpoint.Step != "source_state" || checkpoint.Token != "id-1" {
		t.Fatalf("unexpected checkpoint: %+v", checkpoint)
	}
}

func TestPostgresCheckpointStoreLoadMissReturnsNil(t *testing.T) {
	store, err := NewPostgresCheckpointStore(newTestDB(t))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	checkpoint, err := store.Load(context.Background(), "missing")
	if err != nil {
		t.Fatalf("load miss: %v", err)
	}
	if checkpoint != nil {
		t.Fatalf("expected nil checkpoint, got %+v", checkpoint)
	}
}
