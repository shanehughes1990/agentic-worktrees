package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LocalJSONProvider struct {
	baseDirectory string
}

func NewLocalJSONProvider(baseDirectory string) (*LocalJSONProvider, error) {
	cleanBaseDirectory := strings.TrimSpace(baseDirectory)
	if cleanBaseDirectory == "" {
		return nil, failures.WrapTerminal(errors.New("base directory is required"))
	}
	return &LocalJSONProvider{baseDirectory: cleanBaseDirectory}, nil
}

func (provider *LocalJSONProvider) SyncBoard(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	if err := request.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if request.Source.Kind != domaintracker.SourceKindLocalJSON {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("local json provider does not support source kind %q", request.Source.Kind))
	}
	boardPath := provider.resolvePath(request.Source.Location)
	payload, err := os.ReadFile(boardPath)
	if err != nil {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("read tracker board json: %w", err))
	}
	select {
	case <-ctx.Done():
		return domaintracker.Board{}, ctx.Err()
	default:
	}

	var board domaintracker.Board
	if err := json.Unmarshal(payload, &board); err != nil {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("decode tracker board json: %w", err))
	}
	if strings.TrimSpace(board.RunID) == "" {
		board.RunID = request.RunID
	}
	board.Source = request.Source
	if strings.TrimSpace(board.Source.BoardID) == "" {
		board.Source.BoardID = board.BoardID
	}
	return board, nil
}

func (provider *LocalJSONProvider) resolvePath(path string) string {
	cleanPath := strings.TrimSpace(path)
	if filepath.IsAbs(cleanPath) {
		return cleanPath
	}
	return filepath.Join(provider.baseDirectory, cleanPath)
}
