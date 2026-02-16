package runstate

import (
	"path/filepath"
	"testing"
)

func TestStoreUpsertAndSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoints.json")
	store := NewStore(path)

	if err := store.Upsert(Checkpoint{RunID: "run-1", TaskID: "task-1", Phase: "prepare", Status: "running"}); err != nil {
		t.Fatalf("upsert running: %v", err)
	}
	if err := store.Upsert(Checkpoint{RunID: "run-1", TaskID: "task-1", Phase: "cleanup", Status: "completed"}); err != nil {
		t.Fatalf("upsert completed: %v", err)
	}

	rows, err := store.All()
	if err != nil {
		t.Fatalf("all error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	if rows[0].Phase != "cleanup" {
		t.Fatalf("expected last phase cleanup, got %s", rows[0].Phase)
	}

	summary, err := store.Summary()
	if err != nil {
		t.Fatalf("summary error: %v", err)
	}
	if summary["completed"] != 1 {
		t.Fatalf("expected completed=1, got %+v", summary)
	}
}
