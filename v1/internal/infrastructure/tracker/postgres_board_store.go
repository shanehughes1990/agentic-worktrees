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

type boardSnapshotRecord struct {
	gorm.Model
	RunID      string `gorm:"column:run_id;size:255;not null;uniqueIndex:idx_tracker_board_snapshot,priority:1"`
	BoardID    string `gorm:"column:board_id;size:255;not null;uniqueIndex:idx_tracker_board_snapshot,priority:2"`
	SourceKind string `gorm:"column:source_kind;size:64;not null"`
	SourceRef  string `gorm:"column:source_ref"`
	Payload    []byte `gorm:"column:payload;not null"`
	Revision   int64  `gorm:"column:revision;not null;default:1"`
}

func (boardSnapshotRecord) TableName() string {
	return "tracker_board_snapshots"
}

type taskClaimRecord struct {
	gorm.Model
	ClaimID           string    `gorm:"column:claim_id;size:255;not null;uniqueIndex"`
	RunID             string    `gorm:"column:run_id;size:255;not null;index:idx_tracker_task_claim,priority:1"`
	BoardID           string    `gorm:"column:board_id;size:255;not null;index:idx_tracker_task_claim,priority:2"`
	TaskID            string    `gorm:"column:task_id;size:255;not null;index:idx_tracker_task_claim,priority:3"`
	WorkerID          string    `gorm:"column:worker_id;size:255;not null"`
	State             string    `gorm:"column:state;size:32;not null;index"`
	ClaimedRevision   int64     `gorm:"column:claimed_revision;not null"`
	CommittedRevision int64     `gorm:"column:committed_revision;not null;default:0"`
	ClaimedAt         time.Time `gorm:"column:claimed_at;not null"`
	CompletedAt       time.Time `gorm:"column:completed_at"`
}

func (taskClaimRecord) TableName() string {
	return "tracker_task_claims"
}

type PostgresBoardStore struct {
	db *gorm.DB
}

