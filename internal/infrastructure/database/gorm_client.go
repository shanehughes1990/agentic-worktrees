package database

import (
	"fmt"
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

	dialector, err := dialectorFromParsedURL(parsed)
	if err != nil {
		return nil, err
	}

	cfg := &gorm.Config{
		Logger: NewGormLogrusAdapter(appLogger).LogMode(gormlogger.Info),
	}

	db, err := gorm.Open(dialector, cfg)
	if err != nil {
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

func dialectorFromParsedURL(parsed *dburl.URL) (gorm.Dialector, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed database url cannot be nil")
	}

	switch parsed.Driver {
	case "mysql":
		return mysql.Open(parsed.DSN), nil
	case "sqlite3", "sqlite":
		return sqlite.Open(parsed.DSN), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q", parsed.Driver)
	}
}
