package taskboard

import (
	"context"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type Repository interface {
	GetByBoardID(ctx context.Context, boardID string) (*domaintaskboard.Board, error)
	Save(ctx context.Context, board *domaintaskboard.Board) error
}
