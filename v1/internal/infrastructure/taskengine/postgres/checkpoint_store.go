package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type checkpointRecord struct {
	gorm.Model
	IdempotencyKey string `gorm:"column:idempotency_key;uniqueIndex;size:255;not null"`
	Step           string `gorm:"column:step;not null"`
	Token          string `gorm:"column:token;not null"`
}

func (checkpointRecord) TableName() string {
	return "job_checkpoints"
}

type PostgresCheckpointStore struct {
	db *gorm.DB
}

func NewPostgresCheckpointStore(db *gorm.DB) (*PostgresCheckpointStore, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres checkpoint store: db is required")
	}
	if err := db.AutoMigrate(&checkpointRecord{}); err != nil {
		return nil, fmt.Errorf("postgres checkpoint store: migrate: %w", err)
	}
	return &PostgresCheckpointStore{db: db}, nil
}

func (store *PostgresCheckpointStore) Save(ctx context.Context, idempotencyKey string, checkpoint taskengine.Checkpoint) error {
	if store == nil || store.db == nil {
		return fmt.Errorf("postgres checkpoint store: db is not initialized")
	}
	key := strings.TrimSpace(idempotencyKey)
	if key == "" {
		return fmt.Errorf("postgres checkpoint store: idempotency_key is required")
	}
	if strings.TrimSpace(checkpoint.Step) == "" || strings.TrimSpace(checkpoint.Token) == "" {
		return fmt.Errorf("postgres checkpoint store: checkpoint step and token are required")
	}
	record := checkpointRecord{
		IdempotencyKey: key,
		Step:           strings.TrimSpace(checkpoint.Step),
		Token:          strings.TrimSpace(checkpoint.Token),
	}
	if err := store.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "idempotency_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"step", "token", "updated_at"}),
	}).Create(&record).Error; err != nil {
		return fmt.Errorf("postgres checkpoint store: save: %w", err)
	}
	return nil
}

func (store *PostgresCheckpointStore) Load(ctx context.Context, idempotencyKey string) (*taskengine.Checkpoint, error) {
	if store == nil || store.db == nil {
		return nil, fmt.Errorf("postgres checkpoint store: db is not initialized")
	}
	key := strings.TrimSpace(idempotencyKey)
	if key == "" {
		return nil, fmt.Errorf("postgres checkpoint store: idempotency_key is required")
	}
	var record checkpointRecord
	err := store.db.WithContext(ctx).First(&record, "idempotency_key = ?", key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("postgres checkpoint store: load: %w", err)
	}
	return &taskengine.Checkpoint{Step: record.Step, Token: record.Token}, nil
}

var _ taskengine.CheckpointStore = (*PostgresCheckpointStore)(nil)
