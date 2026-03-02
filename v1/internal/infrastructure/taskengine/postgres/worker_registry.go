package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"agentic-orchestrator/internal/application/taskengine"
	applicationworker "agentic-orchestrator/internal/application/worker"
	domainworker "agentic-orchestrator/internal/domain/worker"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type workerRegistryRecord struct {
	gorm.Model
	WorkerID          string `gorm:"column:worker_id;size:255;not null;uniqueIndex"`
	Epoch             int64  `gorm:"column:epoch;not null"`
	State             string `gorm:"column:state;size:64;not null;index"`
	DesiredState      string `gorm:"column:desired_state;size:64;not null;index"`
	CapabilitiesJSON  []byte `gorm:"column:capabilities_json;not null"`
	LastHeartbeatUnix int64  `gorm:"column:last_heartbeat_unix;not null;index"`
	LeaseExpiresUnix  int64  `gorm:"column:lease_expires_unix;not null;index"`
	RogueReason       string `gorm:"column:rogue_reason;type:text"`
}

func (workerRegistryRecord) TableName() string {
	return "worker_registry"
}

type workerRegistrySettingsRecord struct {
	ID                       uint  `gorm:"primaryKey"`
	HeartbeatIntervalSeconds int64 `gorm:"column:heartbeat_interval_seconds;not null"`
	ResponseDeadlineSeconds  int64 `gorm:"column:response_deadline_seconds;not null"`
	StaleAfterSeconds        int64 `gorm:"column:stale_after_seconds;not null"`
	DrainTimeoutSeconds      int64 `gorm:"column:drain_timeout_seconds;not null"`
	TerminateTimeoutSeconds  int64 `gorm:"column:terminate_timeout_seconds;not null"`
	RogueThreshold           int64 `gorm:"column:rogue_threshold;not null"`
	UpdatedAtUnix            int64 `gorm:"column:updated_at_unix;not null"`
	UpdatedAt                time.Time
	CreatedAt                time.Time
}

func (workerRegistrySettingsRecord) TableName() string {
	return "worker_registry_settings"
}

type WorkerRegistry struct {
	db *gorm.DB
}

func NewWorkerRegistry(db *gorm.DB) (*WorkerRegistry, error) {
	if db == nil {
		return nil, fmt.Errorf("worker registry db is required")
	}
	if err := db.AutoMigrate(&workerRegistryRecord{}, &workerRegistrySettingsRecord{}); err != nil {
		return nil, fmt.Errorf("worker registry migrate: %w", err)
	}
	return &WorkerRegistry{db: db}, nil
}

func (registry *WorkerRegistry) Upsert(ctx context.Context, advertisement taskengine.WorkerCapabilityAdvertisement) error {
	if err := advertisement.Validate(); err != nil {
		return err
	}
	capabilities := make([]taskengine.JobKind, 0, len(advertisement.Capabilities))
	for _, capability := range advertisement.Capabilities {
		capabilities = append(capabilities, capability.Kind)
	}
	_, err := registry.Register(ctx, advertisement.WorkerID, capabilities, time.Now().UTC(), time.Now().UTC().Add(45*time.Second))
	return err
}

func (registry *WorkerRegistry) Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainworker.Worker, error) {
	if registry == nil || registry.db == nil {
		return nil, fmt.Errorf("worker registry is not initialized")
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, fmt.Errorf("worker_id is required")
	}
	record := workerRegistryRecord{WorkerID: workerID}
	transactionErr := registry.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("worker_id = ?", workerID).Take(&record).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("load worker registration: %w", err)
			}
			record.Epoch = 1
		} else {
			record.Epoch += 1
		}
		encodedCapabilities, err := encodeCapabilities(capabilities)
		if err != nil {
			return err
		}
		record.State = string(domainworker.StateHealthy)
		record.DesiredState = string(domainworker.StateHealthy)
		record.CapabilitiesJSON = encodedCapabilities
		record.LastHeartbeatUnix = heartbeatAt.UTC().Unix()
		record.LeaseExpiresUnix = leaseExpiresAt.UTC().Unix()
		record.RogueReason = ""
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "worker_id"}}, DoUpdates: clause.AssignmentColumns([]string{"epoch", "state", "desired_state", "capabilities_json", "last_heartbeat_unix", "lease_expires_unix", "rogue_reason", "updated_at"})}).Create(&record).Error; err != nil {
			return fmt.Errorf("upsert worker registration: %w", err)
		}
		return nil
	})
	if transactionErr != nil {
		return nil, transactionErr
	}
	return registry.toDomain(record)
}

