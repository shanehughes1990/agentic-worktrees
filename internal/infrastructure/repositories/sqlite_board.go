package repositories

import (
	"context"
	"fmt"
	"sort"
	"time"

	entity "github.com/shanehughes1990/agentic-worktrees/internal/domain/entities"
	"gorm.io/gorm"
)

type sqliteBoardRow struct {
	ID        string `gorm:"primaryKey"`
	Title     string
	Epics     []sqliteEpicRow `gorm:"foreignKey:BoardID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type sqliteEpicRow struct {
	ID           string `gorm:"primaryKey"`
	BoardID      string `gorm:"index;not null"`
	Title        string
	Description  string
	Dependencies []string        `gorm:"serializer:json"`
	SortOrder    int             `gorm:"not null"`
	Tasks        []sqliteTaskRow `gorm:"foreignKey:EpicID;references:ID;constraint:OnDelete:CASCADE"`
}

type sqliteTaskRow struct {
	ID           string `gorm:"primaryKey"`
	EpicID       string `gorm:"index;not null"`
	Title        string
	Description  string
	Status       string   `gorm:"not null"`
	Dependencies []string `gorm:"serializer:json"`
	SortOrder    int      `gorm:"not null"`
}

type SQLiteBoardRepository struct {
	db *gorm.DB
}

func NewSQLiteBoardRepository(db *gorm.DB) (*SQLiteBoardRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	if err := db.AutoMigrate(&sqliteBoardRow{}, &sqliteEpicRow{}, &sqliteTaskRow{}); err != nil {
		return nil, fmt.Errorf("migrate board tables: %w", err)
	}
	return &SQLiteBoardRepository{db: db}, nil
}

func (r *SQLiteBoardRepository) Save(ctx context.Context, board entity.Board) error {
	row := mapBoardToRow(board)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("board_id = ?", board.ID).Delete(&sqliteEpicRow{}).Error; err != nil {
			return fmt.Errorf("delete board epics: %w", err)
		}
		if err := tx.Where("id = ?", board.ID).Delete(&sqliteBoardRow{}).Error; err != nil {
			return fmt.Errorf("delete board row: %w", err)
		}
		if err := tx.Create(&row).Error; err != nil {
			return fmt.Errorf("insert board graph: %w", err)
		}
		return nil
	})
}

func (r *SQLiteBoardRepository) GetByID(ctx context.Context, id string) (entity.Board, error) {
	var row sqliteBoardRow
	if err := r.db.WithContext(ctx).
		Preload("Epics", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc") }).
		Preload("Epics.Tasks", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc") }).
		First(&row, "id = ?", id).Error; err != nil {
		return entity.Board{}, fmt.Errorf("get board by id: %w", err)
	}

	return mapRowToBoard(row), nil
}

func (r *SQLiteBoardRepository) List(ctx context.Context) ([]entity.Board, error) {
	var rows []sqliteBoardRow
	if err := r.db.WithContext(ctx).
		Preload("Epics", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc") }).
		Preload("Epics.Tasks", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc") }).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}

	boards := make([]entity.Board, 0, len(rows))
	for _, row := range rows {
		boards = append(boards, mapRowToBoard(row))
	}
	return boards, nil
}

func (r *SQLiteBoardRepository) DeleteByID(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&sqliteBoardRow{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete board by id: %w", err)
	}
	return nil
}

func mapBoardToRow(board entity.Board) sqliteBoardRow {
	epics := make([]sqliteEpicRow, 0, len(board.Epics))
	for epicIndex, epic := range board.Epics {
		tasks := make([]sqliteTaskRow, 0, len(epic.Tasks))
		for taskIndex, task := range epic.Tasks {
			tasks = append(tasks, sqliteTaskRow{
				ID:           task.ID,
				EpicID:       epic.ID,
				Title:        task.Title,
				Description:  task.Description,
				Status:       string(task.Status),
				Dependencies: task.Dependencies,
				SortOrder:    taskIndex,
			})
		}

		epics = append(epics, sqliteEpicRow{
			ID:           epic.ID,
			BoardID:      board.ID,
			Title:        epic.Title,
			Description:  epic.Description,
			Dependencies: epic.Dependencies,
			SortOrder:    epicIndex,
			Tasks:        tasks,
		})
	}

	return sqliteBoardRow{
		ID:        board.ID,
		Title:     board.Title,
		Epics:     epics,
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}
}

func mapRowToBoard(row sqliteBoardRow) entity.Board {
	sort.Slice(row.Epics, func(i, j int) bool { return row.Epics[i].SortOrder < row.Epics[j].SortOrder })
	epics := make([]entity.Epic, 0, len(row.Epics))
	for _, epicRow := range row.Epics {
		sort.Slice(epicRow.Tasks, func(i, j int) bool { return epicRow.Tasks[i].SortOrder < epicRow.Tasks[j].SortOrder })
		tasks := make([]entity.Task, 0, len(epicRow.Tasks))
		for _, taskRow := range epicRow.Tasks {
			tasks = append(tasks, entity.Task{
				ID:           taskRow.ID,
				Title:        taskRow.Title,
				Description:  taskRow.Description,
				Status:       entity.TaskStatus(taskRow.Status),
				Dependencies: taskRow.Dependencies,
			})
		}

		epics = append(epics, entity.Epic{
			ID:           epicRow.ID,
			Title:        epicRow.Title,
			Description:  epicRow.Description,
			Dependencies: epicRow.Dependencies,
			Tasks:        tasks,
		})
	}

	return entity.Board{
		ID:        row.ID,
		Title:     row.Title,
		Epics:     epics,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
