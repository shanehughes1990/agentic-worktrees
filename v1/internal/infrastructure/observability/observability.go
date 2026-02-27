package observability

import "context"

// LogFormat controls log output formatting.
type LogFormat string

const (
	// LogFormatText outputs human-readable text logs and is the default.
	LogFormatText LogFormat = "text"
	// LogFormatJSON outputs JSON logs and is opt-in.
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

// Config defines the external bootstrap contract for the observability platform.
type Config struct {
	ServiceName string
	Environment string
	Version     string

	LogFormat       LogFormat
	LogLevel        LogLevel
	PrettyPrintJSON bool

	OTLPEndpoint string
	OTLPHeaders  map[string]string
}

// CorrelationIDs holds workflow-level IDs used to join logs, traces, and metrics.
//
// This structure is intentionally a living expansion point and may evolve.
type CorrelationIDs struct {
	RunID  string
	TaskID string
	JobID  string
}

// Platform is the exposed observability runtime surface.
//
// It is the first-class API: bootstrap once, then use logging and operation
// instrumentation through this type.
type Platform struct {
	runtime *platformRuntime
}

// Operation represents one started operation scope.
type Operation struct {
	runtime *operationRuntime
}

// Entry is the exposed logging handle for platform and operation scopes.
type Entry struct {
	runtime *entryRuntime
}

// Bootstrap initializes the observability platform for one service instance.
//
// Logging and telemetry are always bootstrapped together internally.
func Bootstrap(ctx context.Context, config Config) (*Platform, error) {
	runtime, err := newPlatformRuntime(ctx, config)
	if err != nil {
		return nil, err
	}
	return &Platform{runtime: runtime}, nil
}

// ServiceEntry returns the service-scoped logger entry.
func (platform *Platform) ServiceEntry() *Entry {
	if platform == nil || platform.runtime == nil {
		return nil
	}
	return wrapPublicEntry(platform.runtime.serviceEntry)
}

// Entry returns a context-enriched logger entry.
func (platform *Platform) Entry(ctx context.Context) *Entry {
	if platform == nil || platform.runtime == nil || platform.runtime.serviceEntry == nil {
		return nil
	}
	return wrapPublicEntry(platform.runtime.serviceEntry.withContext(ctx))
}

// StartOperation begins one traced and metered operation scope.
func (platform *Platform) StartOperation(ctx context.Context, name string) *Operation {
	if platform == nil || platform.runtime == nil {
		return nil
	}
	operation := platform.runtime.startOperation(ctx, name)
	if operation == nil {
		return nil
	}
	return &Operation{runtime: operation}
}

// Shutdown flushes and closes platform runtime resources.
func (platform *Platform) Shutdown(ctx context.Context) error {
	if platform == nil || platform.runtime == nil {
		return nil
	}
	return platform.runtime.shutdown(ctx)
}

// Context returns the operation context.
func (operation *Operation) Context() context.Context {
	if operation == nil || operation.runtime == nil {
		return context.Background()
	}
	return operation.runtime.ctx
}

// Entry returns the operation-scoped logger entry.
func (operation *Operation) Entry() *Entry {
	if operation == nil || operation.runtime == nil {
		return nil
	}
	return wrapPublicEntry(operation.runtime.entry)
}

// End closes the operation scope, records status, and emits metrics.
func (operation *Operation) End(err error) {
	if operation == nil || operation.runtime == nil {
		return
	}
	operation.runtime.end(err)
}

// WithContext clones the entry with context attached and correlation + trace fields injected.
func (entry *Entry) WithContext(ctx context.Context) *Entry {
	if entry == nil || entry.runtime == nil {
		return nil
	}
	return wrapPublicEntry(entry.runtime.withContext(ctx))
}

// WithField adds a single field and returns a new entry.
func (entry *Entry) WithField(key string, value any) *Entry {
	if entry == nil || entry.runtime == nil {
		return nil
	}
	return wrapPublicEntry(entry.runtime.withField(key, value))
}

// WithFields adds multiple fields and returns a new entry.
func (entry *Entry) WithFields(fields map[string]any) *Entry {
	if entry == nil || entry.runtime == nil {
		return nil
	}
	return wrapPublicEntry(entry.runtime.withFields(fields))
}

// WithError adds an error field and returns a new entry.
func (entry *Entry) WithError(err error) *Entry {
	if entry == nil || entry.runtime == nil {
		return nil
	}
	return wrapPublicEntry(entry.runtime.withError(err))
}

func (entry *Entry) Debug(args ...any) {
	if entry == nil || entry.runtime == nil {
		return
	}
	entry.runtime.debug(args...)
}

func (entry *Entry) Info(args ...any) {
	if entry == nil || entry.runtime == nil {
		return
	}
	entry.runtime.info(args...)
}

func (entry *Entry) Warn(args ...any) {
	if entry == nil || entry.runtime == nil {
		return
	}
	entry.runtime.warn(args...)
}

func (entry *Entry) Error(args ...any) {
	if entry == nil || entry.runtime == nil {
		return
	}
	entry.runtime.error(args...)
}

func (entry *Entry) Debugf(format string, args ...any) {
	if entry == nil || entry.runtime == nil {
		return
	}
	entry.runtime.debugf(format, args...)
}

func (entry *Entry) Infof(format string, args ...any) {
	if entry == nil || entry.runtime == nil {
		return
	}
	entry.runtime.infof(format, args...)
}

func (entry *Entry) Warnf(format string, args ...any) {
	if entry == nil || entry.runtime == nil {
		return
	}
	entry.runtime.warnf(format, args...)
}

func (entry *Entry) Errorf(format string, args ...any) {
	if entry == nil || entry.runtime == nil {
		return
	}
	entry.runtime.errorf(format, args...)
}

type correlationContextKey struct{}

// WithCorrelationIDs stores correlation IDs on context used by platform observability.
func WithCorrelationIDs(ctx context.Context, ids CorrelationIDs) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, correlationContextKey{}, ids)
}

// CorrelationIDsFromContext reads correlation IDs from context.
func CorrelationIDsFromContext(ctx context.Context) CorrelationIDs {
	if ctx == nil {
		return CorrelationIDs{}
	}
	ids, ok := ctx.Value(correlationContextKey{}).(CorrelationIDs)
	if !ok {
		return CorrelationIDs{}
	}
	return ids
}

func wrapPublicEntry(runtime *entryRuntime) *Entry {
	if runtime == nil {
		return nil
	}
	return &Entry{runtime: runtime}
}
