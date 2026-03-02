package postgres

import (
	domainobservability "agentic-orchestrator/internal/domain/shared/observability"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xo/dburl"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 25
	defaultConnMaxLifetime = 30 * time.Minute
)

type Config struct {
	DSN                string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    time.Duration
	SlowQueryThreshold time.Duration
}

func (config Config) normalized() Config {
	normalized := config
	normalized.DSN = strings.TrimSpace(normalized.DSN)
	if normalized.MaxOpenConns <= 0 {
		normalized.MaxOpenConns = defaultMaxOpenConns
	}
	if normalized.MaxIdleConns <= 0 {
		normalized.MaxIdleConns = defaultMaxIdleConns
	}
	if normalized.ConnMaxLifetime <= 0 {
		normalized.ConnMaxLifetime = defaultConnMaxLifetime
	}
	return normalized
}

type Client struct {
	db *gorm.DB
}

func Open(ctx context.Context, config Config, entry domainobservability.Entry) (*Client, error) {
	normalized := config.normalized()
	if normalized.DSN == "" {
		return nil, fmt.Errorf("database dsn is required")
	}

	parsedURL, err := dburl.Parse(normalized.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}
	if !isPostgresScheme(parsedURL.Scheme) {
		return nil, fmt.Errorf("database dsn must use postgres scheme, got %q", parsedURL.Scheme)
	}

	db, err := gorm.Open(gormpostgres.Open(parsedURL.DSN), &gorm.Config{
		Logger: newGormLogger(entry, normalized.SlowQueryThreshold),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres client: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("resolve sql db handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(normalized.MaxOpenConns)
	sqlDB.SetMaxIdleConns(normalized.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(normalized.ConnMaxLifetime)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Client{db: db}, nil
}

func (client *Client) DB() *gorm.DB {
	if client == nil {
		return nil
	}
	return client.db
}

func (client *Client) Close() error {
	if client == nil || client.db == nil {
		return nil
	}
	sqlDB, err := client.db.DB()
	if err != nil {
		return fmt.Errorf("resolve sql db handle: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close postgres db: %w", err)
	}
	return nil
}

func isPostgresScheme(scheme string) bool {
	switch strings.ToLower(strings.TrimSpace(scheme)) {
	case "postgres", "postgresql":
		return true
	default:
		return false
	}
}
