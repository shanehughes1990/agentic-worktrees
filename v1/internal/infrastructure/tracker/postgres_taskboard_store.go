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
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type trackerBoardRecord struct {
	gorm.Model
	RunID           string    `gorm:"column:run_id;size:255;not null;uniqueIndex:idx_tracker_board,priority:1"`
	BoardID         string    `gorm:"column:board_id;size:255;not null;uniqueIndex:idx_tracker_board,priority:2"`
	SourceKind      string    `gorm:"column:source_kind;size:64;not null"`
	SourceLocation  string    `gorm:"column:source_location;size:512"`
	SourceBoardID   string    `gorm:"column:source_board_id;size:255"`
	Title           string    `gorm:"column:title;size:512"`
	Goal            string    `gorm:"column:goal;type:text"`
	Status          string    `gorm:"column:status;size:64;not null"`
	MetadataJSON    string    `gorm:"column:metadata_json;type:text"`
	CreatedAtSource time.Time `gorm:"column:created_at_source"`
	UpdatedAtSource time.Time `gorm:"column:updated_at_source"`
}

func (trackerBoardRecord) TableName() string {
	return "tracker_boards"
}

type trackerEpicRecord struct {
	gorm.Model
	RunID           string    `gorm:"column:run_id;size:255;not null;index:idx_tracker_epic,priority:1"`
	BoardID         string    `gorm:"column:board_id;size:255;not null;index:idx_tracker_epic,priority:2"`
	EpicID          string    `gorm:"column:epic_id;size:255;not null;uniqueIndex:idx_tracker_epic_unique,priority:1"`
	Title           string    `gorm:"column:title;size:512;not null"`
	Description     string    `gorm:"column:description;type:text"`
	Status          string    `gorm:"column:status;size:64;not null"`
	Priority        string    `gorm:"column:priority;size:64"`
	DependsOnJSON   string    `gorm:"column:depends_on_json;type:text"`
	MetadataJSON    string    `gorm:"column:metadata_json;type:text"`
	CreatedAtSource time.Time `gorm:"column:created_at_source"`
	UpdatedAtSource time.Time `gorm:"column:updated_at_source"`
}

func (trackerEpicRecord) TableName() string {
	return "tracker_epics"
}

type trackerTaskRecord struct {
	gorm.Model
	RunID           string    `gorm:"column:run_id;size:255;not null;index:idx_tracker_task,priority:1"`
	BoardID         string    `gorm:"column:board_id;size:255;not null;index:idx_tracker_task,priority:2"`
	EpicID          string    `gorm:"column:epic_id;size:255;not null;index"`
	TaskID          string    `gorm:"column:task_id;size:255;not null;uniqueIndex:idx_tracker_task_unique,priority:1"`
	Title           string    `gorm:"column:title;size:512;not null"`
	Description     string    `gorm:"column:description;type:text"`
	Status          string    `gorm:"column:status;size:64;not null"`
	Priority        string    `gorm:"column:priority;size:64"`
	DependsOnJSON   string    `gorm:"column:depends_on_json;type:text"`
	MetadataJSON    string    `gorm:"column:metadata_json;type:text"`
	CreatedAtSource time.Time `gorm:"column:created_at_source"`
	UpdatedAtSource time.Time `gorm:"column:updated_at_source"`
}

func (trackerTaskRecord) TableName() string {
	return "tracker_tasks"
}

type trackerTaskOutcomeRecord struct {
	gorm.Model
	RunID           string    `gorm:"column:run_id;size:255;not null;index:idx_tracker_task_outcome,priority:1"`
	BoardID         string    `gorm:"column:board_id;size:255;not null;index:idx_tracker_task_outcome,priority:2"`
	TaskID          string    `gorm:"column:task_id;size:255;not null;uniqueIndex:idx_tracker_task_outcome_unique,priority:1"`
	Status          string    `gorm:"column:status;size:64;not null"`
	Reason          string    `gorm:"column:reason;type:text"`
	TaskBranch      string    `gorm:"column:task_branch;size:255"`
	Worktree        string    `gorm:"column:worktree;size:1024"`
	ResumeSessionID string    `gorm:"column:resume_session_id;size:255"`
	UpdatedAtSource time.Time `gorm:"column:updated_at_source"`
}

