package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type projectBoardRecord struct {
	gorm.Model
	ID        string    `gorm:"column:id;size:255;primaryKey"`
	ProjectID string    `gorm:"column:project_id;size:255;not null;index:idx_project_boards_project_id"`
	Name      string    `gorm:"column:name;not null"`
	State     string    `gorm:"column:state;size:64;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;index:idx_project_boards_updated_at"`
}

func (projectBoardRecord) TableName() string { return "project_boards" }

type projectBoardEpicRecord struct {
	gorm.Model
	ID               string         `gorm:"column:id;size:255;primaryKey"`
	BoardID          string         `gorm:"column:board_id;size:255;not null;index:idx_project_board_epics_board_id"`
	Title            string         `gorm:"column:title;not null"`
	Objective        string         `gorm:"column:objective"`
	RepositoryIDs    pq.StringArray `gorm:"column:repository_ids;type:text[];not null;default:'{}'"`
	Deliverables     pq.StringArray `gorm:"column:deliverables;type:text[];not null;default:'{}'"`
	State            string         `gorm:"column:state;size:64;not null;index:idx_project_board_epics_state"`
	Rank             int            `gorm:"column:rank;not null"`
	DependsOnEpicIDs pq.StringArray `gorm:"column:depends_on_epic_ids;type:text[];not null;default:'{}'"`
	CreatedAt        time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;not null"`
}

func (projectBoardEpicRecord) TableName() string { return "project_board_epics" }

type projectBoardIngestionDetailsRecord struct {
	gorm.Model
	BoardID    string    `gorm:"column:board_id;size:255;primaryKey"`
	UserPrompt string    `gorm:"column:user_prompt;type:text"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null"`
}

func (projectBoardIngestionDetailsRecord) TableName() string { return "project_board_ingestion_details" }

type projectBoardIngestionFileRecord struct {
	gorm.Model
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	BoardID   string    `gorm:"column:board_id;size:255;not null;index:idx_project_board_ingestion_files_board_pos,priority:1"`
	Position  int       `gorm:"column:position;not null;index:idx_project_board_ingestion_files_board_pos,priority:2"`
	FilePath  string    `gorm:"column:file_path;type:text;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (projectBoardIngestionFileRecord) TableName() string { return "project_board_ingestion_files" }

type projectBoardIngestionAuditRecord struct {
	gorm.Model
	ID                uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	BoardID           string     `gorm:"column:board_id;size:255;not null;index:idx_project_board_ingestion_audits_board_pos,priority:1"`
	Position          int        `gorm:"column:position;not null;index:idx_project_board_ingestion_audits_board_pos,priority:2"`
	ModelProvider     string     `gorm:"column:model_provider;not null"`
	ModelName         string     `gorm:"column:model_name;not null"`
	ModelVersion      string     `gorm:"column:model_version"`
	ModelRunID        string     `gorm:"column:model_run_id;index:idx_project_board_ingestion_audits_model_run_id"`
	AgentSessionID    string     `gorm:"column:agent_session_id"`
	AgentStreamID     string     `gorm:"column:agent_stream_id"`
	PromptFingerprint string     `gorm:"column:prompt_fingerprint"`
	InputTokens       *int       `gorm:"column:input_tokens"`
	OutputTokens      *int       `gorm:"column:output_tokens"`
	StartedAt         *time.Time `gorm:"column:started_at"`
	CompletedAt       *time.Time `gorm:"column:completed_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
}

func (projectBoardIngestionAuditRecord) TableName() string { return "project_board_ingestion_audits" }

