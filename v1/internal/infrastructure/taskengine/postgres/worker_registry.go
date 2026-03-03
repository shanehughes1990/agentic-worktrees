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
	domainrealtime "agentic-orchestrator/internal/domain/realtime"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type workerRegistryRecord struct {
	gorm.Model
	WorkerID          string `gorm:"column:worker_id;size:255;not null;uniqueIndex"`
	Epoch             int64  `gorm:"column:epoch;not null"`
	State             string `gorm:"column:state;size:64;not null;index"`
	CapabilitiesJSON  []byte `gorm:"column:capabilities_json;not null"`
	LastHeartbeatUnix int64  `gorm:"column:last_heartbeat_unix;not null;index"`
	LeaseExpiresUnix  int64  `gorm:"column:lease_expires_unix;not null;index"`
}

func (workerRegistryRecord) TableName() string {
	return "worker_registry"
}

type workerRegistrySettingsRecord struct {
	ID                       uint  `gorm:"primaryKey"`
	HeartbeatIntervalSeconds int64 `gorm:"column:heartbeat_interval_seconds;not null"`
	ResponseDeadlineSeconds  int64 `gorm:"column:response_deadline_seconds;not null"`
	UpdatedAtUnix            int64 `gorm:"column:updated_at_unix;not null"`
	UpdatedAt                time.Time
	CreatedAt                time.Time
}

