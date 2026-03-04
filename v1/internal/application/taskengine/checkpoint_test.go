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
	checkpoint := &Checkpoint{Step: "ensure_repository", Token: "id-1"}
	if CheckpointMatches(checkpoint, "source_state", "id-1") {
		t.Fatalf("expected checkpoint step mismatch")
	}
}

func TestCheckpointMatchesRejectsEmptyExpectedValues(t *testing.T) {
	checkpoint := &Checkpoint{Step: "source_state", Token: "id-1"}
	if CheckpointMatches(checkpoint, "", "id-1") {
		t.Fatalf("expected empty step to be rejected")
	}
	if CheckpointMatches(checkpoint, "source_state", "") {
		t.Fatalf("expected empty token to be rejected")
	}
}

func TestRetryCheckpointContractPrefersResumeCheckpoint(t *testing.T) {
	contract := RetryCheckpointContract{
		ResumeCheckpoint:    &Checkpoint{Step: "source_state", Token: "id-1"},
		CompletedCheckpoint: &Checkpoint{Step: "ensure_repository", Token: "id-1"},
	}
	if !contract.Matches("source_state", "id-1") {
		t.Fatalf("expected resume checkpoint to take precedence")
	}
}

func TestRetryCheckpointContractFallsBackToCompletedCheckpoint(t *testing.T) {
	contract := RetryCheckpointContract{
		CompletedCheckpoint: &Checkpoint{Step: "source_state", Token: "id-1"},
	}
	if !contract.Matches("source_state", "id-1") {
		t.Fatalf("expected completed checkpoint fallback to match")
	}
}

func TestRetryCheckpointContractSupportsLegacyFlattenedFields(t *testing.T) {
	contract := RetryCheckpointContract{
		ResumeCheckpointStep:  " source_state ",
		ResumeCheckpointToken: " id-1 ",
	}
	checkpoint := contract.Checkpoint()
	if checkpoint == nil {
		t.Fatalf("expected checkpoint from legacy fields")
	}
	if checkpoint.Step != "source_state" || checkpoint.Token != "id-1" {
		t.Fatalf("expected trimmed legacy checkpoint, got %+v", checkpoint)
	}
}
