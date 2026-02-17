package database

import (
	"net/url"
	"path/filepath"
	"testing"

	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
	"github.com/xo/dburl"
)

func TestNewGormClientRejectsNilLogger(t *testing.T) {
	databaseURL := "sqlite:///" + filepath.Join(t.TempDir(), "test.db")
	client, err := NewGormClient(nil, databaseURL)
	if err == nil {
		t.Fatalf("expected error for nil logger")
	}
	if client != nil {
		t.Fatalf("expected nil client when constructor fails")
	}
}

func TestNewGormClientRejectsEmptyDatabaseURL(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	client, err := NewGormClient(appLogger, "")
	if err == nil {
		t.Fatalf("expected error for empty database url")
	}
	if client != nil {
		t.Fatalf("expected nil client when constructor fails")
	}
}

func TestNewGormClientRejectsInvalidDatabaseURL(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	client, err := NewGormClient(appLogger, "://bad-url")
	if err == nil {
		t.Fatalf("expected error for invalid database url")
	}
	if client != nil {
		t.Fatalf("expected nil client when constructor fails")
	}
}

func TestNewGormClientRequiresDatabaseInURL(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	client, err := NewGormClient(appLogger, "mysql://user:pass@localhost:3306")
	if err == nil {
		t.Fatalf("expected error when database is missing")
	}
	if client != nil {
		t.Fatalf("expected nil client when constructor fails")
	}
}

func TestNewGormClientRejectsUnsupportedDriver(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	client, err := NewGormClient(appLogger, "postgres://user:pass@localhost:5432/appdb")
	if err == nil {
		t.Fatalf("expected error for unsupported driver")
	}
	if client != nil {
		t.Fatalf("expected nil client when constructor fails")
	}
}

func TestNewGormClientReturnsOpenError(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	databaseURL := "sqlite:///" + t.TempDir()
	client, err := NewGormClient(appLogger, databaseURL)
	if err == nil {
		t.Fatalf("expected open gorm client error")
	}
	if client != nil {
		t.Fatalf("expected nil client when constructor fails")
	}
}

func TestNewGormClientCreatesClient(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	databaseURL := "sqlite:///" + filepath.Join(t.TempDir(), "test.db")
	client, err := NewGormClient(appLogger, databaseURL)
	if err != nil {
		t.Fatalf("new gorm client: %v", err)
	}
	if client == nil {
		t.Fatalf("expected non-nil client")
	}
	if client.DB() == nil {
		t.Fatalf("expected non-nil gorm db")
	}
}

func TestGormClientDBNilReceiver(t *testing.T) {
	var client *GormClient
	if client.DB() != nil {
		t.Fatalf("expected nil db for nil receiver")
	}
}

func TestHasDatabaseName(t *testing.T) {
	if hasDatabaseName(nil) {
		t.Fatalf("expected nil parsed url to return false")
	}

	emptyPath := &dburl.URL{URL: url.URL{Path: "/"}}
	if hasDatabaseName(emptyPath) {
		t.Fatalf("expected slash-only path to return false")
	}

	databasePath := &dburl.URL{URL: url.URL{Path: "/appdb"}}
	if !hasDatabaseName(databasePath) {
		t.Fatalf("expected database path to return true")
	}

	sqliteWithoutDatabase := &dburl.URL{Driver: "sqlite3", DSN: ""}
	if hasDatabaseName(sqliteWithoutDatabase) {
		t.Fatalf("expected sqlite empty dsn to return false")
	}

	sqliteWithDatabase := &dburl.URL{Driver: "sqlite3", DSN: "/tmp/test.db"}
	if !hasDatabaseName(sqliteWithDatabase) {
		t.Fatalf("expected sqlite dsn to return true")
	}
}

func TestDialectorFromParsedURL(t *testing.T) {
	if _, err := dialectorFromParsedURL(nil); err == nil {
		t.Fatalf("expected error for nil parsed url")
	}

	mysqlURL, err := dburl.Parse("mysql://user:pass@localhost:3306/appdb")
	if err != nil {
		t.Fatalf("parse mysql url: %v", err)
	}
	if _, err := dialectorFromParsedURL(mysqlURL); err != nil {
		t.Fatalf("expected mysql dialector: %v", err)
	}

	sqliteURL, err := dburl.Parse("sqlite:///tmp/test.db")
	if err != nil {
		t.Fatalf("parse sqlite url: %v", err)
	}
	if _, err := dialectorFromParsedURL(sqliteURL); err != nil {
		t.Fatalf("expected sqlite dialector: %v", err)
	}

	postgresURL, err := dburl.Parse("postgres://user:pass@localhost:5432/appdb")
	if err != nil {
		t.Fatalf("parse postgres url: %v", err)
	}
	if _, err := dialectorFromParsedURL(postgresURL); err == nil {
		t.Fatalf("expected unsupported driver error")
	}
}
