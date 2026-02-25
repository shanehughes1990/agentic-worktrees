package taskboard

import (
	"context"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type Repository interface {
	ListBoardIDs(ctx context.Context) ([]string, error)
	GetByBoardID(ctx context.Context, boardID string) (*domaintaskboard.Board, error)
	Save(ctx context.Context, board *domaintaskboard.Board) error
}
