package observability

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// Entry wraps logrus.Entry and enforces context-based observability enrichment.
//
// Callers should prefer this type over raw logrus entries so correlation and trace
// fields are injected consistently through WithContext.
type Entry struct {
	raw *logrus.Entry
}

func wrapEntry(raw *logrus.Entry) *Entry {
	if raw == nil {
		return nil
	}
	return &Entry{raw: raw}
}

// Raw returns the underlying logrus entry.
//
// This is an escape hatch for integrations that require native logrus APIs.
func (entry *Entry) Raw() *logrus.Entry {
	if entry == nil {
		return nil
	}
	return entry.raw
}

// WithContext clones the entry with context attached and injects correlation and trace fields.
func (entry *Entry) WithContext(ctx context.Context) *Entry {
	if entry == nil || entry.raw == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	raw := entry.raw.WithContext(ctx)
	injectContextFields(raw, ctx)
	return &Entry{raw: raw}
}

// WithField adds a single field and returns a new wrapped entry.
func (entry *Entry) WithField(key string, value any) *Entry {
	if entry == nil || entry.raw == nil {
		return nil
	}
	return &Entry{raw: entry.raw.WithField(key, value)}
}

// WithFields adds multiple fields and returns a new wrapped entry.
func (entry *Entry) WithFields(fields logrus.Fields) *Entry {
	if entry == nil || entry.raw == nil {
		return nil
	}
	return &Entry{raw: entry.raw.WithFields(fields)}
}

// WithError adds an error field and returns a new wrapped entry.
func (entry *Entry) WithError(err error) *Entry {
	if entry == nil || entry.raw == nil {
		return nil
	}
	return &Entry{raw: entry.raw.WithError(err)}
}

// Debug logs a message at debug level.
func (entry *Entry) Debug(args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Debug(args...)
}

// Info logs a message at info level.
func (entry *Entry) Info(args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Info(args...)
}

// Warn logs a message at warn level.
func (entry *Entry) Warn(args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Warn(args...)
}

// Error logs a message at error level.
func (entry *Entry) Error(args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Error(args...)
}

// Debugf logs a formatted message at debug level.
func (entry *Entry) Debugf(format string, args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Debugf(format, args...)
}

// Infof logs a formatted message at info level.
func (entry *Entry) Infof(format string, args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Infof(format, args...)
}

// Warnf logs a formatted message at warn level.
func (entry *Entry) Warnf(format string, args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Warnf(format, args...)
}

// Errorf logs a formatted message at error level.
func (entry *Entry) Errorf(format string, args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Errorf(format, args...)
}

// NewServiceEntry creates a base service-scoped entry from a bootstrap logger.
//
// Static identity fields are attached once at creation time. Downstream calls should
// use WithContext to inject dynamic correlation and trace fields.
func NewServiceEntry(logger *logrus.Logger, serviceName, environment, version string) *Entry {
	if logger == nil {
		return nil
	}

	fields := logrus.Fields{}
	if serviceName != "" {
		fields["service"] = serviceName
	}
	if environment != "" {
		fields["environment"] = environment
	}
	if version != "" {
		fields["version"] = version
	}
	if len(fields) == 0 {
		return wrapEntry(logrus.NewEntry(logger))
	}

	return wrapEntry(logger.WithFields(fields))
}

func injectContextFields(raw *logrus.Entry, ctx context.Context) {
	if raw == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ids := CorrelationIDsFromContext(ctx)
	if ids.RunID != "" {
		raw.Data["run_id"] = ids.RunID
	}
	if ids.TaskID != "" {
		raw.Data["task_id"] = ids.TaskID
	}
	if ids.JobID != "" {
		raw.Data["job_id"] = ids.JobID
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.HasTraceID() {
		raw.Data["trace_id"] = spanContext.TraceID().String()
	}
	if spanContext.HasSpanID() {
		raw.Data["span_id"] = spanContext.SpanID().String()
	}
}
