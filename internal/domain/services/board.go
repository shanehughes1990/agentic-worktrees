package services

import entity "github.com/shanehughes1990/agentic-worktrees/internal/domain/entities"

type TaskRef struct {
	EpicIndex int
	TaskIndex int
	TaskID    string
}

func NextTask(board entity.Board) (TaskRef, bool) {
	for epicIndex, epic := range board.Epics {
		for taskIndex, task := range epic.Tasks {
			if task.Status != entity.TaskStatusPending {
				continue
			}
			if !CanRunTask(board, task.ID) {
				continue
			}
			return TaskRef{EpicIndex: epicIndex, TaskIndex: taskIndex, TaskID: task.ID}, true
		}
	}
	return TaskRef{}, false
}

func CanRunTask(board entity.Board, taskID string) bool {
	taskRef, ok := FindTask(board, taskID)
	if !ok {
		return false
	}

	task := board.Epics[taskRef.EpicIndex].Tasks[taskRef.TaskIndex]
	for _, depID := range task.Dependencies {
		depRef, found := FindTask(board, depID)
		if !found {
			return false
		}
		depTask := board.Epics[depRef.EpicIndex].Tasks[depRef.TaskIndex]
		if depTask.Status != entity.TaskStatusCompleted {
			return false
		}
	}

	return true
}

func FindTask(board entity.Board, taskID string) (TaskRef, bool) {
	for epicIndex, epic := range board.Epics {
		for taskIndex, task := range epic.Tasks {
			if task.ID == taskID {
				return TaskRef{EpicIndex: epicIndex, TaskIndex: taskIndex, TaskID: task.ID}, true
			}
		}
	}
	return TaskRef{}, false
}

func SetTaskStatus(board *entity.Board, taskID string, status entity.TaskStatus) bool {
	if board == nil {
		return false
	}

	for epicIndex := range board.Epics {
		for taskIndex := range board.Epics[epicIndex].Tasks {
			if board.Epics[epicIndex].Tasks[taskIndex].ID == taskID {
				board.Epics[epicIndex].Tasks[taskIndex].Status = status
				return true
			}
		}
	}
	return false
}
