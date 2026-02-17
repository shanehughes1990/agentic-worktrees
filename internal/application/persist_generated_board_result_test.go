package application

import (
	"context"
	"strings"
	"testing"

	entity "github.com/shanehughes1990/agentic-worktrees/internal/domain/entities"
	domainservices "github.com/shanehughes1990/agentic-worktrees/internal/domain/services"
)

type stubBoardRepository struct {
	saved []entity.Board
}

func (s *stubBoardRepository) Save(_ context.Context, board entity.Board) error {
	s.saved = append(s.saved, board)
	return nil
}

func (s *stubBoardRepository) GetByID(_ context.Context, _ string) (entity.Board, error) {
	return entity.Board{}, nil
}

func (s *stubBoardRepository) List(_ context.Context) ([]entity.Board, error) {
	return nil, nil
}

func (s *stubBoardRepository) DeleteByID(_ context.Context, _ string) error {
	return nil
}

func TestPersistGenerateTaskBoardResultCommandExecutePersistsValidBoard(t *testing.T) {
	repo := &stubBoardRepository{}
	command, err := NewPersistGenerateTaskBoardResultCommand(repo)
	if err != nil {
		t.Fatalf("new command: %v", err)
	}

	message := GenerateTaskBoardResultMessage{
		Metadata: validMetadata(),
		BoardJSON: `{
  "id": "board-1",
  "title": "Board",
  "epics": [
    {
      "id": "epic-1",
      "title": "Epic",
      "description": "Epic desc",
      "dependencies": [],
      "tasks": [
        {
          "id": "task-1",
          "title": "Task",
          "description": "Task desc",
          "status": "pending",
          "dependencies": []
        }
      ]
    }
  ],
  "created_at": "2026-02-16T00:00:00Z",
  "updated_at": "2026-02-16T00:00:00Z"
}`,
	}

	if err := command.Execute(context.Background(), message); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected board saved")
	}
}

func TestPersistGenerateTaskBoardResultCommandExecuteRejectsInvalidStatus(t *testing.T) {
	repo := &stubBoardRepository{}
	command, err := NewPersistGenerateTaskBoardResultCommand(repo)
	if err != nil {
		t.Fatalf("new command: %v", err)
	}

	message := GenerateTaskBoardResultMessage{
		Metadata: validMetadata(),
		BoardJSON: `{
  "id": "board-1",
  "title": "Board",
  "epics": [
    {
      "id": "epic-1",
      "title": "Epic",
      "description": "Epic desc",
      "dependencies": [],
      "tasks": [
        {
          "id": "task-1",
          "title": "Task",
          "description": "Task desc",
          "status": "queued",
          "dependencies": []
        }
      ]
    }
  ],
  "created_at": "2026-02-16T00:00:00Z",
  "updated_at": "2026-02-16T00:00:00Z"
}`,
	}

	if err := command.Execute(context.Background(), message); err == nil {
		t.Fatalf("expected invalid status error")
	}
}

func TestPersistGenerateTaskBoardResultCommandExecuteRejectsUnknownDependency(t *testing.T) {
	repo := &stubBoardRepository{}
	command, err := NewPersistGenerateTaskBoardResultCommand(repo)
	if err != nil {
		t.Fatalf("new command: %v", err)
	}

	message := GenerateTaskBoardResultMessage{
		Metadata: validMetadata(),
		BoardJSON: `{
  "id": "board-1",
  "title": "Board",
  "epics": [
    {
      "id": "epic-1",
      "title": "Epic",
      "description": "Epic desc",
      "dependencies": [],
      "tasks": [
        {
          "id": "task-1",
          "title": "Task",
          "description": "Task desc",
          "status": "pending",
          "dependencies": ["task-unknown"]
        }
      ]
    }
  ],
  "created_at": "2026-02-16T00:00:00Z",
  "updated_at": "2026-02-16T00:00:00Z"
}`,
	}

	err = command.Execute(context.Background(), message)
	if err == nil {
		t.Fatalf("expected dependency validation error")
	}
	if !strings.Contains(err.Error(), "unknown dependency") {
		t.Fatalf("expected unknown dependency error, got %v", err)
	}
}

func validMetadata() domainservices.AgentRequestMetadata {
	return domainservices.AgentRequestMetadata{JobID: "job-1", Model: "gpt-5.3-codex"}
}
