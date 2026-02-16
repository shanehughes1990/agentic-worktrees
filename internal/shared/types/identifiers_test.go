package types

import "testing"

func TestValidateNonEmpty(t *testing.T) {
	if err := ValidateNonEmpty("task_id", TaskID("task-123")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := ValidateNonEmpty("task_id", TaskID("  ")); err == nil {
		t.Fatalf("expected error for empty value")
	}
}

func TestIdentifierStringers(t *testing.T) {
	if RunID("r1").String() != "r1" {
		t.Fatalf("unexpected run id string")
	}
	if TaskID("t1").String() != "t1" {
		t.Fatalf("unexpected task id string")
	}
	if JobID("j1").String() != "j1" {
		t.Fatalf("unexpected job id string")
	}
}

func TestCorrelationValidate(t *testing.T) {
	corr := Correlation{RunID: "r1", TaskID: "t1", JobID: "j1"}
	if err := corr.Validate(); err != nil {
		t.Fatalf("expected valid correlation, got %v", err)
	}

	invalid := Correlation{RunID: "", TaskID: "t1", JobID: "j1"}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}
