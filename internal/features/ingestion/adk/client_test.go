package adk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	boarddomain "github.com/shanehughes1990/agentic-worktrees/internal/features/board/domain"
	"github.com/shanehughes1990/agentic-worktrees/internal/features/ingestion/pipeline"
	sharederrors "github.com/shanehughes1990/agentic-worktrees/internal/shared/errors"
)

func TestPlanBoardSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"schema_version":1,"source_scope":"docs","generated_at":"2026-02-16T00:00:00Z","epics":[{"id":"epic-001","title":"e","dependencies":[],"tasks":[{"id":"task-1","title":"t","dependencies":[],"lane":"lane-a","status":"pending"}]}]}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "", time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	board, err := client.PlanBoard(context.Background(), "docs", []pipeline.ScopeFile{{Path: "01.md", Content: "# scope"}})
	if err != nil {
		t.Fatalf("plan board: %v", err)
	}
	if board.SchemaVersion != 1 || len(board.Epics) != 1 {
		t.Fatalf("unexpected board response: %+v", board)
	}
}

func TestPlanBoardClassifiesTerminal4xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "", time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.PlanBoard(context.Background(), "docs", []pipeline.ScopeFile{{Path: "01.md", Content: "# scope"}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if sharederrors.ClassOf(err) != sharederrors.ClassTerminal {
		t.Fatalf("expected terminal class, got %s", sharederrors.ClassOf(err))
	}
}

func TestPlanBoardClassifiesTransient5xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal", http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "", time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.PlanBoard(context.Background(), "docs", []pipeline.ScopeFile{{Path: "01.md", Content: "# scope"}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if sharederrors.ClassOf(err) != sharederrors.ClassTransient {
		t.Fatalf("expected transient class, got %s", sharederrors.ClassOf(err))
	}
}

func TestNewClientValidatesURL(t *testing.T) {
	if _, err := NewClient("", "", time.Second); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestBoardValidationFailureIsTerminal(t *testing.T) {
	invalidBoard := boarddomain.Board{SchemaVersion: 1, SourceScope: "docs"}
	_ = invalidBoard
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"schema_version":1,"source_scope":"docs","generated_at":"2026-02-16T00:00:00Z","epics":[]}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "", time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.PlanBoard(context.Background(), "docs", []pipeline.ScopeFile{{Path: "01.md", Content: "# scope"}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if sharederrors.ClassOf(err) != sharederrors.ClassTerminal {
		t.Fatalf("expected terminal class, got %s", sharederrors.ClassOf(err))
	}
}
