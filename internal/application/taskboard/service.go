package taskboard

import (
	"context"
	"fmt"
	"strings"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type Service struct {
	repository Repository
	navigator  *domaintaskboard.Navigator
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		navigator:  domaintaskboard.NewNavigator(),
	}
}

func (service *Service) GetNextTask(ctx context.Context, boardID string) (*domaintaskboard.MicroTask, error) {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	return service.navigator.GetNextTask(board)
}

func (service *Service) GetReadyTasks(ctx context.Context, boardID string) ([]*domaintaskboard.MicroTask, error) {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	return service.navigator.GetReadyTasks(board)
}

func (service *Service) MarkTaskCompleted(ctx context.Context, boardID string, microTaskID string) error {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return err
	}

	if err := board.SetMicroTaskStatus(microTaskID, domaintaskboard.StatusCompleted); err != nil {
		return err
	}

	if err := service.repository.Save(ctx, board); err != nil {
		return fmt.Errorf("save board: %w", err)
	}
	return nil
}

func (service *Service) loadBoard(ctx context.Context, boardID string) (*domaintaskboard.Board, error) {
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanBoardID == "" {
		return nil, fmt.Errorf("board_id is required")
	}

	board, err := service.repository.GetByBoardID(ctx, cleanBoardID)
	if err != nil {
		return nil, fmt.Errorf("load board: %w", err)
	}
	if board == nil {
		return nil, fmt.Errorf("board not found: %s", cleanBoardID)
	}
	return board, nil
}
