package observability

import "context"

// LogFormat controls log output formatting.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// LogLevel controls runtime log verbosity.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Config defines the bootstrap contract for observability.
type Config struct {
	ServiceName string
	Environment string
	Version     string

	LogFormat LogFormat
	LogLevel  LogLevel

	OTLPEndpoint string
	OTLPHeaders  map[string]string
}

// CorrelationIDs holds workflow-level IDs used to join logs, traces, and metrics.
type CorrelationIDs struct {
	RunID     string
	TaskID    string
	JobID     string
	ProjectID string
}

// Platform is the exposed observability runtime surface.
type Platform interface {
	ServiceEntry() Entry
	StartOperation(ctx context.Context, name string) Operation
	Shutdown(ctx context.Context) error
}

// Operation represents one started operation scope.
type Operation interface {
	Context() context.Context
	WithField(key string, value any) Operation
	WithFields(fields map[string]any) Operation
	WithError(err error) Operation
	End(err error)
}

// Entry is the exposed logging handle for platform and operation scopes.
type Entry interface {
	WithField(key string, value any) Entry
	WithFields(fields map[string]any) Entry
	WithError(err error) Entry
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)
}

// Factory bootstraps platform instances.
type Factory interface {
	Bootstrap(ctx context.Context, config Config) (Platform, error)
}