func (registry *WorkerRegistry) RenewHeartbeat(ctx context.Context, workerID string, epoch int64, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainworker.Worker, error) {
	if registry == nil || registry.db == nil {
		return nil, fmt.Errorf("worker registry is not initialized")
	}
	record := workerRegistryRecord{}
	err := registry.db.WithContext(ctx).Where("worker_id = ?", strings.TrimSpace(workerID)).Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %q", applicationworker.ErrWorkerNotRegistered, strings.TrimSpace(workerID))
		}
		return nil, fmt.Errorf("load worker heartbeat: %w", err)
	}
	if record.Epoch != epoch {
		return nil, fmt.Errorf("%w: expected=%d actual=%d", applicationworker.ErrEpochMismatch, epoch, record.Epoch)
	}
	record.LastHeartbeatUnix = heartbeatAt.UTC().Unix()
	record.LeaseExpiresUnix = leaseExpiresAt.UTC().Unix()
	record.State = string(domainworker.StateHealthy)
	if err := registry.db.WithContext(ctx).Model(&workerRegistryRecord{}).Where("worker_id = ?", strings.TrimSpace(workerID)).Updates(map[string]any{"state": record.State, "last_heartbeat_unix": record.LastHeartbeatUnix, "lease_expires_unix": record.LeaseExpiresUnix, "updated_at": time.Now().UTC()}).Error; err != nil {
		return nil, fmt.Errorf("update worker heartbeat: %w", err)
	}
	return registry.toDomain(record)
}

func (registry *WorkerRegistry) UpdateState(ctx context.Context, workerID string, epoch int64, state domainworker.State, desiredState domainworker.State, reason string, changedAt time.Time) (*domainworker.Worker, error) {
	if registry == nil || registry.db == nil {
		return nil, fmt.Errorf("worker registry is not initialized")
	}
	record := workerRegistryRecord{}
	err := registry.db.WithContext(ctx).Where("worker_id = ?", strings.TrimSpace(workerID)).Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("worker %q is not registered", strings.TrimSpace(workerID))
		}
		return nil, fmt.Errorf("load worker state: %w", err)
	}
	if epoch > 0 && record.Epoch != epoch {
		return nil, fmt.Errorf("%w: expected=%d actual=%d", applicationworker.ErrEpochMismatch, epoch, record.Epoch)
	}
	record.State = string(state)
	record.DesiredState = string(desiredState)
	record.RogueReason = strings.TrimSpace(reason)
	if err := registry.db.WithContext(ctx).Model(&workerRegistryRecord{}).Where("worker_id = ?", strings.TrimSpace(workerID)).Updates(map[string]any{"state": record.State, "desired_state": record.DesiredState, "rogue_reason": record.RogueReason, "updated_at": changedAt.UTC()}).Error; err != nil {
		return nil, fmt.Errorf("update worker state: %w", err)
	}
	return registry.toDomain(record)
}

func (registry *WorkerRegistry) RemoveRegistration(ctx context.Context, workerID string, epoch int64) error {
	if registry == nil || registry.db == nil {
		return fmt.Errorf("worker registry is not initialized")
	}
	trimmedWorkerID := strings.TrimSpace(workerID)
	if trimmedWorkerID == "" {
		return fmt.Errorf("worker_id is required")
	}
	record := workerRegistryRecord{}
	err := registry.db.WithContext(ctx).Where("worker_id = ?", trimmedWorkerID).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load worker registration for delete: %w", err)
	}
	if epoch > 0 && record.Epoch != epoch {
		return fmt.Errorf("%w: expected=%d actual=%d", applicationworker.ErrEpochMismatch, epoch, record.Epoch)
	}
	if err := registry.db.WithContext(ctx).Unscoped().Where("worker_id = ?", trimmedWorkerID).Delete(&workerRegistryRecord{}).Error; err != nil {
		return fmt.Errorf("delete worker registration: %w", err)
	}
	return nil
}

func (registry *WorkerRegistry) ListWorkers(ctx context.Context, limit int) ([]domainworker.Worker, error) {
	if registry == nil || registry.db == nil {
		return nil, fmt.Errorf("worker registry is not initialized")
	}
	if limit <= 0 {
		limit = 50
	}
	records := make([]workerRegistryRecord, 0)
	if err := registry.db.WithContext(ctx).Order("updated_at DESC").Limit(limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list workers: %w", err)
	}
	workers := make([]domainworker.Worker, 0, len(records))
	for _, record := range records {
		worker, err := registry.toDomain(record)
		if err != nil {
			return nil, err
		}
		workers = append(workers, *worker)
	}
	return workers, nil
}

func (registry *WorkerRegistry) ListStaleWorkers(ctx context.Context, staleBefore time.Time, limit int) ([]domainworker.Worker, error) {
	if registry == nil || registry.db == nil {
		return nil, fmt.Errorf("worker registry is not initialized")
	}
	if limit <= 0 {
		limit = 100
	}
	records := make([]workerRegistryRecord, 0)
	if err := registry.db.WithContext(ctx).Where("lease_expires_unix < ? AND state IN ?", staleBefore.UTC().Unix(), []string{string(domainworker.StateHealthy), string(domainworker.StateDegraded), string(domainworker.StateDraining)}).Order("lease_expires_unix ASC").Limit(limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list stale workers: %w", err)
	}
	workers := make([]domainworker.Worker, 0, len(records))
	for _, record := range records {
		worker, err := registry.toDomain(record)
		if err != nil {
			return nil, err
		}
		workers = append(workers, *worker)
	}
	return workers, nil
}

