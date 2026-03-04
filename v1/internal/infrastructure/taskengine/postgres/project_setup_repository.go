package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type projectSetupRecord struct {
	gorm.Model
	ProjectID   string `gorm:"column:project_id;size:255;not null;uniqueIndex"`
	ProjectName string `gorm:"column:project_name;size:255;not null"`
}

func (projectSetupRecord) TableName() string {
	return "project_setups"
}

type projectRepositoryRecord struct {
	gorm.Model
	ProjectID     string `gorm:"column:project_id;size:255;not null;index"`
	RepositoryID  string `gorm:"column:repository_id;size:255;not null"`
	SCMID         string `gorm:"column:scm_id;size:255;not null"`
	RepositoryURL string `gorm:"column:repository_url;size:1024;not null"`
	IsPrimary     bool   `gorm:"column:is_primary;not null;default:false"`
}

func (projectRepositoryRecord) TableName() string {
	return "project_repositories"
}

type projectSCMRecord struct {
	gorm.Model
	ProjectID   string `gorm:"column:project_id;size:255;not null;index"`
	SCMID       string `gorm:"column:scm_id;size:255;not null"`
	SCMProvider string `gorm:"column:scm_provider;size:64;not null"`
	SCMToken    string `gorm:"column:scm_token;size:512;not null"`
}

func (projectSCMRecord) TableName() string {
	return "project_scms"
}

type trackerProjectBoardRecord struct {
	ID        string    `gorm:"column:id;size:255;primaryKey"`
	ProjectID string    `gorm:"column:project_id;size:255;not null;index"`
	Name      string    `gorm:"column:name;not null"`
	State     string    `gorm:"column:state;size:64;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (trackerProjectBoardRecord) TableName() string {
	return "project_boards"
}

type trackerBoardSnapshotRecord struct {
	ID         uint      `gorm:"primaryKey"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
	RunID      string `gorm:"column:run_id;size:255;not null;uniqueIndex:idx_tracker_board_snapshot,priority:1"`
	BoardID    string `gorm:"column:board_id;size:255;not null;uniqueIndex:idx_tracker_board_snapshot,priority:2"`
	SourceKind string `gorm:"column:source_kind;size:64;not null"`
	SourceRef  string `gorm:"column:source_ref"`
	Payload    []byte `gorm:"column:payload;not null"`
}

func (trackerBoardSnapshotRecord) TableName() string {
	return "project_board_snapshots"
}

type ProjectSetupRepository struct {
	db             *gorm.DB
	scmTokenCrypto *SCMTokenCrypto
}

func NewProjectSetupRepository(db *gorm.DB, scmTokenCrypto *SCMTokenCrypto) (*ProjectSetupRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("project setup repository db is required")
	}
	if scmTokenCrypto == nil {
		return nil, fmt.Errorf("project setup repository scm token crypto is required")
	}
	if err := normalizeLegacyProjectSetupSchema(db); err != nil {
		return nil, fmt.Errorf("project setup repository normalize legacy schema: %w", err)
	}
	if err := db.AutoMigrate(&projectSetupRecord{}, &projectSCMRecord{}, &projectRepositoryRecord{}, &trackerProjectBoardRecord{}, &trackerBoardSnapshotRecord{}); err != nil {
		return nil, fmt.Errorf("project setup repository migrate: %w", err)
	}
	return &ProjectSetupRepository{db: db, scmTokenCrypto: scmTokenCrypto}, nil
}

func normalizeLegacyProjectSetupSchema(db *gorm.DB) error {
	if db == nil || db.Migrator() == nil {
		return nil
	}
	migrator := db.Migrator()
	if migrator.HasTable("tracker_board_snapshots") && !migrator.HasTable("project_board_snapshots") {
		if err := migrator.RenameTable("tracker_board_snapshots", "project_board_snapshots"); err != nil {
			return err
		}
	}
	return nil
}

func (repository *ProjectSetupRepository) RotateSCMTokenEncryptionKeys(ctx context.Context) error {
	if repository == nil || repository.db == nil || repository.scmTokenCrypto == nil {
		return fmt.Errorf("project setup repository is not initialized")
	}
	return repository.scmTokenCrypto.RotateAndReencryptSCMTokens(ctx)
}