func (trackerTaskOutcomeRecord) TableName() string {
	return "tracker_task_outcomes"
}

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

type PostgresTaskboardStore struct {
	db *gorm.DB
}

func NewPostgresTaskboardStore(db *gorm.DB) (*PostgresTaskboardStore, error) {
	if db == nil {
		return nil, failures.WrapTerminal(errors.New("postgres taskboard store db is required"))
	}
	if err := db.AutoMigrate(&boardSnapshotRecord{}, &trackerBoardRecord{}, &trackerEpicRecord{}, &trackerTaskRecord{}, &trackerTaskOutcomeRecord{}); err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("migrate tracker taskboard tables: %w", err))
	}
	return &PostgresTaskboardStore{db: db}, nil
}

func (store *PostgresTaskboardStore) LoadBoard(ctx context.Context, projectID string, boardRef string) (domaintracker.Board, error) {
	if store == nil || store.db == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("postgres taskboard store is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardRef := strings.TrimSpace(boardRef)
	if cleanProjectID == "" || cleanBoardRef == "" {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("project_id and board_ref are required"))
	}

	var record boardSnapshotRecord
	err := store.db.WithContext(ctx).
		Where("run_id = ? AND board_id = ?", cleanProjectID, cleanBoardRef).
		Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("internal tracker board %q not found for project %q", cleanBoardRef, cleanProjectID))
		}
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("load internal tracker board snapshot: %w", err))
	}

	var board domaintracker.Board
	if err := json.Unmarshal(record.Payload, &board); err != nil {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("decode internal tracker board snapshot: %w", err))
	}
	if strings.TrimSpace(board.BoardID) == "" {
		board.BoardID = cleanBoardRef
	}
	if strings.TrimSpace(board.RunID) == "" {
		board.RunID = cleanProjectID
	}
	return board, nil
}