type projectBoardTaskRecord struct {
	gorm.Model
	ID               string         `gorm:"column:id;size:255;primaryKey"`
	BoardID          string         `gorm:"column:board_id;size:255;not null;index:idx_project_board_tasks_board_id"`
	EpicID           string         `gorm:"column:epic_id;size:255;not null;index:idx_project_board_tasks_epic_id"`
	Title            string         `gorm:"column:title;not null"`
	Description      string         `gorm:"column:description"`
	RepositoryIDs    pq.StringArray `gorm:"column:repository_ids;type:text[];not null;default:'{}'"`
	Deliverables     pq.StringArray `gorm:"column:deliverables;type:text[];not null;default:'{}'"`
	TaskType         string         `gorm:"column:task_type;not null"`
	State            string         `gorm:"column:state;size:64;not null;index:idx_project_board_tasks_state"`
	Rank             int            `gorm:"column:rank;not null"`
	DependsOnTaskIDs pq.StringArray `gorm:"column:depends_on_task_ids;type:text[];not null;default:'{}'"`

	ClaimedByAgentID string     `gorm:"column:claimed_by_agent_id"`
	ClaimedAt        *time.Time `gorm:"column:claimed_at"`
	ClaimExpiresAt   *time.Time `gorm:"column:claim_expires_at;index:idx_project_board_tasks_claim_expiry"`
	ClaimToken       string     `gorm:"column:claim_token;size:255"`
	AttemptCount     int        `gorm:"column:attempt_count;not null;default:0"`
	CompletedAt      *time.Time `gorm:"column:completed_at"`

	OutcomeStatus       string `gorm:"column:outcome_status;size:64;not null;index:idx_project_board_tasks_outcome_status"`
	OutcomeSummary      string `gorm:"column:outcome_summary;not null"`
	OutcomeErrorCode    string `gorm:"column:outcome_error_code"`
	OutcomeErrorMessage string `gorm:"column:outcome_error_message"`

	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (projectBoardTaskRecord) TableName() string { return "project_board_tasks" }

type projectBoardTaskAuditRecord struct {
	gorm.Model
	ID                uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	BoardID           string     `gorm:"column:board_id;size:255;not null;index:idx_project_board_task_audits_board_task_pos,priority:1"`
	TaskID            string     `gorm:"column:task_id;size:255;not null;index:idx_project_board_task_audits_board_task_pos,priority:2"`
	Position          int        `gorm:"column:position;not null;index:idx_project_board_task_audits_board_task_pos,priority:3"`
	ModelProvider     string     `gorm:"column:model_provider;not null"`
	ModelName         string     `gorm:"column:model_name;not null"`
	ModelVersion      string     `gorm:"column:model_version"`
	ModelRunID        string     `gorm:"column:model_run_id;index:idx_project_board_task_audits_model_run_id"`
	AgentSessionID    string     `gorm:"column:agent_session_id"`
	AgentStreamID     string     `gorm:"column:agent_stream_id"`
	PromptFingerprint string     `gorm:"column:prompt_fingerprint"`
	InputTokens       *int       `gorm:"column:input_tokens"`
	OutputTokens      *int       `gorm:"column:output_tokens"`
	StartedAt         *time.Time `gorm:"column:started_at"`
	CompletedAt       *time.Time `gorm:"column:completed_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null"`
}

func (projectBoardTaskAuditRecord) TableName() string { return "project_board_task_audits" }

type PostgresBoardStore struct{ db *gorm.DB }

func NewPostgresBoardStore(db *gorm.DB) (*PostgresBoardStore, error) {
	if db == nil {
		return nil, failures.WrapTerminal(errors.New("postgres board store db is required"))
	}
	if err := db.AutoMigrate(&projectBoardRecord{}, &projectBoardEpicRecord{}, &projectBoardTaskRecord{}, &projectBoardIngestionDetailsRecord{}, &projectBoardIngestionFileRecord{}, &projectBoardIngestionAuditRecord{}, &projectBoardTaskAuditRecord{}); err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("migrate project board tables: %w", err))
	}
	return &PostgresBoardStore{db: db}, nil
}