type workerRegistrationSubmissionRecord struct {
	ID               uint   `gorm:"primaryKey"`
	SubmissionID     string `gorm:"column:submission_id;size:255;not null;uniqueIndex"`
	WorkerID         string `gorm:"column:worker_id;size:255;not null;index"`
	RequestedAtUnix  int64  `gorm:"column:requested_at_unix;not null;index"`
	ExpiresAtUnix    int64  `gorm:"column:expires_at_unix;not null;index"`
	Status           string `gorm:"column:status;size:64;not null;index"`
	CapabilitiesJSON []byte `gorm:"column:capabilities_json;not null"`
	ReasonsJSON      []byte `gorm:"column:reasons_json;not null"`
	ResolvedAtUnix   int64  `gorm:"column:resolved_at_unix;not null;default:0"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (workerRegistrationSubmissionRecord) TableName() string {
	return "worker_registration_submissions"
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
	if err := db.AutoMigrate(&workerRegistryRecord{}, &workerRegistrySettingsRecord{}, &workerRegistrationSubmissionRecord{}); err != nil {
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

func (registry *WorkerRegistry) Register(ctx context.Context, workerID string, capabilities []taskengine.JobKind, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainrealtime.Worker, error) {
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
		record.State = string(domainrealtime.StateHealthy)
		record.CapabilitiesJSON = encodedCapabilities
		record.LastHeartbeatUnix = heartbeatAt.UTC().Unix()
		record.LeaseExpiresUnix = leaseExpiresAt.UTC().Unix()
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "worker_id"}}, DoUpdates: clause.AssignmentColumns([]string{"epoch", "state", "capabilities_json", "last_heartbeat_unix", "lease_expires_unix", "updated_at"})}).Create(&record).Error; err != nil {
			return fmt.Errorf("upsert worker registration: %w", err)
		}
		return nil
	})
	if transactionErr != nil {
		return nil, transactionErr
	}
	return registry.toDomain(record)
}

func (registry *WorkerRegistry) TouchHeartbeat(ctx context.Context, workerID string, epoch int64, heartbeatAt time.Time, leaseExpiresAt time.Time) (*domainrealtime.Worker, error) {
	if registry == nil || registry.db == nil {
		return nil, fmt.Errorf("worker registry is not initialized")
	}
	trimmedWorkerID := strings.TrimSpace(workerID)
	if trimmedWorkerID == "" {
		return nil, fmt.Errorf("worker_id is required")
	}
	record := workerRegistryRecord{}
	err := registry.db.WithContext(ctx).Where("worker_id = ?", trimmedWorkerID).Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("worker %q is not registered", trimmedWorkerID)
		}
		return nil, fmt.Errorf("load worker registration: %w", err)
	}
	if epoch > 0 && record.Epoch != epoch {
		return nil, fmt.Errorf("%w: expected=%d actual=%d", applicationworker.ErrEpochMismatch, epoch, record.Epoch)
	}
	record.LastHeartbeatUnix = heartbeatAt.UTC().Unix()
	record.LeaseExpiresUnix = leaseExpiresAt.UTC().Unix()
	record.State = string(domainrealtime.StateHealthy)
	if err := registry.db.WithContext(ctx).Model(&workerRegistryRecord{}).Where("worker_id = ?", trimmedWorkerID).Updates(map[string]any{
		"state":               record.State,
		"last_heartbeat_unix": record.LastHeartbeatUnix,
		"lease_expires_unix":  record.LeaseExpiresUnix,
		"updated_at":          heartbeatAt.UTC(),
	}).Error; err != nil {
		return nil, fmt.Errorf("update worker heartbeat: %w", err)
	}
	return registry.toDomain(record)
}

func (registry *WorkerRegistry) UpdateState(ctx context.Context, workerID string, epoch int64, state domainrealtime.State, changedAt time.Time) (*domainrealtime.Worker, error) {
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
	if err := registry.db.WithContext(ctx).Model(&workerRegistryRecord{}).Where("worker_id = ?", strings.TrimSpace(workerID)).Updates(map[string]any{"state": record.State, "updated_at": changedAt.UTC()}).Error; err != nil {
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

func (registry *WorkerRegistry) ListWorkers(ctx context.Context, limit int) ([]domainrealtime.Worker, error) {
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
	workers := make([]domainrealtime.Worker, 0, len(records))
	for _, record := range records {
		worker, err := registry.toDomain(record)
		if err != nil {
			return nil, err
		}
		workers = append(workers, *worker)
	}
	return workers, nil
}

func (registry *WorkerRegistry) GetSettings(ctx context.Context) (domainrealtime.Settings, error) {
	if registry == nil || registry.db == nil {
		return domainrealtime.Settings{}, fmt.Errorf("worker registry is not initialized")
	}
	record := workerRegistrySettingsRecord{}
	err := registry.db.WithContext(ctx).Where("id = 1").Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainrealtime.Settings{}, applicationworker.ErrSettingsNotFound
	} else if err != nil {
		return domainrealtime.Settings{}, fmt.Errorf("load worker settings: %w", err)
	}
	settings := domainrealtime.Settings{HeartbeatInterval: time.Duration(record.HeartbeatIntervalSeconds) * time.Second, ResponseDeadline: time.Duration(record.ResponseDeadlineSeconds) * time.Second, UpdatedAt: time.Unix(record.UpdatedAtUnix, 0).UTC()}
	if err := settings.Validate(); err != nil {
		return domainrealtime.Settings{}, err
	}
	return settings, nil
}

func (registry *WorkerRegistry) UpsertSettings(ctx context.Context, settings domainrealtime.Settings) (domainrealtime.Settings, error) {
	if registry == nil || registry.db == nil {
		return domainrealtime.Settings{}, fmt.Errorf("worker registry is not initialized")
	}
	if err := settings.Validate(); err != nil {
		return domainrealtime.Settings{}, err
	}
	record := workerRegistrySettingsRecord{ID: 1, HeartbeatIntervalSeconds: int64(settings.HeartbeatInterval.Seconds()), ResponseDeadlineSeconds: int64(settings.ResponseDeadline.Seconds()), UpdatedAtUnix: settings.UpdatedAt.UTC().Unix()}
	if err := registry.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"heartbeat_interval_seconds", "response_deadline_seconds", "updated_at_unix", "updated_at"})}).Create(&record).Error; err != nil {
		return domainrealtime.Settings{}, fmt.Errorf("upsert worker settings: %w", err)
	}
	return settings, nil
}

func (registry *WorkerRegistry) CreateRegistrationSubmission(ctx context.Context, submission domainrealtime.RegistrationSubmission) (domainrealtime.RegistrationSubmission, error) {
	if registry == nil || registry.db == nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("worker registry is not initialized")
	}
	if err := submission.Validate(); err != nil {
		return domainrealtime.RegistrationSubmission{}, err
	}
	capabilitiesJSON, err := json.Marshal(submission.Capabilities)
	if err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("encode submission capabilities: %w", err)
	}
	reasonsJSON, err := json.Marshal(submission.RejectReasons)
	if err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("encode submission reasons: %w", err)
	}
	record := workerRegistrationSubmissionRecord{
		SubmissionID:     submission.SubmissionID,
		WorkerID:         submission.WorkerID,
		RequestedAtUnix:  submission.RequestedAt.UTC().Unix(),
		ExpiresAtUnix:    submission.ExpiresAt.UTC().Unix(),
		Status:           string(submission.Status),
		CapabilitiesJSON: capabilitiesJSON,
		ReasonsJSON:      reasonsJSON,
		ResolvedAtUnix:   0,
	}
	if err := registry.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "submission_id"}}, DoNothing: true}).Create(&record).Error; err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("create registration submission: %w", err)
	}
	return submission, nil
}

func (registry *WorkerRegistry) GetRegistrationSubmission(ctx context.Context, submissionID string) (domainrealtime.RegistrationSubmission, error) {
	if registry == nil || registry.db == nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("worker registry is not initialized")
	}
	trimmedSubmissionID := strings.TrimSpace(submissionID)
	if trimmedSubmissionID == "" {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("submission_id is required")
	}
	record := workerRegistrationSubmissionRecord{}
	if err := registry.db.WithContext(ctx).Where("submission_id = ?", trimmedSubmissionID).Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainrealtime.RegistrationSubmission{}, fmt.Errorf("%w: %q", applicationworker.ErrSubmissionNotFound, trimmedSubmissionID)
		}
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("load registration submission: %w", err)
	}
	return submissionToDomain(record)
}

func (registry *WorkerRegistry) ListPendingRegistrationSubmissions(ctx context.Context, limit int) ([]domainrealtime.RegistrationSubmission, error) {
	if registry == nil || registry.db == nil {
		return nil, fmt.Errorf("worker registry is not initialized")
	}
	if limit <= 0 {
		limit = 200
	}
	records := make([]workerRegistrationSubmissionRecord, 0)
	if err := registry.db.WithContext(ctx).Where("status = ?", string(domainrealtime.RegistrationStatusPending)).Order("requested_at_unix ASC").Limit(limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list pending registration submissions: %w", err)
	}
	result := make([]domainrealtime.RegistrationSubmission, 0, len(records))
	for _, record := range records {
		submission, err := submissionToDomain(record)
		if err != nil {
			return nil, err
		}
		result = append(result, submission)
	}
	return result, nil
}

func (registry *WorkerRegistry) ResolveRegistrationSubmission(ctx context.Context, submissionID string, status domainrealtime.RegistrationStatus, reasons []string, resolvedAt time.Time) (domainrealtime.RegistrationSubmission, error) {
	if registry == nil || registry.db == nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("worker registry is not initialized")
	}
	if status != domainrealtime.RegistrationStatusAccepted && status != domainrealtime.RegistrationStatusRejected {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("resolved status must be accepted or rejected")
	}
	trimmedSubmissionID := strings.TrimSpace(submissionID)
	if trimmedSubmissionID == "" {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("submission_id is required")
	}
	reasonsJSON, err := json.Marshal(reasons)
	if err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("encode resolve reasons: %w", err)
	}
	if err := registry.db.WithContext(ctx).Model(&workerRegistrationSubmissionRecord{}).Where("submission_id = ? AND status = ?", trimmedSubmissionID, string(domainrealtime.RegistrationStatusPending)).Updates(map[string]any{
		"status":           string(status),
		"reasons_json":     reasonsJSON,
		"resolved_at_unix": resolvedAt.UTC().Unix(),
		"updated_at":       resolvedAt.UTC(),
	}).Error; err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("resolve registration submission: %w", err)
	}
	record := workerRegistrationSubmissionRecord{}
	if err := registry.db.WithContext(ctx).Where("submission_id = ?", trimmedSubmissionID).Take(&record).Error; err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("load resolved registration submission: %w", err)
	}
	return submissionToDomain(record)
}

func (registry *WorkerRegistry) RevokeRegistrationSubmission(ctx context.Context, submissionID string, reason string, revokedAt time.Time) (domainrealtime.RegistrationSubmission, error) {
	if registry == nil || registry.db == nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("worker registry is not initialized")
	}
	trimmedSubmissionID := strings.TrimSpace(submissionID)
	if trimmedSubmissionID == "" {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("submission_id is required")
	}
	if strings.TrimSpace(reason) == "" {
		reason = "revoked"
	}
	reasonsJSON, err := json.Marshal([]string{reason})
	if err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("encode revoke reason: %w", err)
	}
	if err := registry.db.WithContext(ctx).Model(&workerRegistrationSubmissionRecord{}).Where("submission_id = ? AND status = ?", trimmedSubmissionID, string(domainrealtime.RegistrationStatusPending)).Updates(map[string]any{
		"status":           string(domainrealtime.RegistrationStatusRevoked),
		"reasons_json":     reasonsJSON,
		"resolved_at_unix": revokedAt.UTC().Unix(),
		"updated_at":       revokedAt.UTC(),
	}).Error; err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("revoke registration submission: %w", err)
	}
	record := workerRegistrationSubmissionRecord{}
	if err := registry.db.WithContext(ctx).Where("submission_id = ?", trimmedSubmissionID).Take(&record).Error; err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("load revoked registration submission: %w", err)
	}
	return submissionToDomain(record)
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

func (registry *WorkerRegistry) toDomain(record workerRegistryRecord) (*domainrealtime.Worker, error) {
	capabilities, err := decodeCapabilities(record.CapabilitiesJSON)
	if err != nil {
		return nil, err
	}
	worker := &domainrealtime.Worker{
		WorkerID:       strings.TrimSpace(record.WorkerID),
		Epoch:          record.Epoch,
		State:          domainrealtime.State(strings.TrimSpace(record.State)),
		Capabilities:   capabilities,
		LastHeartbeat:  time.Unix(record.LastHeartbeatUnix, 0).UTC(),
		LeaseExpiresAt: time.Unix(record.LeaseExpiresUnix, 0).UTC(),
		UpdatedAt:      record.UpdatedAt.UTC(),
	}
	if err := worker.Validate(); err != nil {
		return nil, err
	}
	return worker, nil
}

func submissionToDomain(record workerRegistrationSubmissionRecord) (domainrealtime.RegistrationSubmission, error) {
	capabilities := make([]domainrealtime.Capability, 0)
	if err := json.Unmarshal(record.CapabilitiesJSON, &capabilities); err != nil {
		return domainrealtime.RegistrationSubmission{}, fmt.Errorf("decode submission capabilities: %w", err)
	}
	reasons := make([]string, 0)
	if len(record.ReasonsJSON) > 0 {
		if err := json.Unmarshal(record.ReasonsJSON, &reasons); err != nil {
			return domainrealtime.RegistrationSubmission{}, fmt.Errorf("decode submission reasons: %w", err)
		}
	}
	submission := domainrealtime.RegistrationSubmission{
		SubmissionID:  strings.TrimSpace(record.SubmissionID),
		WorkerID:      strings.TrimSpace(record.WorkerID),
		RequestedAt:   time.Unix(record.RequestedAtUnix, 0).UTC(),
		ExpiresAt:     time.Unix(record.ExpiresAtUnix, 0).UTC(),
		Status:        domainrealtime.RegistrationStatus(strings.TrimSpace(record.Status)),
		Capabilities:  capabilities,
		RejectReasons: reasons,
	}
	if record.ResolvedAtUnix > 0 {
		submission.ResolvedAt = time.Unix(record.ResolvedAtUnix, 0).UTC()
	}
	if err := submission.Validate(); err != nil {
		return domainrealtime.RegistrationSubmission{}, err
	}
	return submission, nil
}

var _ applicationworker.Repository = (*WorkerRegistry)(nil)