func (repository *ProjectSetupRepository) MigrateLegacySCMTokensToEncrypted(ctx context.Context) error {
	if repository == nil || repository.db == nil || repository.scmTokenCrypto == nil {
		return fmt.Errorf("project setup repository is not initialized")
	}
	return repository.scmTokenCrypto.MigrateLegacyPlaintextSCMTokens(ctx)
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
		setup, err := repository.loadProjectSetup(ctx, record)
		if err != nil {
			return nil, err
		}
		result = append(result, setup)
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
	setup, err := repository.loadProjectSetup(ctx, record)
	if err != nil {
		return nil, err
	}
	return &setup, nil
}

func (repository *ProjectSetupRepository) UpsertProjectSetup(ctx context.Context, setup applicationcontrolplane.ProjectSetup) (*applicationcontrolplane.ProjectSetup, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project setup repository is not initialized")
	}
	projectRecord := projectSetupRecord{
		ProjectID:   strings.TrimSpace(setup.ProjectID),
		ProjectName: strings.TrimSpace(setup.ProjectName),
	}
	existingSCMByID := map[string]projectSCMRecord{}
	if projectRecord.ProjectID != "" {
		existingSCMRecords := make([]projectSCMRecord, 0)
		if err := repository.db.WithContext(ctx).
			Model(&projectSCMRecord{}).
			Where("project_id = ?", projectRecord.ProjectID).
			Find(&existingSCMRecords).Error; err != nil {
			return nil, fmt.Errorf("load existing project scms: %w", err)
		}
		for _, record := range existingSCMRecords {
			existingSCMByID[strings.TrimSpace(record.SCMID)] = record
		}
	}
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "project_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"project_name", "updated_at"}),
		}).Create(&projectRecord).Error; err != nil {
			return fmt.Errorf("upsert project setup: %w", err)
		}
		if err := tx.Where("project_id = ?", projectRecord.ProjectID).Delete(&projectSCMRecord{}).Error; err != nil {
			return fmt.Errorf("delete project scms: %w", err)
		}
		if err := tx.Where("project_id = ?", projectRecord.ProjectID).Delete(&projectRepositoryRecord{}).Error; err != nil {
			return fmt.Errorf("delete project repositories: %w", err)
		}
		if len(setup.SCMs) > 0 {
			scms := make([]projectSCMRecord, 0, len(setup.SCMs))
			for _, scmSetup := range setup.SCMs {
				scmID := strings.TrimSpace(scmSetup.SCMID)
				scmProvider := strings.TrimSpace(scmSetup.SCMProvider)
				trimmedToken := strings.TrimSpace(scmSetup.SCMToken)
				encryptedToken := ""
				if trimmedToken == "" {
					existing, ok := existingSCMByID[scmID]
					if !ok || strings.TrimSpace(existing.SCMProvider) != scmProvider {
						return fmt.Errorf("scm token is required for scm_id %q", scmID)
					}
					encryptedToken = existing.SCMToken
				} else {
					var encryptErr error
					encryptedToken, encryptErr = repository.scmTokenCrypto.Encrypt(ctx, trimmedToken)
					if encryptErr != nil {
						return fmt.Errorf("encrypt scm token for scm_id %q: %w", scmID, encryptErr)
					}
				}
				scms = append(scms, projectSCMRecord{
					ProjectID:   projectRecord.ProjectID,
					SCMID:       scmID,
					SCMProvider: scmProvider,
					SCMToken:    encryptedToken,
				})
			}
			if err := tx.Create(&scms).Error; err != nil {
				return fmt.Errorf("insert project scms: %w", err)
			}
		}
		if len(setup.Repositories) > 0 {
			repositories := make([]projectRepositoryRecord, 0, len(setup.Repositories))
			for _, repositorySetup := range setup.Repositories {
				repositories = append(repositories, projectRepositoryRecord{
					ProjectID:     projectRecord.ProjectID,
					RepositoryID:  strings.TrimSpace(repositorySetup.RepositoryID),
					SCMID:         strings.TrimSpace(repositorySetup.SCMID),
					RepositoryURL: strings.TrimSpace(repositorySetup.RepositoryURL),
					IsPrimary:     repositorySetup.IsPrimary,
				})
			}
			if err := tx.Create(&repositories).Error; err != nil {
				return fmt.Errorf("insert project repositories: %w", err)
			}
		}
		if len(setup.Boards) > 0 {
			boards := make([]trackerProjectBoardRecord, 0, len(setup.Boards))
			snapshots := make([]trackerBoardSnapshotRecord, 0)
			now := time.Now().UTC()
			for _, boardSetup := range setup.Boards {
				trimmedTrackerProvider := strings.TrimSpace(boardSetup.TrackerProvider)
				trimmedBoardID := strings.TrimSpace(boardSetup.BoardID)
				if trimmedBoardID == "" {
					continue
				}
				boardName := strings.TrimSpace(boardSetup.TaskboardName)
				if boardName == "" {
					boardName = trimmedBoardID
				}
				boards = append(boards, trackerProjectBoardRecord{
					ID:        trimmedBoardID,
					ProjectID: projectRecord.ProjectID,
					Name:      boardName,
					State:     "pending",
					CreatedAt: now,
					UpdatedAt: now,
				})
				if trimmedTrackerProvider == "internal" {
					payload, err := json.Marshal(map[string]any{
						"board_id":   trimmedBoardID,
						"run_id":     projectRecord.ProjectID,
						"title":      boardName,
						"status":     "not-started",
						"epics":      []any{},
						"created_at": now,
						"updated_at": now,
					})
					if err != nil {
						return fmt.Errorf("encode internal tracker board snapshot payload: %w", err)
					}
					snapshots = append(snapshots, trackerBoardSnapshotRecord{
						RunID:      projectRecord.ProjectID,
						BoardID:    trimmedBoardID,
						SourceKind: "internal",
						SourceRef:  trimmedBoardID,
						Payload:    payload,
					})
				}
			}
			if len(boards) > 0 {
				for _, board := range boards {
					if err := tx.Clauses(clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}},
						DoUpdates: clause.AssignmentColumns([]string{"project_id", "name", "updated_at"}),
					}).Create(&board).Error; err != nil {
						return fmt.Errorf("upsert project boards: %w", err)
					}
				}
			}
			if len(snapshots) > 0 {
				for _, snapshot := range snapshots {
					if err := tx.Clauses(clause.OnConflict{
						Columns:   []clause.Column{{Name: "run_id"}, {Name: "board_id"}},
						DoNothing: true,
					}).Create(&snapshot).Error; err != nil {
						return fmt.Errorf("insert internal tracker snapshot: %w", err)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	stored, err := repository.GetProjectSetup(ctx, projectRecord.ProjectID)
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
	result := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", projectID).Delete(&projectSCMRecord{}).Error; err != nil {
			return fmt.Errorf("delete project scms: %w", err)
		}
		if err := tx.Where("project_id = ?", projectID).Delete(&projectRepositoryRecord{}).Error; err != nil {
			return fmt.Errorf("delete project repositories: %w", err)
		}
		if err := tx.Where("project_id = ?", projectID).Delete(&trackerProjectBoardRecord{}).Error; err != nil {
			return fmt.Errorf("delete project boards: %w", err)
		}
		deleteResult := tx.Where("project_id = ?", projectID).Delete(&projectSetupRecord{})
		if deleteResult.Error != nil {
			return fmt.Errorf("delete project setup: %w", deleteResult.Error)
		}
		if deleteResult.RowsAffected == 0 {
			return fmt.Errorf("project setup not found")
		}
		return nil
	})
	if result != nil {
		return result
	}
	return nil
}

