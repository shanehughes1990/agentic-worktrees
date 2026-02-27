package dashboard

import (
	"context"
	"reflect"
	"testing"
	"time"

	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
)

func TestNewUI(t *testing.T) {
	ui := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, ".", "revamp", "redis://localhost:6379/0", 3)
	if ui == nil {
		t.Fatalf("expected ui instance")
	}
	ui.Stop()
}

func TestRunIngestionExecutePreservesDashboardFlow(t *testing.T) {
	type ingestCall struct {
		request  apptaskboard.IngestRequest
		redisURI string
	}
	ingestCalls := make(chan ingestCall, 1)

	ui := New(
		func(_ context.Context, request apptaskboard.IngestRequest, redisURI string) (apptaskboard.IngestionResult, error) {
			ingestCalls <- ingestCall{request: request, redisURI: redisURI}
			return apptaskboard.IngestionResult{BoardID: "board-1", RunID: "run-1"}, nil
		},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		".", "revamp", "redis://localhost:6379/0", 3,
	)
	defer ui.Stop()

	ui.runIngestionInput.SetText("/tmp/scope")
	ui.runIngestionDepth.SetText("2")
	ui.runIngestionIgnorePaths.SetText("vendor, .git")
	ui.runIngestionIgnoreExtensions.SetText(".tmp, md")

	execute := ui.runIngestionCommands.GetItemSelectedFunc(0)
	if execute == nil {
		t.Fatalf("expected execute command")
	}
	execute()

	select {
	case call := <-ingestCalls:
		if call.request.SourcePath != "/tmp/scope" {
			t.Fatalf("expected source path to be forwarded, got %q", call.request.SourcePath)
		}
		if call.request.SourceType != apptaskboard.IngestionSourceTypeFolder {
			t.Fatalf("expected folder source type, got %q", call.request.SourceType)
		}
		if call.request.Folder.WalkDepth != 2 {
			t.Fatalf("expected walk depth to be forwarded, got %d", call.request.Folder.WalkDepth)
		}
		if !reflect.DeepEqual(call.request.Folder.IgnorePaths, []string{"vendor", ".git"}) {
			t.Fatalf("expected ignore paths to be forwarded, got %#v", call.request.Folder.IgnorePaths)
		}
		if !reflect.DeepEqual(call.request.Folder.IgnoreExtensions, []string{".tmp", "md"}) {
			t.Fatalf("expected ignore extensions to be forwarded, got %#v", call.request.Folder.IgnoreExtensions)
		}
		if call.redisURI != "redis://localhost:6379/0" {
			t.Fatalf("expected default redis URI to be used, got %q", call.redisURI)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected ingest command to be invoked")
	}
}

func TestRunWorktreeExecuteStartsWhenNoReadyTasks(t *testing.T) {
	type startCall struct {
		boardID      string
		sourceBranch string
		maxTasks     int
		redisURI     string
	}

	startCalls := make(chan startCall, 1)

	ui := New(
		nil,
		func(_ context.Context, boardID string, sourceBranch string, maxTasks int, redisURI string) (string, error) {
			startCalls <- startCall{boardID: boardID, sourceBranch: sourceBranch, maxTasks: maxTasks, redisURI: redisURI}
			return "queue-task-1", nil
		},
		nil,
		func(_ context.Context) ([]string, error) { return []string{"board-1"}, nil },
		func(_ context.Context, boardID string) ([]string, error) {
			if boardID != "board-1" {
				t.Fatalf("unexpected board id %q", boardID)
			}
			return []string{}, nil
		},
		nil, nil, nil, nil, nil, nil,
		".", "revamp", "redis://localhost:6379/0", 3,
	)
	defer ui.Stop()

	time.Sleep(50 * time.Millisecond)
	ui.selectedBoardID = "board-1"
	ui.runGitSourceInput.SetText("revamp")
	ui.runGitMaxTasksInput.SetText("2")

	execute := ui.runGitCommands.GetItemSelectedFunc(2)
	if execute == nil {
		t.Fatalf("expected execute command")
	}
	execute()

	select {
	case call := <-startCalls:
		if call.boardID != "board-1" {
			t.Fatalf("expected board id board-1, got %q", call.boardID)
		}
		if call.sourceBranch != "revamp" {
			t.Fatalf("expected source branch revamp, got %q", call.sourceBranch)
		}
		if call.maxTasks != 2 {
			t.Fatalf("expected max tasks 2, got %d", call.maxTasks)
		}
		if call.redisURI != "redis://localhost:6379/0" {
			t.Fatalf("expected redis uri redis://localhost:6379/0, got %q", call.redisURI)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected start task tree command to be invoked")
	}

}
