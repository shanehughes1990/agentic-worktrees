package taskboard

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

func BuildBoardFromResponse(runID string, response string) (*domaintaskboard.Board, error) {
	payload, err := extractJSONObject(strings.TrimSpace(response))
	if err != nil {
		return nil, err
	}

	board := &domaintaskboard.Board{}
	if err := json.Unmarshal([]byte(payload), board); err != nil {
		return nil, fmt.Errorf("unmarshal taskboard json: %w", err)
	}

	now := time.Now().UTC()
	if strings.TrimSpace(board.BoardID) == "" {
		board.BoardID = runID
	}
	if strings.TrimSpace(board.RunID) == "" {
		board.RunID = runID
	}
	board.CreatedAt = now
	board.UpdatedAt = now

	for epicIndex := range board.Epics {
		if strings.TrimSpace(board.Epics[epicIndex].BoardID) == "" {
			board.Epics[epicIndex].BoardID = board.BoardID
		}
		board.Epics[epicIndex].CreatedAt = now
		board.Epics[epicIndex].UpdatedAt = now
		for taskIndex := range board.Epics[epicIndex].Tasks {
			if strings.TrimSpace(board.Epics[epicIndex].Tasks[taskIndex].BoardID) == "" {
				board.Epics[epicIndex].Tasks[taskIndex].BoardID = board.BoardID
			}
			board.Epics[epicIndex].Tasks[taskIndex].CreatedAt = now
			board.Epics[epicIndex].Tasks[taskIndex].UpdatedAt = now
			if board.Epics[epicIndex].Tasks[taskIndex].Outcome != nil {
				board.Epics[epicIndex].Tasks[taskIndex].Outcome.UpdatedAt = now
			}
		}
	}

	if err := board.ValidateBasics(); err != nil {
		return nil, fmt.Errorf("validate taskboard: %w", err)
	}
	if err := board.ValidateComplete(); err != nil {
		return nil, fmt.Errorf("validate complete taskboard: %w", err)
	}

	return board, nil
}

func extractJSONObject(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty response payload")
	}

	clean := strings.TrimSpace(input)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start < 0 || end <= start {
		return "", fmt.Errorf("response does not contain json object")
	}
	return clean[start : end+1], nil
}
