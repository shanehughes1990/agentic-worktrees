package postgres

import (
	domainobservability "agentic-orchestrator/internal/domain/shared/observability"
	"context"
	"fmt"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

const defaultSlowQueryThreshold = 500 * time.Millisecond

type gormLogAdapter struct {
	entry              domainobservability.Entry
	slowQueryThreshold time.Duration
}

func newGormLogger(entry domainobservability.Entry, slowQueryThreshold time.Duration) gormlogger.Interface {
	if slowQueryThreshold <= 0 {
		slowQueryThreshold = defaultSlowQueryThreshold
	}
	return &gormLogAdapter{entry: entry, slowQueryThreshold: slowQueryThreshold}
}

func (logger *gormLogAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	_ = level
	return logger
}

func (logger *gormLogAdapter) Info(ctx context.Context, message string, args ...any) {
	logger.debug(ctx, message, args...)
}

func (logger *gormLogAdapter) Warn(ctx context.Context, message string, args ...any) {
	logger.debug(ctx, message, args...)
}

func (logger *gormLogAdapter) Error(ctx context.Context, message string, args ...any) {
	logger.debug(ctx, message, args...)
}

func (logger *gormLogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if logger == nil || logger.entry == nil || fc == nil {
		return
	}
	query, rowsAffected := fc()
	elapsed := time.Since(begin)
	entry := logger.entry.WithFields(map[string]any{
		"component":     "postgres",
		"sql":           query,
		"rows_affected": rowsAffected,
		"duration_ms":   elapsed.Milliseconds(),
	})
	_ = ctx
	if err != nil {
		entry = entry.WithError(err)
	}
	if isSlowQuery(elapsed, logger.slowQueryThreshold) {
		entry.Warn("postgres slow query")
		return
	}
	entry.Debug("postgres query")
}

func (logger *gormLogAdapter) debug(ctx context.Context, message string, args ...any) {
	if logger == nil || logger.entry == nil {
		return
	}
	logger.entry.WithField("component", "postgres").Debug(fmt.Sprintf(message, args...))
	_ = ctx
}

func isSlowQuery(elapsed time.Duration, threshold time.Duration) bool {
	if threshold <= 0 {
		threshold = defaultSlowQueryThreshold
	}
	return elapsed >= threshold
}