func (store *PostgresBoardStore) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	if store == nil || store.db == nil {
		return failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	if err := board.Validate(); err != nil {
		return err
	}
	return store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		projectID := strings.TrimSpace(board.ProjectID)
		if projectID == "" {
			projectID = strings.TrimSpace(board.RunID)
		}
		boardRecord := projectBoardRecord{ID: strings.TrimSpace(board.BoardID), ProjectID: projectID, Name: strings.TrimSpace(board.Name), State: string(board.State), CreatedAt: safeTime(board.CreatedAt, now), UpdatedAt: now}
		if boardRecord.Name == "" {
			boardRecord.Name = boardRecord.ID
		}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.Assignments(map[string]any{"project_id": boardRecord.ProjectID, "name": boardRecord.Name, "state": boardRecord.State, "updated_at": now})}).Create(&boardRecord).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("upsert project board: %w", err))
		}
		if err := tx.Where("board_id = ?", boardRecord.ID).Delete(&projectBoardTaskAuditRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete existing board task audits: %w", err))
		}
		if err := tx.Where("board_id = ?", boardRecord.ID).Delete(&projectBoardIngestionAuditRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete existing board ingestion audits: %w", err))
		}
		if err := tx.Where("board_id = ?", boardRecord.ID).Delete(&projectBoardIngestionFileRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete existing board ingestion files: %w", err))
		}
		if err := tx.Where("board_id = ?", boardRecord.ID).Delete(&projectBoardIngestionDetailsRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete existing board ingestion details: %w", err))
		}
		if err := tx.Where("board_id = ?", boardRecord.ID).Delete(&projectBoardTaskRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete existing board tasks: %w", err))
		}
		if err := tx.Where("board_id = ?", boardRecord.ID).Delete(&projectBoardEpicRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete existing board epics: %w", err))
		}
		epicRecords := make([]projectBoardEpicRecord, 0, len(board.Epics))
		taskRecords := make([]projectBoardTaskRecord, 0)
		taskAuditRecords := make([]projectBoardTaskAuditRecord, 0)
		for _, epic := range board.Epics {
			epicRecord := projectBoardEpicRecord{ID: strings.TrimSpace(string(epic.ID)), BoardID: boardRecord.ID, Title: strings.TrimSpace(epic.Title), Objective: strings.TrimSpace(epic.Objective), RepositoryIDs: pq.StringArray(normalizeStringSlice(epic.RepositoryIDs)), Deliverables: pq.StringArray(normalizeStringSlice(epic.Deliverables)), State: string(epic.State), Rank: epic.Rank, DependsOnEpicIDs: pq.StringArray(workItemIDsToStrings(epic.DependsOnEpicIDs)), CreatedAt: safeTime(epic.CreatedAt, now), UpdatedAt: now}
			epicRecords = append(epicRecords, epicRecord)
			for _, task := range epic.Tasks {
				taskRecord := projectBoardTaskRecord{ID: strings.TrimSpace(string(task.ID)), BoardID: boardRecord.ID, EpicID: epicRecord.ID, Title: strings.TrimSpace(task.Title), Description: strings.TrimSpace(task.Description), RepositoryIDs: pq.StringArray(normalizeStringSlice(task.RepositoryIDs)), Deliverables: pq.StringArray(normalizeStringSlice(task.Deliverables)), TaskType: strings.TrimSpace(task.TaskType), State: string(task.State), Rank: task.Rank, DependsOnTaskIDs: pq.StringArray(workItemIDsToStrings(task.DependsOnTaskIDs)), ClaimedByAgentID: strings.TrimSpace(task.ClaimedByAgentID), ClaimedAt: task.ClaimedAt, ClaimExpiresAt: task.ClaimExpiresAt, ClaimToken: strings.TrimSpace(task.ClaimToken), AttemptCount: task.AttemptCount, OutcomeStatus: outcomeStatus(task), OutcomeSummary: outcomeSummary(task), OutcomeErrorCode: outcomeErrorCode(task), OutcomeErrorMessage: outcomeErrorMessage(task), CreatedAt: safeTime(task.CreatedAt, now), UpdatedAt: now}
				taskRecords = append(taskRecords, taskRecord)
				for auditIndex, audit := range task.Audits {
					taskAuditRecords = append(taskAuditRecords, taskModelAuditToTaskRecord(boardRecord.ID, taskRecord.ID, auditIndex, audit, now))
				}
			}
		}
		if len(epicRecords) > 0 {
			if err := tx.Create(&epicRecords).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("insert board epics: %w", err))
			}
		}
		if len(taskRecords) > 0 {
			if err := tx.Create(&taskRecords).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("insert board tasks: %w", err))
			}
		}
		if len(taskAuditRecords) > 0 {
			if err := tx.Create(&taskAuditRecords).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("insert board task audits: %w", err))
			}
		}

		if board.IngestionDetails != nil {
			ingestionDetailsRecord := projectBoardIngestionDetailsRecord{BoardID: boardRecord.ID, UserPrompt: strings.TrimSpace(board.IngestionDetails.UserPrompt), CreatedAt: now, UpdatedAt: now}
			if err := tx.Create(&ingestionDetailsRecord).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("insert board ingestion details: %w", err))
			}
			if len(board.IngestionDetails.FilesAdded) > 0 {
				fileRecords := make([]projectBoardIngestionFileRecord, 0, len(board.IngestionDetails.FilesAdded))
				for fileIndex, file := range normalizeStringSlice(board.IngestionDetails.FilesAdded) {
					fileRecords = append(fileRecords, projectBoardIngestionFileRecord{BoardID: boardRecord.ID, Position: fileIndex, FilePath: strings.TrimSpace(file), CreatedAt: now, UpdatedAt: now})
				}
				if len(fileRecords) > 0 {
					if err := tx.Create(&fileRecords).Error; err != nil {
						return failures.WrapTransient(fmt.Errorf("insert board ingestion files: %w", err))
					}
				}
			}
		}

		if len(board.IngestionAudits) > 0 {
			ingestionAuditRecords := make([]projectBoardIngestionAuditRecord, 0, len(board.IngestionAudits))
			for auditIndex, audit := range board.IngestionAudits {
				ingestionAuditRecords = append(ingestionAuditRecords, taskModelAuditToIngestionRecord(boardRecord.ID, auditIndex, audit, now))
			}
			if err := tx.Create(&ingestionAuditRecords).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("insert board ingestion audits: %w", err))
			}
		}
		return nil
	})
}

