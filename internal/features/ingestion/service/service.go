package service

import (
	"context"

	boarddomain "github.com/shanehughes1990/agentic-worktrees/internal/features/board/domain"
	"github.com/shanehughes1990/agentic-worktrees/internal/features/ingestion/pipeline"
)

type Planner interface {
	PlanBoard(ctx context.Context, sourceScope string, files []pipeline.ScopeFile) (boarddomain.Board, error)
}

type Service struct {
	planner Planner
}

func New(planner Planner) *Service {
	return &Service{planner: planner}
}

func (s *Service) Plan(ctx context.Context, scopePath string) (boarddomain.Board, error) {
	files, err := pipeline.CollectScopeFiles(scopePath)
	if err != nil {
		return boarddomain.Board{}, err
	}

	return s.planner.PlanBoard(ctx, scopePath, files)
}
