package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	checkpointdomain "github.com/shanehughes1990/agentic-worktrees/internal/features/checkpoints/domain"
)

func TestAppendCreatesFileAndPersistsRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoints.json")
	store := NewJSONStore(path)

	record := checkpointdomain.Record{
		RunID:  "run-1",
		TaskID: "task-1",
		JobID:  "job-1",
		State:  "queued",
	}
	if err := store.Append(record); err != nil {
		t.Fatalf("append record: %v", err)
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read checkpoint file: %v", err)
	}

	var records []checkpointdomain.Record
	if err := json.Unmarshal(payload, &records); err != nil {
		t.Fatalf("unmarshal records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Timestamp.IsZero() {
		t.Fatalf("expected non-zero timestamp")
	}
}

func TestAppendNormalizesTimestampToUTC(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoints.json")
	store := NewJSONStore(path)

	nonUTC := time.Date(2026, 2, 16, 12, 0, 0, 0, time.FixedZone("Offset", 2*3600))
	record := checkpointdomain.Record{
		RunID:     "run-1",
		TaskID:    "task-1",
		JobID:     "job-1",
		State:     "executing",
		Timestamp: nonUTC,
	}
	if err := store.Append(record); err != nil {
		t.Fatalf("append record: %v", err)
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read checkpoint file: %v", err)
	}

	var records []checkpointdomain.Record
	if err := json.Unmarshal(payload, &records); err != nil {
		t.Fatalf("unmarshal records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Timestamp.Location() != time.UTC {
		t.Fatalf("expected UTC timestamp, got %s", records[0].Timestamp.Location())
	}
}
