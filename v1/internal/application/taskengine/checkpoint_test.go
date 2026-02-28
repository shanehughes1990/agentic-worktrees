package taskengine

import "testing"

func TestCheckpointMatches(t *testing.T) {
	checkpoint := &Checkpoint{Step: "source_state", Token: "id-1"}
	if !CheckpointMatches(checkpoint, "source_state", "id-1") {
		t.Fatalf("expected checkpoint to match")
	}
}

func TestCheckpointMatchesRejectsTokenMismatch(t *testing.T) {
	checkpoint := &Checkpoint{Step: "source_state", Token: "id-2"}
	if CheckpointMatches(checkpoint, "source_state", "id-1") {
		t.Fatalf("expected checkpoint token mismatch")
	}
}

func TestCheckpointMatchesRejectsStepMismatch(t *testing.T) {
	checkpoint := &Checkpoint{Step: "ensure_worktree", Token: "id-1"}
	if CheckpointMatches(checkpoint, "source_state", "id-1") {
		t.Fatalf("expected checkpoint step mismatch")
	}
}
