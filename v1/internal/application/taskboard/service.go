package taskboard

import (
	"context"
	"fmt"
	"strings"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	"github.com/sirupsen/logrus"
)

type Service struct {
	repository Repository
	navigator  *domaintaskboard.Navigator
	logger     *logrus.Logger
}

func NewService(repository Repository, loggers ...*logrus.Logger) *Service {
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &Service{
		repository: repository,
		navigator:  domaintaskboard.NewNavigator(),
		logger:     logger,
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
		service.entry().WithError(err).WithField("event", "taskboard.list_board_ids").Error("failed to list board ids")
		return nil, fmt.Errorf("list boards: %w", err)
	}
	service.entry().WithFields(logrus.Fields{"event": "taskboard.list_board_ids", "board_count": len(boardIDs)}).Info("listed board ids")
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

func (service *Service) MarkTaskCanceledWithOutcome(ctx context.Context, boardID string, taskID string, outcome domaintaskboard.TaskOutcome) error {
	return service.markTaskStatusAndOutcome(ctx, boardID, taskID, domaintaskboard.StatusNotStarted, &outcome)
}

func (service *Service) CheckpointTaskResumeSession(ctx context.Context, boardID string, taskID string, resumeSessionID string) error {
	cleanSessionID := strings.TrimSpace(resumeSessionID)
	if cleanSessionID == "" {
		return nil
	}

	entry := service.entry().WithFields(logrus.Fields{
		"event":             "taskboard.checkpoint_resume_session",
		"board_id":          strings.TrimSpace(boardID),
		"task_id":           strings.TrimSpace(taskID),
		"resume_session_id": cleanSessionID,
	})

	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		return err
	}

	cleanTaskID := strings.TrimSpace(taskID)
	if cleanTaskID == "" {
		return fmt.Errorf("task_id is required")
	}

	for epicIndex := range board.Epics {
		for taskIndex := range board.Epics[epicIndex].Tasks {
			task := &board.Epics[epicIndex].Tasks[taskIndex]
			if task.ID != cleanTaskID {
				continue
			}
			if task.Outcome != nil && strings.TrimSpace(task.Outcome.ResumeSessionID) == cleanSessionID {
				return nil
			}
			outcome := domaintaskboard.TaskOutcome{}
			if task.Outcome != nil {
				outcome = *task.Outcome
			}
			if strings.TrimSpace(outcome.Status) == "" {
				outcome.Status = string(task.Status)
			}
			if strings.TrimSpace(outcome.Reason) == "" {
				outcome.Reason = "resume session checkpoint"
			}
			outcome.ResumeSessionID = cleanSessionID
			if err := board.SetTaskOutcome(cleanTaskID, outcome); err != nil {
				return err
			}
			if err := service.repository.Save(ctx, board); err != nil {
				entry.WithError(err).Error("failed to save board while checkpointing resume session")
				return fmt.Errorf("save board: %w", err)
			}
			entry.Info("checkpointed task resume session")
			return nil
		}
	}

	return fmt.Errorf("task not found: %s", cleanTaskID)
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
		service.entry().WithFields(logrus.Fields{"event": "taskboard.annotate_precompleted", "board_id": strings.TrimSpace(boardID), "modified": false}).Debug("no completed tasks required annotation")
		return nil
	}
	if err := service.repository.Save(ctx, board); err != nil {
		service.entry().WithError(err).WithFields(logrus.Fields{"event": "taskboard.annotate_precompleted", "board_id": strings.TrimSpace(boardID)}).Error("failed to save board after annotating completed tasks")
		return fmt.Errorf("save board: %w", err)
	}
	service.entry().WithFields(logrus.Fields{"event": "taskboard.annotate_precompleted", "board_id": strings.TrimSpace(boardID), "modified": true}).Info("annotated completed tasks without outcomes")
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
			resumeSessionID := ""
			if task.Outcome != nil {
				resumeSessionID = strings.TrimSpace(task.Outcome.ResumeSessionID)
			}
			if err := board.SetTaskStatus(task.ID, domaintaskboard.StatusNotStarted); err != nil {
				return 0, err
			}
			if err := board.SetTaskOutcome(task.ID, domaintaskboard.TaskOutcome{Status: "interrupted", Reason: cleanReason, ResumeSessionID: resumeSessionID}); err != nil {
				return 0, err
			}
			requeuedCount++
		}
	}

	if requeuedCount == 0 {
		service.entry().WithFields(logrus.Fields{"event": "taskboard.requeue_in_progress", "board_id": strings.TrimSpace(boardID), "requeued_count": 0}).Debug("no in-progress tasks to requeue")
		return 0, nil
	}
	if err := service.repository.Save(ctx, board); err != nil {
		service.entry().WithError(err).WithFields(logrus.Fields{"event": "taskboard.requeue_in_progress", "board_id": strings.TrimSpace(boardID)}).Error("failed to save board after requeueing in-progress tasks")
		return 0, fmt.Errorf("save board: %w", err)
	}
	service.entry().WithFields(logrus.Fields{"event": "taskboard.requeue_in_progress", "board_id": strings.TrimSpace(boardID), "requeued_count": requeuedCount}).Info("requeued in-progress tasks")
	return requeuedCount, nil
}

func (service *Service) markTaskStatus(ctx context.Context, boardID string, taskID string, status domaintaskboard.Status) error {
	return service.markTaskStatusAndOutcome(ctx, boardID, taskID, status, nil)
}

func (service *Service) markTaskStatusAndOutcome(ctx context.Context, boardID string, taskID string, status domaintaskboard.Status, outcome *domaintaskboard.TaskOutcome) error {
	entry := service.entry().WithFields(logrus.Fields{
		"event":    "taskboard.mark_task",
		"board_id": strings.TrimSpace(boardID),
		"task_id":  strings.TrimSpace(taskID),
		"status":   string(status),
	})
	board, err := service.loadBoard(ctx, boardID)
	if err != nil {
		entry.WithError(err).Error("failed to load board before marking task")
		return err
	}

	if err := board.SetTaskStatus(taskID, status); err != nil {
		entry.WithError(err).Error("failed to set task status")
		return err
	}
	if outcome != nil {
		if err := board.SetTaskOutcome(taskID, *outcome); err != nil {
			entry.WithError(err).Error("failed to set task outcome")
			return err
		}
	}

	if err := service.repository.Save(ctx, board); err != nil {
		entry.WithError(err).Error("failed to save board after marking task")
		return fmt.Errorf("save board: %w", err)
	}
	entry.Info("task status updated")
	return nil
}

func (service *Service) loadBoard(ctx context.Context, boardID string) (*domaintaskboard.Board, error) {
	cleanBoardID := strings.TrimSpace(boardID)
	entry := service.entry().WithFields(logrus.Fields{"event": "taskboard.load_board", "board_id": cleanBoardID})
	if cleanBoardID == "" {
		entry.Error("board_id is required")
		return nil, fmt.Errorf("board_id is required")
	}

	board, err := service.repository.GetByBoardID(ctx, cleanBoardID)
	if err != nil {
		entry.WithError(err).Error("failed to load board from repository")
		return nil, fmt.Errorf("load board: %w", err)
	}
	if board == nil {
		entry.Warn("board not found")
		return nil, fmt.Errorf("board not found: %s", cleanBoardID)
	}
	return board, nil
}

func (service *Service) entry() *logrus.Entry {
	if service == nil || service.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(service.logger)
}