func (store *PostgresBoardStore) ListBoards(ctx context.Context, projectID string) ([]domaintracker.Board, error) {
	if store == nil || store.db == nil {
		return nil, failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	if cleanProjectID == "" {
		return nil, failures.WrapTerminal(errors.New("project_id is required"))
	}
	records := make([]projectBoardRecord, 0)
	if err := store.db.WithContext(ctx).Where("project_id = ?", cleanProjectID).Order("updated_at desc").Find(&records).Error; err != nil {
		return nil, failures.WrapTransient(fmt.Errorf("list boards: %w", err))
	}
	boards := make([]domaintracker.Board, 0, len(records))
	for _, record := range records {
		board, err := store.LoadBoard(ctx, cleanProjectID, record.ID)
		if err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	return boards, nil
}

func (store *PostgresBoardStore) DeleteBoard(ctx context.Context, projectID string, boardID string) error {
	if store == nil || store.db == nil {
		return failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanProjectID == "" || cleanBoardID == "" {
		return failures.WrapTerminal(errors.New("project_id and board_id are required"))
	}
	return store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("board_id = ?", cleanBoardID).Delete(&projectBoardTaskAuditRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete board task audits: %w", err))
		}
		if err := tx.Where("board_id = ?", cleanBoardID).Delete(&projectBoardIngestionAuditRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete board ingestion audits: %w", err))
		}
		if err := tx.Where("board_id = ?", cleanBoardID).Delete(&projectBoardIngestionFileRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete board ingestion files: %w", err))
		}
		if err := tx.Where("board_id = ?", cleanBoardID).Delete(&projectBoardIngestionDetailsRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete board ingestion details: %w", err))
		}
		if err := tx.Where("board_id = ?", cleanBoardID).Delete(&projectBoardTaskRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete board tasks: %w", err))
		}
		if err := tx.Where("board_id = ?", cleanBoardID).Delete(&projectBoardEpicRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("delete board epics: %w", err))
		}
		result := tx.Where("id = ? AND project_id = ?", cleanBoardID, cleanProjectID).Delete(&projectBoardRecord{})
		if result.Error != nil {
			return failures.WrapTransient(fmt.Errorf("delete board: %w", result.Error))
		}
		if result.RowsAffected == 0 {
			return failures.WrapTerminal(errors.New("board not found"))
		}
		return nil
	})
}

