package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type PostgresInternalProvider struct {
	db *gorm.DB
}

func NewPostgresInternalProvider(db *gorm.DB) (*PostgresInternalProvider, error) {
	if db == nil {
		return nil, failures.WrapTerminal(errors.New("postgres internal provider db is required"))
	}
	if err := db.AutoMigrate(&boardSnapshotRecord{}); err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("migrate internal tracker snapshots: %w", err))
	}
	return &PostgresInternalProvider{db: db}, nil
}

func (provider *PostgresInternalProvider) SyncBoard(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	if provider == nil || provider.db == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("postgres internal provider is not initialized"))
	}
	if err := request.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if request.Source.Kind != domaintracker.SourceKindInternal {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("postgres internal provider does not support source kind %q", request.Source.Kind))
	}

	snapshotKey := strings.TrimSpace(request.Source.BoardID)
	if snapshotKey == "" {
		snapshotKey = strings.TrimSpace(request.Source.Location)
	}
	if snapshotKey == "" {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("internal tracker board source reference is required"))
	}

	var record boardSnapshotRecord
	err := provider.db.WithContext(ctx).
		Where("run_id = ? AND board_id = ?", strings.TrimSpace(request.ProjectID), snapshotKey).
		Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("internal tracker board %q not found for project %q", snapshotKey, request.ProjectID))
		}
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load internal tracker board snapshot: %w", err))
	}

	var board domaintracker.Board
	if err := json.Unmarshal(record.Payload, &board); err != nil {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("decode internal tracker board snapshot: %w", err))
	}
	board.RunID = strings.TrimSpace(request.RunID)
	board.Source = request.Source
	if strings.TrimSpace(board.Source.BoardID) == "" {
		board.Source.BoardID = snapshotKey
	}
	if strings.TrimSpace(board.BoardID) == "" {
		board.BoardID = snapshotKey
	}
	return board, nil
}

var _ applicationtracker.Provider = (*PostgresInternalProvider)(nil)