func (registry *WorkerRegistry) GetSettings(ctx context.Context) (domainworker.Settings, error) {
	if registry == nil || registry.db == nil {
		return domainworker.Settings{}, fmt.Errorf("worker registry is not initialized")
	}
	record := workerRegistrySettingsRecord{}
	err := registry.db.WithContext(ctx).Where("id = 1").Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainworker.Settings{}, applicationworker.ErrSettingsNotFound
	} else if err != nil {
		return domainworker.Settings{}, fmt.Errorf("load worker settings: %w", err)
	}
	settings := domainworker.Settings{HeartbeatInterval: time.Duration(record.HeartbeatIntervalSeconds) * time.Second, ResponseDeadline: time.Duration(record.ResponseDeadlineSeconds) * time.Second, StaleAfter: time.Duration(record.StaleAfterSeconds) * time.Second, DrainTimeout: time.Duration(record.DrainTimeoutSeconds) * time.Second, TerminateTimeout: time.Duration(record.TerminateTimeoutSeconds) * time.Second, RogueThreshold: int(record.RogueThreshold), UpdatedAt: time.Unix(record.UpdatedAtUnix, 0).UTC()}
	if err := settings.Validate(); err != nil {
		return domainworker.Settings{}, err
	}
	return settings, nil
}

func (registry *WorkerRegistry) UpsertSettings(ctx context.Context, settings domainworker.Settings) (domainworker.Settings, error) {
	if registry == nil || registry.db == nil {
		return domainworker.Settings{}, fmt.Errorf("worker registry is not initialized")
	}
	if err := settings.Validate(); err != nil {
		return domainworker.Settings{}, err
	}
	record := workerRegistrySettingsRecord{ID: 1, HeartbeatIntervalSeconds: int64(settings.HeartbeatInterval.Seconds()), ResponseDeadlineSeconds: int64(settings.ResponseDeadline.Seconds()), StaleAfterSeconds: int64(settings.StaleAfter.Seconds()), DrainTimeoutSeconds: int64(settings.DrainTimeout.Seconds()), TerminateTimeoutSeconds: int64(settings.TerminateTimeout.Seconds()), RogueThreshold: int64(settings.RogueThreshold), UpdatedAtUnix: settings.UpdatedAt.UTC().Unix()}
	if err := registry.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"heartbeat_interval_seconds", "response_deadline_seconds", "stale_after_seconds", "drain_timeout_seconds", "terminate_timeout_seconds", "rogue_threshold", "updated_at_unix", "updated_at"})}).Create(&record).Error; err != nil {
		return domainworker.Settings{}, fmt.Errorf("upsert worker settings: %w", err)
	}
	return settings, nil
}

func encodeCapabilities(capabilities []taskengine.JobKind) ([]byte, error) {
	normalized := make([]string, 0, len(capabilities))
	seen := map[string]struct{}{}
	for _, capability := range capabilities {
		value := strings.TrimSpace(string(capability))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("at least one capability is required")
	}
	return json.Marshal(normalized)
}

func decodeCapabilities(raw []byte) ([]taskengine.JobKind, error) {
	decoded := make([]string, 0)
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("decode capabilities: %w", err)
	}
	capabilities := make([]taskengine.JobKind, 0, len(decoded))
	for _, value := range decoded {
		capabilities = append(capabilities, taskengine.JobKind(strings.TrimSpace(value)))
	}
	return capabilities, nil
}

func (registry *WorkerRegistry) toDomain(record workerRegistryRecord) (*domainworker.Worker, error) {
	capabilities, err := decodeCapabilities(record.CapabilitiesJSON)
	if err != nil {
		return nil, err
	}
	worker := &domainworker.Worker{
		WorkerID:       strings.TrimSpace(record.WorkerID),
		Epoch:          record.Epoch,
		State:          domainworker.State(strings.TrimSpace(record.State)),
		DesiredState:   domainworker.State(strings.TrimSpace(record.DesiredState)),
		Capabilities:   capabilities,
		LastHeartbeat:  time.Unix(record.LastHeartbeatUnix, 0).UTC(),
		LeaseExpiresAt: time.Unix(record.LeaseExpiresUnix, 0).UTC(),
		RogueReason:    strings.TrimSpace(record.RogueReason),
		UpdatedAt:      record.UpdatedAt.UTC(),
	}
	if err := worker.Validate(); err != nil {
		return nil, err
	}
	return worker, nil
}

var _ applicationworker.Repository = (*WorkerRegistry)(nil)
