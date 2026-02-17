package repositories

import (
	"context"

	entity "github.com/shanehughes1990/agentic-worktrees/internal/domain/entities"
)

type BoardRepository interface {
	Save(ctx context.Context, board entity.Board) error
	GetByID(ctx context.Context, id string) (entity.Board, error)
	List(ctx context.Context) ([]entity.Board, error)
	DeleteByID(ctx context.Context, id string) error
}