func (store *PostgresBoardStore) ClaimNextTask(ctx context.Context, projectID string, boardID string, agentID string, leaseTTL time.Duration) (domaintracker.Board, domaintracker.Task, string, error) {
	if store == nil || store.db == nil {
		return domaintracker.Board{}, domaintracker.Task{}, "", failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	cleanAgentID := strings.TrimSpace(agentID)
	if cleanProjectID == "" || cleanBoardID == "" || cleanAgentID == "" {
		return domaintracker.Board{}, domaintracker.Task{}, "", failures.WrapTerminal(errors.New("project_id, board_id and agent_id are required"))
	}
	if leaseTTL <= 0 {
		return domaintracker.Board{}, domaintracker.Task{}, "", failures.WrapTerminal(errors.New("lease_ttl must be greater than zero"))
	}

	var claimedTask projectBoardTaskRecord
	claimToken := fmt.Sprintf("claim-%s-%d", cleanAgentID, time.Now().UTC().UnixNano())
	now := time.Now().UTC()
	err := store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		candidateSQL := `
SELECT t.*
FROM project_board_tasks t
JOIN project_board_epics e ON e.id = t.epic_id AND e.board_id = t.board_id
JOIN project_boards b ON b.id = t.board_id
WHERE b.project_id = @project_id
  AND t.board_id = @board_id
  AND t.state = 'planned'
  AND (t.claim_expires_at IS NULL OR t.claim_expires_at <= @now)
  AND NOT EXISTS (
    SELECT 1
	FROM unnest(COALESCE(t.depends_on_task_ids, '{}'::text[])) dep(task_id)
    JOIN project_board_tasks td ON td.id = dep.task_id AND td.board_id = t.board_id
    WHERE td.state NOT IN ('completed', 'no_work_needed')
  )
  AND NOT EXISTS (
    SELECT 1
	FROM unnest(COALESCE(e.depends_on_epic_ids, '{}'::text[])) dep(epic_id)
    JOIN project_board_epics ed ON ed.id = dep.epic_id AND ed.board_id = e.board_id
    WHERE ed.state <> 'completed'
  )
ORDER BY e.rank ASC, t.rank ASC
FOR UPDATE SKIP LOCKED
LIMIT 1`
		if err := tx.Raw(candidateSQL, map[string]any{"project_id": cleanProjectID, "board_id": cleanBoardID, "now": now}).Scan(&claimedTask).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("select next task: %w", err))
		}
		if claimedTask.ID == "" {
			return failures.WrapTerminal(errors.New("no_task_available"))
		}
		leaseExpiry := now.Add(leaseTTL)
		updates := map[string]any{"state": string(domaintracker.TaskStateInProgress), "claimed_by_agent_id": cleanAgentID, "claimed_at": now, "claim_expires_at": leaseExpiry, "claim_token": claimToken, "attempt_count": gorm.Expr("attempt_count + 1"), "updated_at": now}
		if err := tx.Model(&projectBoardTaskRecord{}).Where("id = ? AND board_id = ?", claimedTask.ID, cleanBoardID).Updates(updates).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("claim next task: %w", err))
		}
		return nil
	})
	if err != nil {
		return domaintracker.Board{}, domaintracker.Task{}, "", err
	}
	board, loadErr := store.LoadBoard(ctx, cleanProjectID, cleanBoardID)
	if loadErr != nil {
		return domaintracker.Board{}, domaintracker.Task{}, "", loadErr
	}
	for _, epic := range board.Epics {
		for _, task := range epic.Tasks {
			if strings.TrimSpace(string(task.ID)) == claimedTask.ID {
				return board, task, claimToken, nil
			}
		}
	}
	return domaintracker.Board{}, domaintracker.Task{}, "", failures.WrapTerminal(errors.New("claimed task not found after claim"))
}

