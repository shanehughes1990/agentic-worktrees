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

func (service *Service) GetNextTask(ctx context.Context, boardID string) (*domaintaskboard.Task, error) {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	return service.navigator.GetNextTask(board)
}

func (service *Service) GetReadyTasks(ctx context.Context, boardID string) ([]*domaintaskboard.Task, error) {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	return service.navigator.GetReadyTasks(board)
}

func (service *Service) ListBoardIDs(ctx context.Context) ([]string, error) {
	boardIDs, err := service.repository.ListBoardIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}
	return boardIDs, nil
}

func (service *Service) MarkTaskCompleted(ctx context.Context, boardID string, taskID string) error {
	return service.markTaskStatus(ctx, boardID, taskID, domaintaskboard.StatusCompleted)
}

func (service *Service) MarkTaskCompletedWithOutcome(ctx context.Context, boardID string, taskID string, outcome domaintaskboard.TaskOutcome) error {
	return service.markTaskStatusAndOutcome(ctx, boardID, taskID, domaintaskboard.StatusCompleted, &outcome)
}

func (service *Service) MarkTaskInProgress(ctx context.Context, boardID string, taskID string) error {
	return service.markTaskStatus(ctx, boardID, taskID, domaintaskboard.StatusInProgress)
}

func (service *Service) MarkTaskBlocked(ctx context.Context, boardID string, taskID string) error {
	return service.markTaskStatus(ctx, boardID, taskID, domaintaskboard.StatusBlocked)
}

func (service *Service) MarkTaskBlockedWithOutcome(ctx context.Context, boardID string, taskID string, outcome domaintaskboard.TaskOutcome) error {
	return service.markTaskStatusAndOutcome(ctx, boardID, taskID, domaintaskboard.StatusBlocked, &outcome)
}

func (service *Service) IsBoardCompleted(ctx context.Context, boardID string) (bool, error) {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return false, err
	}

	for _, epic := range board.Epics {
		for _, task := range epic.Tasks {
			if task.Status != domaintaskboard.StatusCompleted {
				return false, nil
			}
		}
	}

	return true, nil
}

func (service *Service) GetTaskByID(ctx context.Context, boardID string, taskID string) (*domaintaskboard.Task, error) {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	cleanTaskID := strings.TrimSpace(taskID)
	if cleanTaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	for epicIndex := range board.Epics {
		for taskIndex := range board.Epics[epicIndex].Tasks {
			task := &board.Epics[epicIndex].Tasks[taskIndex]
			if task.ID == cleanTaskID {
				copiedTask := *task
				return &copiedTask, nil
			}
		}
	}
	return nil, fmt.Errorf("task not found: %s", cleanTaskID)
}

func (service *Service) AnnotateCompletedTasksWithoutOutcome(ctx context.Context, boardID string, reason string) error {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return err
	}

	cleanReason := strings.TrimSpace(reason)
	if cleanReason == "" {
		cleanReason = "task already marked completed before current execution run"
	}

	modified := false
	for epicIndex := range board.Epics {
		for taskIndex := range board.Epics[epicIndex].Tasks {
			task := &board.Epics[epicIndex].Tasks[taskIndex]
			if task.Status != domaintaskboard.StatusCompleted {
				continue
			}
			if task.Outcome != nil {
				continue
			}
			outcome := domaintaskboard.TaskOutcome{Status: "precompleted", Reason: cleanReason}
			if err := board.SetTaskOutcome(task.ID, outcome); err != nil {
				return err
			}
			modified = true
		}
	}

	if !modified {
		return nil
	}
	if err := service.repository.Save(ctx, board); err != nil {
		return fmt.Errorf("save board: %w", err)
	}
	return nil
}

func (service *Service) RequeueInProgressTasks(ctx context.Context, boardID string, reason string) (int, error) {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return 0, err
	}

	cleanReason := strings.TrimSpace(reason)
	if cleanReason == "" {
		cleanReason = "task was in-progress when runner stopped; re-queued for resume"
	}

	requeuedCount := 0
	for epicIndex := range board.Epics {
		for taskIndex := range board.Epics[epicIndex].Tasks {
			task := &board.Epics[epicIndex].Tasks[taskIndex]
			if task.Status != domaintaskboard.StatusInProgress {
				continue
			}
			if err := board.SetTaskStatus(task.ID, domaintaskboard.StatusNotStarted); err != nil {
				return 0, err
			}
			if err := board.SetTaskOutcome(task.ID, domaintaskboard.TaskOutcome{Status: "interrupted", Reason: cleanReason}); err != nil {
				return 0, err
			}
			requeuedCount++
		}
	}

	if requeuedCount == 0 {
		return 0, nil
	}
	if err := service.repository.Save(ctx, board); err != nil {
		return 0, fmt.Errorf("save board: %w", err)
	}
	return requeuedCount, nil
}

func (service *Service) markTaskStatus(ctx context.Context, boardID string, taskID string, status domaintaskboard.Status) error {
	return service.markTaskStatusAndOutcome(ctx, boardID, taskID, status, nil)
}

func (service *Service) markTaskStatusAndOutcome(ctx context.Context, boardID string, taskID string, status domaintaskboard.Status, outcome *domaintaskboard.TaskOutcome) error {
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return err
	}

	if err := board.SetTaskStatus(taskID, status); err != nil {
		return err
	}
	if outcome != nil {
		if err := board.SetTaskOutcome(taskID, *outcome); err != nil {
			return err
		}
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
