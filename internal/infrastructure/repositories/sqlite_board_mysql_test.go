package repositories

import (
	"context"
	"reflect"
	"testing"
	"time"

	entity "github.com/shanehughes1990/agentic-worktrees/internal/domain/entities"
	"github.com/shanehughes1990/agentic-worktrees/tests/containers"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestSQLiteBoardRepositoryCRUDWithMySQLContainer(t *testing.T) {
	ctx := context.Background()

	mysqlContainer, release, err := containers.AcquireMySQL(ctx)
	if err != nil {
		t.Skipf("mysql testcontainer unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = release(context.Background())
	})

	db, err := gorm.Open(mysql.Open(mysqlContainer.DSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("open mysql gorm db: %v", err)
	}

	repo, err := NewSQLiteBoardRepository(db)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	board := entity.Board{
		ID:        "board-mysql-1",
		Title:     "mysql board",
		CreatedAt: now,
		UpdatedAt: now,
		Epics: []entity.Epic{
			{
				ID:           "epic-1",
				Title:        "epic title",
				Description:  "epic description",
				Dependencies: []string{"epic-dep"},
				Tasks: []entity.Task{
					{
						ID:           "task-1",
						Title:        "task title",
						Description:  "task description",
						Status:       entity.TaskStatusPending,
						Dependencies: []string{"task-dep"},
					},
				},
			},
		},
	}

	if err := repo.Save(ctx, board); err != nil {
		t.Fatalf("save board: %v", err)
	}

	stored, err := repo.GetByID(ctx, board.ID)
	if err != nil {
		t.Fatalf("get board: %v", err)
	}
	if !reflect.DeepEqual(stored, board) {
		t.Fatalf("stored board mismatch\nwant: %#v\n got: %#v", board, stored)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("expected at least one board in list")
	}

	if err := repo.DeleteByID(ctx, board.ID); err != nil {
		t.Fatalf("delete board: %v", err)
	}
	if _, err := repo.GetByID(ctx, board.ID); err == nil {
		t.Fatalf("expected error when loading deleted board")
	}
}
