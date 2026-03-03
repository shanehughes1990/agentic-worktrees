package asynq

import (
	"context"
	"testing"
)

func TestDeadLetterManagerListValidatesQueue(t *testing.T) {
	platform := &APIPlatform{}
	if _, err := platform.ListDeadLetters(context.Background(), "", 10); err == nil {
		t.Fatalf("expected queue validation error")
	}
}

func TestDeadLetterManagerRequeueValidatesInputs(t *testing.T) {
	platform := &APIPlatform{}
	if err := platform.RequeueDeadLetter(context.Background(), "", "task-1"); err == nil {
		t.Fatalf("expected queue validation error")
	}
	if err := platform.RequeueDeadLetter(context.Background(), "scm", ""); err == nil {
		t.Fatalf("expected task id validation error")
	}
}

func TestDeadLetterManagerNilPlatformReturnsError(t *testing.T) {
	var platform *APIPlatform
	if _, err := platform.ListDeadLetters(context.Background(), "scm", 10); err == nil {
		t.Fatalf("expected nil platform error")
	}
	if err := platform.RequeueDeadLetter(context.Background(), "scm", "task-1"); err == nil {
		t.Fatalf("expected nil platform error")
	}
}
