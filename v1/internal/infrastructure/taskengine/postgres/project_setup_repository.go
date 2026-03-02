package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type projectSetupRecord struct {
	gorm.Model
	ProjectID       string `gorm:"column:project_id;size:255;not null;uniqueIndex"`
	ProjectName     string `gorm:"column:project_name;size:255;not null"`
	SCMProvider     string `gorm:"column:scm_provider;size:64;not null"`
	RepositoryURL   string `gorm:"column:repository_url;size:1024;not null"`
	TrackerProvider string `gorm:"column:tracker_provider;size:64;not null"`
	TrackerLocation string `gorm:"column:tracker_location;size:1024"`
	TrackerBoardID  string `gorm:"column:tracker_board_id;size:255"`
}

func (projectSetupRecord) TableName() string {
	return "project_setups"
}

type ProjectSetupRepository struct {
	db *gorm.DB
}

func NewProjectSetupRepository(db *gorm.DB) (*ProjectSetupRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("project setup repository db is required")
	}
	if err := db.AutoMigrate(&projectSetupRecord{}); err != nil {
		return nil, fmt.Errorf("project setup repository migrate: %w", err)
	}
	return &ProjectSetupRepository{db: db}, nil
}

func (repository *ProjectSetupRepository) ListProjectSetups(ctx context.Context, limit int) ([]applicationcontrolplane.ProjectSetup, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project setup repository is not initialized")
	}
	if limit <= 0 {
		limit = 50
	}
	records := make([]projectSetupRecord, 0)
	if err := repository.db.WithContext(ctx).Model(&projectSetupRecord{}).Order("updated_at DESC").Limit(limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list project setups: %w", err)
	}
	result := make([]applicationcontrolplane.ProjectSetup, 0, len(records))
	for _, record := range records {
		result = append(result, mapProjectSetupRecord(record))
	}
	return result, nil
}

func (repository *ProjectSetupRepository) GetProjectSetup(ctx context.Context, projectID string) (*applicationcontrolplane.ProjectSetup, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project setup repository is not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	var record projectSetupRecord
	err := repository.db.WithContext(ctx).Model(&projectSetupRecord{}).Where("project_id = ?", projectID).Take(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get project setup: %w", err)
	}
	mapped := mapProjectSetupRecord(record)
	return &mapped, nil
}

func (repository *ProjectSetupRepository) UpsertProjectSetup(ctx context.Context, setup applicationcontrolplane.ProjectSetup) (*applicationcontrolplane.ProjectSetup, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project setup repository is not initialized")
	}
	record := projectSetupRecord{
		ProjectID:       strings.TrimSpace(setup.ProjectID),
		ProjectName:     strings.TrimSpace(setup.ProjectName),
		SCMProvider:     strings.TrimSpace(setup.SCMProvider),
		RepositoryURL:   strings.TrimSpace(setup.RepositoryURL),
		TrackerProvider: strings.TrimSpace(setup.TrackerProvider),
		TrackerLocation: strings.TrimSpace(setup.TrackerLocation),
		TrackerBoardID:  strings.TrimSpace(setup.TrackerBoardID),
	}
	if err := repository.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"project_name", "scm_provider", "repository_url", "tracker_provider", "tracker_location", "tracker_board_id", "updated_at"}),
	}).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("upsert project setup: %w", err)
	}
	stored, err := repository.GetProjectSetup(ctx, record.ProjectID)
	if err != nil {
		return nil, err
	}
	return stored, nil
}

func (repository *ProjectSetupRepository) DeleteProjectSetup(ctx context.Context, projectID string) error {
	if repository == nil || repository.db == nil {
		return fmt.Errorf("project setup repository is not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	result := repository.db.WithContext(ctx).Where("project_id = ?", projectID).Delete(&projectSetupRecord{})
	if result.Error != nil {
		return fmt.Errorf("delete project setup: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("project setup not found")
	}
	return nil
}

func mapProjectSetupRecord(record projectSetupRecord) applicationcontrolplane.ProjectSetup {
	return applicationcontrolplane.ProjectSetup{
		ProjectID:       record.ProjectID,
		ProjectName:     record.ProjectName,
		SCMProvider:     record.SCMProvider,
		RepositoryURL:   record.RepositoryURL,
		TrackerProvider: record.TrackerProvider,
		TrackerLocation: record.TrackerLocation,
		TrackerBoardID:  record.TrackerBoardID,
		CreatedAt:       record.CreatedAt.UTC(),
		UpdatedAt:       record.UpdatedAt.UTC(),
	}
}

var _ applicationcontrolplane.ProjectSetupRepository = (*ProjectSetupRepository)(nil)
