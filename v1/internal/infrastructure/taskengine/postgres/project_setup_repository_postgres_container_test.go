package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	testcontainers "agentic-orchestrator/tests/containers"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestProjectSetupRepositoryMigratesOnFreshPostgres(t *testing.T) {
	if os.Getenv("RUN_DOCKER_TESTS") != "1" {
		t.Skip("set RUN_DOCKER_TESTS=1 to run docker-backed postgres migration tests")
	}

	ctx := context.Background()
	started, err := testcontainers.Start(ctx, []testcontainers.RequestSpec{testcontainers.NewPostgresRequest("project-setup-migration")})
	if err != nil {
		if testcontainers.IsDockerUnavailable(err) {
			t.Skipf("docker unavailable: %v", err)
		}
		t.Fatalf("start postgres container: %v", err)
	}
	defer func() {
		_ = testcontainers.Terminate(context.Background(), started)
	}()

	resource, ok := started[0].Resource.(testcontainers.PostgresResource)
	if !ok {
		t.Fatalf("unexpected postgres resource type: %T", started[0].Resource)
	}

	var db *gorm.DB
	var errOpen error
	for attempt := 1; attempt <= 10; attempt++ {
		db, errOpen = gorm.Open(postgres.Open(resource.DSN), &gorm.Config{})
		if errOpen == nil {
			sqlDB, sqlErr := db.DB()
			if sqlErr == nil {
				pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				pingErr := sqlDB.PingContext(pingCtx)
				cancel()
				if pingErr == nil {
					break
				}
				errOpen = pingErr
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	if errOpen != nil {
		t.Fatalf("open postgres gorm db: %v", errOpen)
	}

	crypto, err := NewSCMTokenCrypto(db)
	if err != nil {
		t.Fatalf("new scm token crypto: %v", err)
	}

	repo, err := NewProjectSetupRepository(db, crypto)
	if err != nil {
		t.Fatalf("new project setup repository: %v", err)
	}

	if _, err := repo.UpsertProjectSetup(ctx, sampleProjectSetup()); err != nil {
		t.Fatalf("upsert project setup on fresh postgres: %v", err)
	}

	loaded, err := repo.GetProjectSetup(ctx, "project-1")
	if err != nil {
		t.Fatalf("get project setup: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected project setup after upsert")
	}
}
