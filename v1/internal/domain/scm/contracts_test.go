package scm

import (
	"agentic-orchestrator/internal/domain/failures"
	"testing"
)

func TestRepositoryValidateRequiresFields(t *testing.T) {
	err := (Repository{}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestWorktreeSpecValidateRequiresFields(t *testing.T) {
	err := (WorktreeSpec{BaseBranch: "main", TargetBranch: "feature/one"}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestMergeReadinessValidateRequiresReasonWhenNotMergeable(t *testing.T) {
	err := (MergeReadiness{CanMerge: false}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}