func NewPostgresBoardStore(db *gorm.DB) (*PostgresBoardStore, error) {
	if db == nil {
		return nil, failures.WrapTerminal(errors.New("postgres board store db is required"))
	}
	if err := db.AutoMigrate(&boardSnapshotRecord{}, &taskClaimRecord{}); err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("migrate tracker board tables: %w", err))
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_tracker_task_claim_active_unique ON tracker_task_claims (run_id, board_id, task_id) WHERE state = 'active'`).Error; err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("ensure active claim uniqueness index: %w", err))
	}
	return &PostgresBoardStore{db: db}, nil
}

func (store *PostgresBoardStore) LoadBoard(ctx context.Context, projectID string, boardID string) (domaintracker.Board, error) {
	if store == nil || store.db == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	board, _, err := store.loadBoardSnapshot(store.db.WithContext(ctx), projectID, boardID, false)
	if err != nil {
		return domaintracker.Board{}, err
	}
	return board, nil
}

func (store *PostgresBoardStore) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	if store == nil || store.db == nil {
		return failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	if err := board.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(board)
	if err != nil {
		return failures.WrapTerminal(fmt.Errorf("encode board snapshot payload: %w", err))
	}
	sourceRef := boardSourceRef(board)
	runID := strings.TrimSpace(board.RunID)
	boardID := strings.TrimSpace(board.BoardID)
	sourceKind := strings.TrimSpace(string(board.Source.Kind))

	return store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing boardSnapshotRecord
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("run_id = ? AND board_id = ?", runID, boardID).
			Take(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			record := boardSnapshotRecord{
				RunID:      runID,
				BoardID:    boardID,
				SourceKind: sourceKind,
				SourceRef:  sourceRef,
				Payload:    payload,
				Revision:   1,
			}
			if err := tx.Create(&record).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("create board snapshot: %w", err))
			}
			return nil
		case err != nil:
			return failures.WrapTransient(fmt.Errorf("load board snapshot for upsert: %w", err))
		default:
			nextRevision := existing.Revision + 1
			if nextRevision <= 0 {
				nextRevision = 1
			}
			if err := tx.Model(&boardSnapshotRecord{}).
				Where("id = ?", existing.ID).
				Updates(map[string]any{
					"source_kind": sourceKind,
					"source_ref":  sourceRef,
					"payload":     payload,
					"revision":    nextRevision,
					"updated_at":  time.Now().UTC(),
				}).Error; err != nil {
				return failures.WrapTransient(fmt.Errorf("update board snapshot: %w", err))
			}
			return nil
		}
	})
}

func (store *PostgresBoardStore) ClaimNextTask(ctx context.Context, projectID string, boardID string, workerID string) (domaintracker.Board, domaintracker.Task, string, int64, error) {
	if store == nil || store.db == nil {
		return domaintracker.Board{}, domaintracker.Task{}, "", 0, failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	cleanWorkerID := strings.TrimSpace(workerID)
	if cleanProjectID == "" || cleanBoardID == "" || cleanWorkerID == "" {
		return domaintracker.Board{}, domaintracker.Task{}, "", 0, failures.WrapTerminal(errors.New("project_id, board_id, and worker_id are required"))
	}

	var claimedBoard domaintracker.Board
	var claimedTask domaintracker.Task
	var claimID string
	var revision int64

	err := store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		board, snapshot, loadErr := store.loadBoardSnapshot(tx, cleanProjectID, cleanBoardID, true)
		if loadErr != nil {
			return loadErr
		}

		now := time.Now().UTC()
		claimed := false
		for epicIndex := range board.Epics {
			for taskIndex := range board.Epics[epicIndex].Tasks {
				task := &board.Epics[epicIndex].Tasks[taskIndex]
				if task.Status != domaintracker.StatusNotStarted {
					continue
				}
				task.Status = domaintracker.StatusInProgress
				task.UpdatedAt = now
				claimedTask = *task
				claimed = true
				break
			}
			if claimed {
				break
			}
		}
		if !claimed {
			return failures.WrapTerminal(errors.New("no claimable task found"))
		}

		board.UpdatedAt = now
		revision = snapshot.Revision + 1
		if revision <= 0 {
			revision = 1
		}
		if err := store.persistSnapshotTx(tx, snapshot.ID, board, revision); err != nil {
			return err
		}

		claimID = fmt.Sprintf("claim-%s-%d", cleanWorkerID, now.UnixNano())
		claim := taskClaimRecord{
			ClaimID:           claimID,
			RunID:             cleanProjectID,
			BoardID:           cleanBoardID,
			TaskID:            strings.TrimSpace(string(claimedTask.ID)),
			WorkerID:          cleanWorkerID,
			State:             "active",
			ClaimedRevision:   revision,
			CommittedRevision: 0,
			ClaimedAt:         now,
		}
		if err := tx.Create(&claim).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("persist task claim: %w", err))
		}
		claimedBoard = board
		return nil
	})
	if err != nil {
		return domaintracker.Board{}, domaintracker.Task{}, "", 0, err
	}
	return claimedBoard, claimedTask, claimID, revision, nil
}

func (store *PostgresBoardStore) ApplyTaskResult(ctx context.Context, projectID string, boardID string, claimID string, taskID string, nextStatus domaintracker.Status, outcome domaintracker.TaskOutcome) (domaintracker.Board, int64, error) {
	if store == nil || store.db == nil {
		return domaintracker.Board{}, 0, failures.WrapTerminal(errors.New("postgres board store is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	cleanClaimID := strings.TrimSpace(claimID)
	cleanTaskID := strings.TrimSpace(taskID)
	if cleanProjectID == "" || cleanBoardID == "" || cleanClaimID == "" || cleanTaskID == "" {
		return domaintracker.Board{}, 0, failures.WrapTerminal(errors.New("project_id, board_id, claim_id, and task_id are required"))
	}

	var updatedBoard domaintracker.Board
	var committedRevision int64

	err := store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var claim taskClaimRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("claim_id = ?", cleanClaimID).
			Take(&claim).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return failures.WrapTerminal(errors.New("task claim not found"))
			}
			return failures.WrapTransient(fmt.Errorf("load task claim: %w", err))
		}
		if claim.State != "active" {
			return failures.WrapTerminal(errors.New("task claim is no longer active"))
		}
		if claim.RunID != cleanProjectID || claim.BoardID != cleanBoardID || claim.TaskID != cleanTaskID {
			return failures.WrapTerminal(errors.New("task claim does not match target task"))
		}

		board, snapshot, loadErr := store.loadBoardSnapshot(tx, cleanProjectID, cleanBoardID, true)
		if loadErr != nil {
			return loadErr
		}

		now := time.Now().UTC()
		updated := false
		for epicIndex := range board.Epics {
			for taskIndex := range board.Epics[epicIndex].Tasks {
				task := &board.Epics[epicIndex].Tasks[taskIndex]
				if strings.TrimSpace(string(task.ID)) != cleanTaskID {
					continue
				}
				task.Status = nextStatus
				outcome.UpdatedAt = now
				task.Outcome = &outcome
				task.UpdatedAt = now
				updated = true
				break
			}
			if updated {
				break
			}
		}
		if !updated {
			return failures.WrapTerminal(errors.New("task not found in latest board snapshot"))
		}

		board.UpdatedAt = now
		committedRevision = snapshot.Revision + 1
		if committedRevision <= 0 {
			committedRevision = 1
		}
		if err := store.persistSnapshotTx(tx, snapshot.ID, board, committedRevision); err != nil {
			return err
		}

		if err := tx.Model(&taskClaimRecord{}).
			Where("id = ?", claim.ID).
			Updates(map[string]any{
				"state":               "completed",
				"committed_revision":  committedRevision,
				"completed_at":        now,
			}).Error; err != nil {
			return failures.WrapTransient(fmt.Errorf("complete task claim: %w", err))
		}
		updatedBoard = board
		return nil
	})
	if err != nil {
		return domaintracker.Board{}, 0, err
	}
	return updatedBoard, committedRevision, nil
}

func (store *PostgresBoardStore) persistSnapshotTx(tx *gorm.DB, snapshotID uint, board domaintracker.Board, revision int64) error {
	payload, err := json.Marshal(board)
	if err != nil {
		return failures.WrapTerminal(fmt.Errorf("encode board snapshot payload: %w", err))
	}
	if err := tx.Model(&boardSnapshotRecord{}).
		Where("id = ?", snapshotID).
		Updates(map[string]any{
			"source_kind": strings.TrimSpace(string(board.Source.Kind)),
			"source_ref":  boardSourceRef(board),
			"payload":     payload,
			"revision":    revision,
			"updated_at":  time.Now().UTC(),
		}).Error; err != nil {
		return failures.WrapTransient(fmt.Errorf("persist latest board snapshot: %w", err))
	}
	return nil
}

func (store *PostgresBoardStore) loadBoardSnapshot(tx *gorm.DB, projectID string, boardID string, forUpdate bool) (domaintracker.Board, boardSnapshotRecord, error) {
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanProjectID == "" || cleanBoardID == "" {
		return domaintracker.Board{}, boardSnapshotRecord{}, failures.WrapTerminal(errors.New("project_id and board_id are required"))
	}

	query := tx.Where("run_id = ? AND board_id = ?", cleanProjectID, cleanBoardID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var record boardSnapshotRecord
	if err := query.Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domaintracker.Board{}, boardSnapshotRecord{}, failures.WrapTerminal(fmt.Errorf("internal tracker board %q not found for project %q", cleanBoardID, cleanProjectID))
		}
		return domaintracker.Board{}, boardSnapshotRecord{}, failures.WrapTransient(fmt.Errorf("load internal tracker board snapshot: %w", err))
	}

	var board domaintracker.Board
	if err := json.Unmarshal(record.Payload, &board); err != nil {
		return domaintracker.Board{}, boardSnapshotRecord{}, failures.WrapTerminal(fmt.Errorf("decode internal tracker board snapshot: %w", err))
	}
	if strings.TrimSpace(board.BoardID) == "" {
		board.BoardID = cleanBoardID
	}
	if strings.TrimSpace(board.RunID) == "" {
		board.RunID = cleanProjectID
	}
	return board, record, nil
}

func boardSourceRef(board domaintracker.Board) string {
	sourceRef := strings.TrimSpace(board.Source.BoardID)
	if sourceRef == "" {
		sourceRef = strings.TrimSpace(board.Source.Location)
	}
	if sourceRef == "" {
		sourceRef = strings.TrimSpace(board.BoardID)
	}
	return sourceRef
}

var _ applicationtracker.BoardStore = (*PostgresBoardStore)(nil)
