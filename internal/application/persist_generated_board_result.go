package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	entity "github.com/shanehughes1990/agentic-worktrees/internal/domain/entities"
	domainrepositories "github.com/shanehughes1990/agentic-worktrees/internal/domain/repositories"
)

type PersistGenerateTaskBoardResultCommand struct {
	repository domainrepositories.BoardRepository
}

func NewPersistGenerateTaskBoardResultCommand(repository domainrepositories.BoardRepository) (*PersistGenerateTaskBoardResultCommand, error) {
	if repository == nil {
		return nil, fmt.Errorf("board repository cannot be nil")
	}
	return &PersistGenerateTaskBoardResultCommand{repository: repository}, nil
}

func (c *PersistGenerateTaskBoardResultCommand) Execute(ctx context.Context, message GenerateTaskBoardResultMessage) error {
	if c == nil {
		return fmt.Errorf("command cannot be nil")
	}
	if err := message.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(message.Error) != "" {
		return fmt.Errorf("generation failed: %s", message.Error)
	}
	if strings.TrimSpace(message.BoardJSON) == "" {
		return fmt.Errorf("board_json is required")
	}

	board, err := parseAndValidateBoardJSON(message.BoardJSON)
	if err != nil {
		return err
	}
	return c.repository.Save(ctx, board)
}

type boardResultDTO struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Epics     []boardEpicDTO `json:"epics"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type boardEpicDTO struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Dependencies []string       `json:"dependencies"`
	Tasks        []boardTaskDTO `json:"tasks"`
}

type boardTaskDTO struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies"`
}

func parseAndValidateBoardJSON(boardJSON string) (entity.Board, error) {
	decoder := json.NewDecoder(strings.NewReader(boardJSON))
	decoder.DisallowUnknownFields()

	var dto boardResultDTO
	if err := decoder.Decode(&dto); err != nil {
		return entity.Board{}, fmt.Errorf("decode board json: %w", err)
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return entity.Board{}, fmt.Errorf("decode board json: trailing content is not allowed")
		}
		return entity.Board{}, fmt.Errorf("decode board json trailing content: %w", err)
	}

	if strings.TrimSpace(dto.ID) == "" {
		return entity.Board{}, fmt.Errorf("board.id is required")
	}
	if strings.TrimSpace(dto.Title) == "" {
		return entity.Board{}, fmt.Errorf("board.title is required")
	}
	if len(dto.Epics) == 0 {
		return entity.Board{}, fmt.Errorf("board.epics is required")
	}
	if dto.CreatedAt.IsZero() {
		return entity.Board{}, fmt.Errorf("board.created_at is required")
	}
	if dto.UpdatedAt.IsZero() {
		return entity.Board{}, fmt.Errorf("board.updated_at is required")
	}

	epicIDs := make(map[string]struct{}, len(dto.Epics))
	taskIDs := make(map[string]struct{})

	board := entity.Board{
		ID:        strings.TrimSpace(dto.ID),
		Title:     strings.TrimSpace(dto.Title),
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		Epics:     make([]entity.Epic, 0, len(dto.Epics)),
	}

	for epicIndex, epicDTO := range dto.Epics {
		epicID := strings.TrimSpace(epicDTO.ID)
		if epicID == "" {
			return entity.Board{}, fmt.Errorf("board.epics[%d].id is required", epicIndex)
		}
		if _, exists := epicIDs[epicID]; exists {
			return entity.Board{}, fmt.Errorf("duplicate epic id: %s", epicID)
		}
		epicIDs[epicID] = struct{}{}

		if strings.TrimSpace(epicDTO.Title) == "" {
			return entity.Board{}, fmt.Errorf("board.epics[%d].title is required", epicIndex)
		}
		if strings.TrimSpace(epicDTO.Description) == "" {
			return entity.Board{}, fmt.Errorf("board.epics[%d].description is required", epicIndex)
		}
		if len(epicDTO.Tasks) == 0 {
			return entity.Board{}, fmt.Errorf("board.epics[%d].tasks is required", epicIndex)
		}

		epic := entity.Epic{
			ID:           epicID,
			Title:        strings.TrimSpace(epicDTO.Title),
			Description:  strings.TrimSpace(epicDTO.Description),
			Dependencies: normalizeIDs(epicDTO.Dependencies),
			Tasks:        make([]entity.Task, 0, len(epicDTO.Tasks)),
		}

		for taskIndex, taskDTO := range epicDTO.Tasks {
			taskID := strings.TrimSpace(taskDTO.ID)
			if taskID == "" {
				return entity.Board{}, fmt.Errorf("board.epics[%d].tasks[%d].id is required", epicIndex, taskIndex)
			}
			if _, exists := taskIDs[taskID]; exists {
				return entity.Board{}, fmt.Errorf("duplicate task id: %s", taskID)
			}
			taskIDs[taskID] = struct{}{}

			status, err := parseTaskStatus(taskDTO.Status)
			if err != nil {
				return entity.Board{}, fmt.Errorf("board.epics[%d].tasks[%d].status: %w", epicIndex, taskIndex, err)
			}
			if strings.TrimSpace(taskDTO.Title) == "" {
				return entity.Board{}, fmt.Errorf("board.epics[%d].tasks[%d].title is required", epicIndex, taskIndex)
			}
			if strings.TrimSpace(taskDTO.Description) == "" {
				return entity.Board{}, fmt.Errorf("board.epics[%d].tasks[%d].description is required", epicIndex, taskIndex)
			}

			epic.Tasks = append(epic.Tasks, entity.Task{
				ID:           taskID,
				Title:        strings.TrimSpace(taskDTO.Title),
				Description:  strings.TrimSpace(taskDTO.Description),
				Status:       status,
				Dependencies: normalizeIDs(taskDTO.Dependencies),
			})
		}

		board.Epics = append(board.Epics, epic)
	}

	for _, epic := range board.Epics {
		for _, dependencyID := range epic.Dependencies {
			if dependencyID == epic.ID {
				return entity.Board{}, fmt.Errorf("epic %s cannot depend on itself", epic.ID)
			}
			if _, exists := epicIDs[dependencyID]; !exists {
				return entity.Board{}, fmt.Errorf("epic %s has unknown dependency %s", epic.ID, dependencyID)
			}
		}
		for _, task := range epic.Tasks {
			for _, dependencyID := range task.Dependencies {
				if dependencyID == task.ID {
					return entity.Board{}, fmt.Errorf("task %s cannot depend on itself", task.ID)
				}
				if _, exists := taskIDs[dependencyID]; !exists {
					return entity.Board{}, fmt.Errorf("task %s has unknown dependency %s", task.ID, dependencyID)
				}
			}
		}
	}

	return board, nil
}

func parseTaskStatus(status string) (entity.TaskStatus, error) {
	resolved := entity.TaskStatus(strings.TrimSpace(status))
	switch resolved {
	case entity.TaskStatusPending, entity.TaskStatusInProgress, entity.TaskStatusCompleted, entity.TaskStatusFailed:
		return resolved, nil
	default:
		return "", fmt.Errorf("invalid status %q", status)
	}
}

func normalizeIDs(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}