func (store *PostgresBoardStore) ApplyTaskResult(ctx context.Context, projectID string, boardID string, claimToken string, taskID string, nextState domaintracker.TaskState, outcome domaintracker.TaskOutcome) (domaintracker.Board, error) {
	if store == nil || store.db == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	cleanClaimToken := strings.TrimSpace(claimToken)
	cleanTaskID := strings.TrimSpace(taskID)
	if cleanProjectID == "" || cleanBoardID == "" || cleanClaimToken == "" || cleanTaskID == "" {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("project_id, board_id, claim_token and task_id are required"))
	}
	if err := nextState.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if err := outcome.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if err := store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var boardRecord projectBoardRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND project_id = ?", cleanBoardID, cleanProjectID).Take(&boardRecord).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return failures.WrapTerminal(errors.New("board not found"))
			}
			return failures.WrapTransient(fmt.Errorf("load board: %w", err))
		}
		now := time.Now().UTC()
		result := tx.Model(&projectBoardTaskRecord{}).
			Where("id = ? AND board_id = ? AND claim_token = ?", cleanTaskID, cleanBoardID, cleanClaimToken).
			Updates(map[string]any{"state": string(nextState), "outcome_status": string(outcome.Status), "outcome_summary": strings.TrimSpace(outcome.Summary), "outcome_error_code": strings.TrimSpace(outcome.ErrorCode), "outcome_error_message": strings.TrimSpace(outcome.ErrorMessage), "claim_token": "", "claimed_by_agent_id": "", "claimed_at": nil, "claim_expires_at": nil, "completed_at": now, "updated_at": now})
		if result.Error != nil {
			return failures.WrapTransient(fmt.Errorf("apply task result: %w", result.Error))
		}
		if result.RowsAffected == 0 {
			return failures.WrapTerminal(errors.New("task claim mismatch or task not found"))
		}
		return nil
	}); err != nil {
		return domaintracker.Board{}, err
	}
	return store.LoadBoard(ctx, cleanProjectID, cleanBoardID)
}

