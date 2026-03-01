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
	"gorm.io/gorm/clause"
)

type boardSnapshotRecord struct {
	gorm.Model
	RunID      string `gorm:"column:run_id;size:255;not null;uniqueIndex:idx_tracker_board_snapshot,priority:1"`
	BoardID    string `gorm:"column:board_id;size:255;not null;uniqueIndex:idx_tracker_board_snapshot,priority:2"`
	SourceKind string `gorm:"column:source_kind;size:64;not null"`
	SourceRef  string `gorm:"column:source_ref"`
	Payload    []byte `gorm:"column:payload;not null"`
}

func (boardSnapshotRecord) TableName() string {
	return "tracker_board_snapshots"
}

type PostgresBoardSnapshotProvider struct {
	db       *gorm.DB
	upstream applicationtracker.Provider
}

func NewPostgresBoardSnapshotProvider(db *gorm.DB, upstream applicationtracker.Provider) (*PostgresBoardSnapshotProvider, error) {
	if db == nil {
		return nil, failures.WrapTerminal(errors.New("postgres board snapshot provider db is required"))
	}
	if upstream == nil {
		return nil, failures.WrapTerminal(errors.New("postgres board snapshot provider upstream is required"))
	}
	if err := db.AutoMigrate(&boardSnapshotRecord{}); err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("migrate tracker board snapshots: %w", err))
	}
	return &PostgresBoardSnapshotProvider{db: db, upstream: upstream}, nil
}

func (provider *PostgresBoardSnapshotProvider) SyncBoard(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	if provider == nil || provider.db == nil || provider.upstream == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("postgres board snapshot provider is not initialized"))
	}
	board, err := provider.upstream.SyncBoard(ctx, request)
	if err != nil {
		return domaintracker.Board{}, err
	}
	if err := board.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	payload, err := json.Marshal(board)
	if err != nil {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("encode board snapshot: %w", err))
	}
	record := boardSnapshotRecord{
		RunID:      strings.TrimSpace(board.RunID),
		BoardID:    strings.TrimSpace(board.BoardID),
		SourceKind: strings.TrimSpace(string(board.Source.Kind)),
		SourceRef:  strings.TrimSpace(board.Source.BoardID),
		Payload:    payload,
	}
	if err := provider.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "run_id"}, {Name: "board_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"source_kind", "source_ref", "payload", "updated_at"}),
	}).Create(&record).Error; err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("persist board snapshot: %w", err))
	}
	return board, nil
}

var _ applicationtracker.Provider = (*PostgresBoardSnapshotProvider)(nil)
