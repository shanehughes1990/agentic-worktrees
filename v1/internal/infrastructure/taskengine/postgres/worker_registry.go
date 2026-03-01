package postgres

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type workerRegistryRecord struct {
	gorm.Model
	WorkerID      string `gorm:"column:worker_id;size:255;not null;uniqueIndex"`
	Capabilities  []byte `gorm:"column:capabilities;not null"`
	LastHeartbeat int64  `gorm:"column:last_heartbeat;not null"`
}

func (workerRegistryRecord) TableName() string {
	return "worker_registry"
}

type WorkerRegistry struct {
	db *gorm.DB
}

func NewWorkerRegistry(db *gorm.DB) (*WorkerRegistry, error) {
	if db == nil {
		return nil, errors.New("worker registry db is required")
	}
	if err := db.AutoMigrate(&workerRegistryRecord{}); err != nil {
		return nil, fmt.Errorf("worker registry migrate: %w", err)
	}
	return &WorkerRegistry{db: db}, nil
}

func (registry *WorkerRegistry) Upsert(ctx context.Context, advertisement taskengine.WorkerCapabilityAdvertisement) error {
	if registry == nil || registry.db == nil {
		return errors.New("worker registry is not initialized")
	}
	if err := advertisement.Validate(); err != nil {
		return err
	}
	encodedCapabilities, err := json.Marshal(advertisement.Capabilities)
	if err != nil {
		return fmt.Errorf("worker registry encode capabilities: %w", err)
	}
	record := workerRegistryRecord{
		WorkerID:      strings.TrimSpace(advertisement.WorkerID),
		Capabilities:  encodedCapabilities,
		LastHeartbeat: time.Now().UTC().Unix(),
	}
	if err := registry.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "worker_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"capabilities", "last_heartbeat", "updated_at"}),
	}).Create(&record).Error; err != nil {
		return fmt.Errorf("worker registry upsert: %w", err)
	}
	return nil
}
