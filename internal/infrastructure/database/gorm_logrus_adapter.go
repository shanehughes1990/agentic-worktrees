package database

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type gormLogrusAdapter struct {
	logger        *logrus.Logger
	logLevel      gormlogger.LogLevel
	slowThreshold time.Duration
}

func NewGormLogrusAdapter(appLogger *logrus.Logger) gormlogger.Interface {
	if appLogger == nil {
		appLogger = logrus.New()
		appLogger.SetLevel(logrus.TraceLevel)
	}

	return &gormLogrusAdapter{
		logger:        appLogger,
		logLevel:      gormlogger.Info,
		slowThreshold: 200 * time.Millisecond,
	}
}

func (l *gormLogrusAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	copy := *l
	copy.logLevel = level
	return &copy
}

func (l *gormLogrusAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Info {
		return
	}
	l.logger.WithContext(ctx).Debugf(msg, data...)
}

func (l *gormLogrusAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Warn {
		return
	}
	l.logger.WithContext(ctx).Warnf(msg, data...)
}

func (l *gormLogrusAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Error {
		return
	}
	l.logger.WithContext(ctx).Errorf(msg, data...)
}

func (l *gormLogrusAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.logLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		l.logger.WithContext(ctx).WithError(err).Tracef("gorm trace error elapsed=%s rows=%d sql=%s", elapsed, rows, sql)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.logLevel >= gormlogger.Warn:
		l.logger.WithContext(ctx).Tracef("gorm trace slow_query elapsed=%s rows=%d sql=%s", elapsed, rows, sql)
	case l.logLevel >= gormlogger.Info:
		l.logger.WithContext(ctx).Tracef("gorm trace elapsed=%s rows=%d sql=%s", elapsed, rows, sql)
	}
}