func (store *PostgresTaskboardStore) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	if store == nil || store.db == nil {
		return failures.WrapTerminal(errors.New("postgres taskboard store is not initialized"))
	}
	if err := board.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(board)
	if err != nil {
		return failures.WrapTerminal(fmt.Errorf("encode board snapshot payload: %w", err))
	}
	boardMetadata, err := json.Marshal(board.Metadata)
	if err != nil {
		return failures.WrapTerminal(fmt.Errorf("encode board metadata: %w", err))
	}

	sourceRef := strings.TrimSpace(board.Source.BoardID)
	if sourceRef == "" {
		sourceRef = strings.TrimSpace(board.Source.Location)
	}
	if sourceRef == "" {
		sourceRef = strings.TrimSpace(board.BoardID)
	}

	boardRecord := trackerBoardRecord{
		RunID:           strings.TrimSpace(board.RunID),
		BoardID:         strings.TrimSpace(board.BoardID),
		SourceKind:      strings.TrimSpace(string(board.Source.Kind)),
		SourceLocation:  strings.TrimSpace(board.Source.Location),
		SourceBoardID:   strings.TrimSpace(board.Source.BoardID),
		Title:           strings.TrimSpace(board.Title),
		Goal:            strings.TrimSpace(board.Goal),
		Status:          strings.TrimSpace(string(board.Status)),
		MetadataJSON:    string(boardMetadata),
		CreatedAtSource: board.CreatedAt,
		UpdatedAtSource: board.UpdatedAt,
	}
	snapshot := boardSnapshotRecord{
		RunID:      strings.TrimSpace(board.RunID),
		BoardID:    strings.TrimSpace(board.BoardID),
		SourceKind: strings.TrimSpace(string(board.Source.Kind)),
		SourceRef:  sourceRef,
		Payload:    payload,
	}

	return store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "run_id"}, {Name: "board_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"source_kind", "source_ref", "payload", "updated_at"}),
		}).Create(&snapshot).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("persist board snapshot: %w", err))
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "run_id"}, {Name: "board_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"source_kind", "source_location", "source_board_id", "title", "goal", "status", "metadata_json", "created_at_source", "updated_at_source", "updated_at"}),
		}).Create(&boardRecord).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("persist normalized board: %w", err))
		}
		if err := tx.Where("run_id = ? AND board_id = ?", board.RunID, board.BoardID).Delete(&trackerEpicRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("clear normalized epics: %w", err))
		}
		if err := tx.Where("run_id = ? AND board_id = ?", board.RunID, board.BoardID).Delete(&trackerTaskRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("clear normalized tasks: %w", err))
		}
		if err := tx.Where("run_id = ? AND board_id = ?", board.RunID, board.BoardID).Delete(&trackerTaskOutcomeRecord{}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("clear normalized task outcomes: %w", err))
		}

		epicRecords := make([]trackerEpicRecord, 0)
		taskRecords := make([]trackerTaskRecord, 0)
		outcomeRecords := make([]trackerTaskOutcomeRecord, 0)
		for _, epic := range board.Epics {
			epicDependsOn, encodeErr := json.Marshal(epic.DependsOn)
			if encodeErr != nil {
				return failures.WrapTerminal(fmt.Errorf("encode epic dependencies: %w", encodeErr))
			}
			epicMetadata, encodeErr := json.Marshal(epic.Metadata)
			if encodeErr != nil {
				return failures.WrapTerminal(fmt.Errorf("encode epic metadata: %w", encodeErr))
			}
			epicRecords = append(epicRecords, trackerEpicRecord{
				RunID:           board.RunID,
				BoardID:         board.BoardID,
				EpicID:          strings.TrimSpace(string(epic.ID)),
				Title:           strings.TrimSpace(epic.Title),
				Description:     strings.TrimSpace(epic.Description),
				Status:          strings.TrimSpace(string(epic.Status)),
				Priority:        strings.TrimSpace(string(epic.Priority)),
				DependsOnJSON:   string(epicDependsOn),
				MetadataJSON:    string(epicMetadata),
				CreatedAtSource: epic.CreatedAt,
				UpdatedAtSource: epic.UpdatedAt,
			})
			for _, task := range epic.Tasks {
				taskDependsOn, encodeErr := json.Marshal(task.DependsOn)
				if encodeErr != nil {
					return failures.WrapTerminal(fmt.Errorf("encode task dependencies: %w", encodeErr))
				}
				taskMetadata, encodeErr := json.Marshal(task.Metadata)
				if encodeErr != nil {
					return failures.WrapTerminal(fmt.Errorf("encode task metadata: %w", encodeErr))
				}
				taskRecords = append(taskRecords, trackerTaskRecord{
					RunID:           board.RunID,
					BoardID:         board.BoardID,
					EpicID:          strings.TrimSpace(string(epic.ID)),
					TaskID:          strings.TrimSpace(string(task.ID)),
					Title:           strings.TrimSpace(task.Title),
					Description:     strings.TrimSpace(task.Description),
					Status:          strings.TrimSpace(string(task.Status)),
					Priority:        strings.TrimSpace(string(task.Priority)),
					DependsOnJSON:   string(taskDependsOn),
					MetadataJSON:    string(taskMetadata),
					CreatedAtSource: task.CreatedAt,
					UpdatedAtSource: task.UpdatedAt,
				})
				if task.Outcome != nil {
					outcomeRecords = append(outcomeRecords, trackerTaskOutcomeRecord{
						RunID:           board.RunID,
						BoardID:         board.BoardID,
						TaskID:          strings.TrimSpace(string(task.ID)),
						Status:          strings.TrimSpace(task.Outcome.Status),
						Reason:          strings.TrimSpace(task.Outcome.Reason),
						TaskBranch:      strings.TrimSpace(task.Outcome.TaskBranch),
						Worktree:        strings.TrimSpace(task.Outcome.Worktree),
						ResumeSessionID: strings.TrimSpace(task.Outcome.ResumeSessionID),
						UpdatedAtSource: task.Outcome.UpdatedAt,
					})
				}
			}
		}
		if len(epicRecords) > 0 {
			if err := tx.Create(&epicRecords).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("persist normalized epics: %w", err))
			}
		}
		if len(taskRecords) > 0 {
			if err := tx.Create(&taskRecords).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("persist normalized tasks: %w", err))
			}
		}
		if len(outcomeRecords) > 0 {
			if err := tx.Create(&outcomeRecords).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("persist normalized task outcomes: %w", err))
			}
		}
		return nil
	})
}

var _ applicationtracker.BoardStore = (*PostgresTaskboardStore)(nil)
