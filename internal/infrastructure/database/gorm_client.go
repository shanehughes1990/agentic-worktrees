package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/xo/dburl"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type GormClient struct {
	db *gorm.DB
}

func NewGormClient(appLogger *logrus.Logger, databaseURL string) (*GormClient, error) {
	if appLogger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("database url cannot be empty")
	}

	parsed, err := dburl.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	if !hasDatabaseName(parsed) {
		return nil, fmt.Errorf("database must be present in url")
	}

	dialector, err := dialectorFromParsedURL(parsed, appLogger)
	if err != nil {
		return nil, err
	}

	cfg := &gorm.Config{
		Logger: NewGormLogrusAdapter(appLogger).LogMode(gormlogger.Info),
	}

	db, err := gorm.Open(dialector, cfg)
	if err != nil {
		if parsed.Driver == "sqlite3" || parsed.Driver == "sqlite" {
			return nil, fmt.Errorf("open gorm client for sqlite path %q: %w", strings.TrimSpace(parsed.DSN), err)
		}
		return nil, fmt.Errorf("open gorm client: %w", err)
	}

	return &GormClient{db: db}, nil
}

func (c *GormClient) DB() *gorm.DB {
	if c == nil {
		return nil
	}
	return c.db
}

func hasDatabaseName(parsed *dburl.URL) bool {
	if parsed == nil {
		return false
	}

	switch parsed.Driver {
	case "sqlite3", "sqlite":
		dsn := strings.TrimSpace(parsed.DSN)
		return dsn != "" && dsn != "/"
	default:
		path := strings.TrimSpace(parsed.Path)
		return path != "" && path != "/"
	}
}

func dialectorFromParsedURL(parsed *dburl.URL, appLogger *logrus.Logger) (gorm.Dialector, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed database url cannot be nil")
	}

	switch parsed.Driver {
	case "mysql":
		return mysql.Open(parsed.DSN), nil
	case "sqlite3", "sqlite":
		resolvedPath, err := ensureSQLitePath(parsed.DSN)
		if err != nil {
			return nil, err
		}
		parsed.DSN = resolvedPath
		if appLogger != nil {
			appLogger.WithField("sqlite_path", resolvedPath).Info("using sqlite database path")
		}
		return sqlite.Open(resolvedPath), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q", parsed.Driver)
	}
}

func ensureSQLitePath(rawDSN string) (string, error) {
	dsn := strings.TrimSpace(rawDSN)
	if dsn == "" || dsn == "/" {
		return "", fmt.Errorf("sqlite database path cannot be empty")
	}

	cleanPath := strings.TrimPrefix(dsn, "file:")
	if idx := strings.Index(cleanPath, "?"); idx >= 0 {
		cleanPath = cleanPath[:idx]
	}

	if cleanPath == "" || cleanPath == "/" || cleanPath == ":memory:" {
		return dsn, nil
	}

	resolvedPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("resolve sqlite database path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return "", fmt.Errorf("create sqlite database directory %q: %w", filepath.Dir(resolvedPath), err)
	}
	if _, err := os.Stat(resolvedPath); err != nil {
		if os.IsNotExist(err) {
			file, createErr := os.OpenFile(resolvedPath, os.O_CREATE, 0o644)
			if createErr != nil {
				return "", fmt.Errorf("create sqlite database file %q: %w", resolvedPath, createErr)
			}
			if closeErr := file.Close(); closeErr != nil {
				return "", fmt.Errorf("close sqlite database file %q: %w", resolvedPath, closeErr)
			}
		} else {
			return "", fmt.Errorf("stat sqlite database file %q: %w", resolvedPath, err)
		}
	}
	return resolvedPath, nil
}
