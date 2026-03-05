package realtime

import (
	"testing"
	"time"
)

func TestTableChangeEventValidateRequiresFields(t *testing.T) {
	now := time.Now().UTC()
	event := TableChangeEvent{
		Topic:           "lifecycle_project_session_history",
		Table:           "project_session_history",
		Operation:       "insert",
		ProjectID:       "project-1",
		ProjectEventSeq: 1,
		SessionEventSeq: 1,
		OccurredAt:      now,
	}
	if err := event.Validate(); err != nil {
		t.Fatalf("expected valid table change event, got %v", err)
	}

	event.Topic = ""
	if err := event.Validate(); err == nil {
		t.Fatalf("expected topic validation error")
	}
}
