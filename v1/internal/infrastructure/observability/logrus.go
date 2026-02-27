package observability

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type loggerOptions struct {
	format          LogFormat
	level           logrus.Level
	timestampFormat string
	prettyPrintJSON bool
}

type entryRuntime struct {
	raw *logrus.Entry
}

func loggerOptionsFromConfig(config Config) loggerOptions {
	format := config.LogFormat
	if format == "" {
		format = LogFormatText
	}

	level := logrus.InfoLevel
	switch strings.ToLower(string(config.LogLevel)) {
	case "debug":
		level = logrus.DebugLevel
	case "warn", "warning":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	case "info", "":
		level = logrus.InfoLevel
	}

	timestampFormat := config.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = DefaultTimestampFormat
	}

	return loggerOptions{
		format:          format,
		level:           level,
		timestampFormat: timestampFormat,
		prettyPrintJSON: config.PrettyPrintJSON,
	}
}

func buildLogger(options loggerOptions) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(options.level)

	switch options.format {
	case LogFormatJSON:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:   options.timestampFormat,
			DisableHTMLEscape: true,
			PrettyPrint:       options.prettyPrintJSON,
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        options.timestampFormat,
			DisableLevelTruncation: true,
			PadLevelText:           true,
			DisableSorting:         false,
			ForceQuote:             true,
			QuoteEmptyFields:       true,
			DisableQuote:           false,
		})
	}

	return logger
}

func buildServiceEntry(logger *logrus.Logger, serviceName, environment, version string) *entryRuntime {
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
		return &entryRuntime{raw: logrus.NewEntry(logger)}
	}

	return &entryRuntime{raw: logger.WithFields(fields)}
}

func (entry *entryRuntime) withContext(ctx context.Context) *entryRuntime {
	if entry == nil || entry.raw == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	raw := entry.raw.WithContext(ctx)
	injectContextFields(raw, ctx)
	return &entryRuntime{raw: raw}
}

func (entry *entryRuntime) withField(key string, value any) *entryRuntime {
	if entry == nil || entry.raw == nil {
		return nil
	}
	return &entryRuntime{raw: entry.raw.WithField(key, value)}
}

func (entry *entryRuntime) withFields(fields map[string]any) *entryRuntime {
	if entry == nil || entry.raw == nil {
		return nil
	}
	if len(fields) == 0 {
		return entry
	}
	converted := logrus.Fields{}
	for key, value := range fields {
		converted[key] = value
	}
	return &entryRuntime{raw: entry.raw.WithFields(converted)}
}

func (entry *entryRuntime) withError(err error) *entryRuntime {
	if entry == nil || entry.raw == nil {
		return nil
	}
	return &entryRuntime{raw: entry.raw.WithError(err)}
}

func (entry *entryRuntime) debug(args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Debug(args...)
}

func (entry *entryRuntime) info(args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Info(args...)
}

func (entry *entryRuntime) warn(args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Warn(args...)
}

func (entry *entryRuntime) error(args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Error(args...)
}

func (entry *entryRuntime) debugf(format string, args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Debugf(format, args...)
}

func (entry *entryRuntime) infof(format string, args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Infof(format, args...)
}

func (entry *entryRuntime) warnf(format string, args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Warnf(format, args...)
}

func (entry *entryRuntime) errorf(format string, args ...any) {
	if entry == nil || entry.raw == nil {
		return
	}
	entry.raw.Errorf(format, args...)
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