func (store *PostgresBoardStore) LoadBoard(ctx context.Context, projectID string, boardID string) (domaintracker.Board, error) {
	if store == nil || store.db == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanProjectID == "" || cleanBoardID == "" {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("project_id and board_id are required"))
	}
	var boardRecord projectBoardRecord
	if err := store.db.WithContext(ctx).Where("id = ? AND project_id = ?", cleanBoardID, cleanProjectID).Take(&boardRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domaintracker.Board{}, failures.WrapTerminal(errors.New("board not found"))
		}
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load board: %w", err))
	}
	var epicRecords []projectBoardEpicRecord
	if err := store.db.WithContext(ctx).Where("board_id = ?", cleanBoardID).Order("rank asc").Find(&epicRecords).Error; err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load epics: %w", err))
	}
	var taskRecords []projectBoardTaskRecord
	if err := store.db.WithContext(ctx).Where("board_id = ?", cleanBoardID).Order("epic_id asc, rank asc").Find(&taskRecords).Error; err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load tasks: %w", err))
	}
	var taskAuditRecords []projectBoardTaskAuditRecord
	if err := store.db.WithContext(ctx).Where("board_id = ?", cleanBoardID).Order("task_id asc, position asc").Find(&taskAuditRecords).Error; err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load task audits: %w", err))
	}
	taskAuditsByTask := map[string][]domaintracker.TaskModelAudit{}
	for _, rec := range taskAuditRecords {
		taskAuditsByTask[rec.TaskID] = append(taskAuditsByTask[rec.TaskID], mapAuditRecord(rec))
	}
	tasksByEpic := map[string][]domaintracker.Task{}
	for _, rec := range taskRecords {
		task := mapTaskRecord(rec, taskAuditsByTask[rec.ID])
		tasksByEpic[rec.EpicID] = append(tasksByEpic[rec.EpicID], task)
	}
	epics := make([]domaintracker.Epic, 0, len(epicRecords))
	for _, rec := range epicRecords {
		epics = append(epics, domaintracker.Epic{ID: domaintracker.WorkItemID(rec.ID), BoardID: rec.BoardID, Title: rec.Title, Objective: rec.Objective, RepositoryIDs: normalizeStringSlice([]string(rec.RepositoryIDs)), Deliverables: normalizeStringSlice([]string(rec.Deliverables)), State: domaintracker.EpicState(rec.State), Rank: rec.Rank, DependsOnEpicIDs: stringsToWorkItemIDs([]string(rec.DependsOnEpicIDs)), Tasks: tasksByEpic[rec.ID], CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt})
	}
	var ingestionDetailsRecord projectBoardIngestionDetailsRecord
	ingestionDetails := (*domaintracker.BoardIngestionDetails)(nil)
	if err := store.db.WithContext(ctx).Where("board_id = ?", cleanBoardID).Take(&ingestionDetailsRecord).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load board ingestion details: %w", err))
		}
	} else {
		ingestionDetails = &domaintracker.BoardIngestionDetails{UserPrompt: strings.TrimSpace(ingestionDetailsRecord.UserPrompt)}
		var ingestionFiles []projectBoardIngestionFileRecord
		if err := store.db.WithContext(ctx).Where("board_id = ?", cleanBoardID).Order("position asc").Find(&ingestionFiles).Error; err != nil {
			return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load board ingestion files: %w", err))
		}
		for _, file := range ingestionFiles {
			ingestionDetails.FilesAdded = append(ingestionDetails.FilesAdded, strings.TrimSpace(file.FilePath))
		}
	}
	var ingestionAuditRecords []projectBoardIngestionAuditRecord
	if err := store.db.WithContext(ctx).Where("board_id = ?", cleanBoardID).Order("position asc").Find(&ingestionAuditRecords).Error; err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load board ingestion audits: %w", err))
	}
	ingestionAudits := make([]domaintracker.TaskModelAudit, 0, len(ingestionAuditRecords))
	for _, rec := range ingestionAuditRecords {
		ingestionAudits = append(ingestionAudits, mapIngestionAuditRecord(rec))
	}
	board := domaintracker.Board{BoardID: boardRecord.ID, RunID: boardRecord.ProjectID, ProjectID: boardRecord.ProjectID, Name: boardRecord.Name, State: domaintracker.BoardState(boardRecord.State), Epics: epics, IngestionDetails: ingestionDetails, IngestionAudits: ingestionAudits, CreatedAt: boardRecord.CreatedAt, UpdatedAt: boardRecord.UpdatedAt}
	if err := board.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	return board, nil
}

func mapTaskRecord(rec projectBoardTaskRecord, audits []domaintracker.TaskModelAudit) domaintracker.Task {
	task := domaintracker.Task{ID: domaintracker.WorkItemID(rec.ID), BoardID: rec.BoardID, EpicID: domaintracker.WorkItemID(rec.EpicID), Title: rec.Title, Description: rec.Description, RepositoryIDs: normalizeStringSlice([]string(rec.RepositoryIDs)), Deliverables: normalizeStringSlice([]string(rec.Deliverables)), TaskType: rec.TaskType, State: domaintracker.TaskState(rec.State), Rank: rec.Rank, DependsOnTaskIDs: stringsToWorkItemIDs([]string(rec.DependsOnTaskIDs)), Audits: audits, ClaimedByAgentID: rec.ClaimedByAgentID, ClaimedAt: rec.ClaimedAt, ClaimExpiresAt: rec.ClaimExpiresAt, ClaimToken: rec.ClaimToken, AttemptCount: rec.AttemptCount, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt}
	if strings.TrimSpace(rec.OutcomeStatus) != "" {
		task.Outcome = &domaintracker.TaskOutcome{Status: domaintracker.OutcomeStatus(rec.OutcomeStatus), Summary: rec.OutcomeSummary, ErrorCode: rec.OutcomeErrorCode, ErrorMessage: rec.OutcomeErrorMessage}
	}
	return task
}