func (repository *ProjectSetupRepository) loadProjectSetup(ctx context.Context, projectRecord projectSetupRecord) (applicationcontrolplane.ProjectSetup, error) {
	repositoryRecords := make([]projectRepositoryRecord, 0)
	if err := repository.db.WithContext(ctx).Model(&projectRepositoryRecord{}).Where("project_id = ?", projectRecord.ProjectID).Order("created_at ASC").Find(&repositoryRecords).Error; err != nil {
		return applicationcontrolplane.ProjectSetup{}, fmt.Errorf("load project repositories: %w", err)
	}
	scmRecords := make([]projectSCMRecord, 0)
	if err := repository.db.WithContext(ctx).Model(&projectSCMRecord{}).Where("project_id = ?", projectRecord.ProjectID).Order("created_at ASC").Find(&scmRecords).Error; err != nil {
		return applicationcontrolplane.ProjectSetup{}, fmt.Errorf("load project scms: %w", err)
	}
	boardRecords := make([]trackerProjectBoardRecord, 0)
	boardQuery := repository.db.WithContext(ctx).Model(&trackerProjectBoardRecord{}).Where("project_id = ?", projectRecord.ProjectID)
	if repository.db.Migrator() != nil && repository.db.Migrator().HasTable("workflow_jobs") {
		legacyRunIDs := repository.db.WithContext(ctx).
			Model(&admissionLedgerRecord{}).
			Distinct("run_id").
			Where("project_id = ?", projectRecord.ProjectID)
		boardQuery = boardQuery.Or("project_id IN (?)", legacyRunIDs)
	}
	if err := boardQuery.Order("created_at ASC").Find(&boardRecords).Error; err != nil {
		return applicationcontrolplane.ProjectSetup{}, fmt.Errorf("load project boards: %w", err)
	}
	scms := make([]applicationcontrolplane.ProjectSCM, 0, len(scmRecords))
	for _, scmRecord := range scmRecords {
		decryptedToken, decryptErr := repository.scmTokenCrypto.Decrypt(ctx, strings.TrimSpace(scmRecord.SCMToken))
		if decryptErr != nil {
			return applicationcontrolplane.ProjectSetup{}, fmt.Errorf("decrypt scm token for scm_id %q: %w", strings.TrimSpace(scmRecord.SCMID), decryptErr)
		}
		scms = append(scms, applicationcontrolplane.ProjectSCM{
			SCMID:       strings.TrimSpace(scmRecord.SCMID),
			SCMProvider: strings.TrimSpace(scmRecord.SCMProvider),
			SCMToken:    strings.TrimSpace(decryptedToken),
		})
	}
	repositories := make([]applicationcontrolplane.ProjectRepository, 0, len(repositoryRecords))
	for _, repositoryRecord := range repositoryRecords {
		repositories = append(repositories, applicationcontrolplane.ProjectRepository{
			RepositoryID:  strings.TrimSpace(repositoryRecord.RepositoryID),
			SCMID:         strings.TrimSpace(repositoryRecord.SCMID),
			RepositoryURL: strings.TrimSpace(repositoryRecord.RepositoryURL),
			IsPrimary:     repositoryRecord.IsPrimary,
		})
	}
	boards := make([]applicationcontrolplane.ProjectBoard, 0, len(boardRecords))
	for _, boardRecord := range boardRecords {
		boardID := strings.TrimSpace(boardRecord.ID)
		if boardID == "" {
			continue
		}
		boardName := strings.TrimSpace(boardRecord.Name)
		if boardName == "" {
			boardName = boardID
		}
		boards = append(boards, applicationcontrolplane.ProjectBoard{
			BoardID:                  boardID,
			TrackerProvider:          "internal",
			TaskboardName:            boardName,
			AppliesToAllRepositories: true,
			RepositoryIDs:            nil,
		})
	}
	return applicationcontrolplane.ProjectSetup{
		ProjectID:    projectRecord.ProjectID,
		ProjectName:  projectRecord.ProjectName,
		SCMs:         scms,
		Repositories: repositories,
		Boards:       boards,
		CreatedAt:    projectRecord.CreatedAt.UTC(),
		UpdatedAt:    projectRecord.UpdatedAt.UTC(),
	}, nil
}

var _ applicationcontrolplane.ProjectSetupRepository = (*ProjectSetupRepository)(nil)