func taskModelAuditToTaskRecord(boardID string, taskID string, position int, audit domaintracker.TaskModelAudit, now time.Time) projectBoardTaskAuditRecord {
	return projectBoardTaskAuditRecord{BoardID: boardID, TaskID: taskID, Position: position, ModelProvider: strings.TrimSpace(audit.ModelProvider), ModelName: strings.TrimSpace(audit.ModelName), ModelVersion: strings.TrimSpace(audit.ModelVersion), ModelRunID: strings.TrimSpace(audit.ModelRunID), AgentSessionID: strings.TrimSpace(audit.AgentSessionID), AgentStreamID: strings.TrimSpace(audit.AgentStreamID), PromptFingerprint: strings.TrimSpace(audit.PromptFingerprint), InputTokens: audit.InputTokens, OutputTokens: audit.OutputTokens, StartedAt: audit.StartedAt, CompletedAt: audit.CompletedAt, CreatedAt: now, UpdatedAt: now}
}

func taskModelAuditToIngestionRecord(boardID string, position int, audit domaintracker.TaskModelAudit, now time.Time) projectBoardIngestionAuditRecord {
	return projectBoardIngestionAuditRecord{BoardID: boardID, Position: position, ModelProvider: strings.TrimSpace(audit.ModelProvider), ModelName: strings.TrimSpace(audit.ModelName), ModelVersion: strings.TrimSpace(audit.ModelVersion), ModelRunID: strings.TrimSpace(audit.ModelRunID), AgentSessionID: strings.TrimSpace(audit.AgentSessionID), AgentStreamID: strings.TrimSpace(audit.AgentStreamID), PromptFingerprint: strings.TrimSpace(audit.PromptFingerprint), InputTokens: audit.InputTokens, OutputTokens: audit.OutputTokens, StartedAt: audit.StartedAt, CompletedAt: audit.CompletedAt, CreatedAt: now, UpdatedAt: now}
}

func mapAuditRecord(rec projectBoardTaskAuditRecord) domaintracker.TaskModelAudit {
	return domaintracker.TaskModelAudit{ModelProvider: rec.ModelProvider, ModelName: rec.ModelName, ModelVersion: rec.ModelVersion, ModelRunID: rec.ModelRunID, AgentSessionID: rec.AgentSessionID, AgentStreamID: rec.AgentStreamID, PromptFingerprint: rec.PromptFingerprint, InputTokens: rec.InputTokens, OutputTokens: rec.OutputTokens, StartedAt: rec.StartedAt, CompletedAt: rec.CompletedAt}
}

func mapIngestionAuditRecord(rec projectBoardIngestionAuditRecord) domaintracker.TaskModelAudit {
	return domaintracker.TaskModelAudit{ModelProvider: rec.ModelProvider, ModelName: rec.ModelName, ModelVersion: rec.ModelVersion, ModelRunID: rec.ModelRunID, AgentSessionID: rec.AgentSessionID, AgentStreamID: rec.AgentStreamID, PromptFingerprint: rec.PromptFingerprint, InputTokens: rec.InputTokens, OutputTokens: rec.OutputTokens, StartedAt: rec.StartedAt, CompletedAt: rec.CompletedAt}
}

func workItemIDsToStrings(ids []domaintracker.WorkItemID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(string(id))
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func stringsToWorkItemIDs(values []string) []domaintracker.WorkItemID {
	result := make([]domaintracker.WorkItemID, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, domaintracker.WorkItemID(trimmed))
	}
	return result
}

func normalizeStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func safeTime(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}

func outcomeStatus(task domaintracker.Task) string {
	if task.Outcome == nil {
		return string(domaintracker.OutcomeStatusPartial)
	}
	return string(task.Outcome.Status)
}

func outcomeSummary(task domaintracker.Task) string {
	if task.Outcome == nil || strings.TrimSpace(task.Outcome.Summary) == "" {
		return "pending"
	}
	return strings.TrimSpace(task.Outcome.Summary)
}

func outcomeErrorCode(task domaintracker.Task) string {
	if task.Outcome == nil {
		return ""
	}
	return strings.TrimSpace(task.Outcome.ErrorCode)
}

func outcomeErrorMessage(task domaintracker.Task) string {
	if task.Outcome == nil {
		return ""
	}
	return strings.TrimSpace(task.Outcome.ErrorMessage)
}

var _ applicationtracker.BoardStore = (*PostgresBoardStore)(nil)
