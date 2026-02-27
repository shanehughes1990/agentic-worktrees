package observability

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Config configures the observability bundle for service-level runtime behavior.
type Config struct {
	ServiceName     string
	Environment     string
	Version         string
	LoggerType      LoggerType
	LoggerLevel     logrus.Level
	TimestampFormat string
	PrettyPrintJSON bool
}

// Bundle groups logger, wrapped entry, and internal OpenTelemetry handles.
//
// OpenTelemetry internals are intentionally hidden behind this type so callers
// interact through observability-native APIs only.
type Bundle struct {
	serviceName        string
	environment        string
	version            string
	loggerType         LoggerType
	loggerLevel        logrus.Level
	Logger             *logrus.Logger
	baseEntry          *Entry
	tracer             trace.Tracer
	meter              metric.Meter
	operationCounter   metric.Int64Counter
	operationDurations metric.Float64Histogram
}

// Operation represents one started operation scope.
//
// It carries the context, wrapped logger entry, and internal span handle until End is called.
type Operation struct {
	ctx       context.Context
	entry     *Entry
	span      trace.Span
	bundle    *Bundle
	name      string
	ids       CorrelationIDs
	startedAt time.Time
	ended     bool
}

// NewBundle builds a service-scoped observability bundle.
func NewBundle(config Config) (*Bundle, error) {
	if config.ServiceName == "" {
		return nil, errors.New("service name is required")
	}
	if config.Environment == "" {
		config.Environment = "unknown"
	}
	if config.Version == "" {
		config.Version = "dev"
	}

	loggerOptions := LoggerOptions{
		Type:            config.LoggerType,
		Level:           config.LoggerLevel,
		TimestampFormat: config.TimestampFormat,
		PrettyPrintJSON: config.PrettyPrintJSON,
	}.withDefaults()

	tracer := otel.Tracer(config.ServiceName)
	meter := otel.Meter(config.ServiceName)

	operationCounter, err := meter.Int64Counter(
		"agentic_orchestrator_operations_total",
		metric.WithDescription("total operation executions"),
	)
	if err != nil {
		return nil, err
	}

	operationDurations, err := meter.Float64Histogram(
		"agentic_orchestrator_operation_duration_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("operation execution duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	logger := NewLogrusLogger(loggerOptions)
	baseEntry := NewServiceEntry(logger, config.ServiceName, config.Environment, config.Version)
	if baseEntry == nil {
		return nil, errors.New("failed to create service log entry")
	}

	return &Bundle{
		serviceName:        config.ServiceName,
		environment:        config.Environment,
		version:            config.Version,
		loggerType:         loggerOptions.Type,
		loggerLevel:        loggerOptions.Level,
		Logger:             logger,
		baseEntry:          baseEntry,
		tracer:             tracer,
		meter:              meter,
		operationCounter:   operationCounter,
		operationDurations: operationDurations,
	}, nil
}

// LoggerType returns the configured logger output type for the bundle.
func (bundle *Bundle) LoggerType() LoggerType {
	return bundle.loggerType
}

// LoggerLevel returns the configured logrus level for the bundle.
func (bundle *Bundle) LoggerLevel() logrus.Level {
	return bundle.loggerLevel
}

// ServiceEntry returns the base service-scoped entry.
func (bundle *Bundle) ServiceEntry() *Entry {
	return bundle.baseEntry
}

// Entry returns a context-bound entry enriched with correlation and trace fields.
func (bundle *Bundle) Entry(ctx context.Context) *Entry {
	if bundle.baseEntry == nil {
		return nil
	}
	return bundle.baseEntry.WithContext(ctx)
}

// StartOperation begins an operation scope and returns an observability-native wrapper.
//
// Call End on the returned operation to close span scope and emit metrics.
func (bundle *Bundle) StartOperation(
	ctx context.Context,
	operation string,
	ids CorrelationIDs,
	extraFields logrus.Fields,
) *Operation {
	ctx = WithCorrelationIDs(ctx, ids)
	attrs := bundle.attrs(operation, ids, "started")
	ctx, span := bundle.tracer.Start(ctx, operation, trace.WithAttributes(attrs...))

	entry := bundle.Entry(ctx)
	if entry != nil {
		entry = entry.WithField("operation", operation)
		if len(extraFields) > 0 {
			entry = entry.WithFields(extraFields)
		}
	}

	return &Operation{
		ctx:       ctx,
		entry:     entry,
		span:      span,
		bundle:    bundle,
		name:      operation,
		ids:       ids,
		startedAt: time.Now(),
	}
}

// Context returns the operation context containing correlation and span scope.
func (operation *Operation) Context() context.Context {
	if operation == nil {
		return context.Background()
	}
	return operation.ctx
}

// Entry returns the wrapped logger entry for this operation scope.
func (operation *Operation) Entry() *Entry {
	if operation == nil {
		return nil
	}
	return operation.entry
}

// End closes the operation scope, emits metrics, and ends the internal span.
func (operation *Operation) End(err error) {
	if operation == nil || operation.ended {
		return
	}
	operation.ended = true

	status := "ok"
	if err != nil {
		status = "error"
	}

	attrs := operation.bundle.attrs(operation.name, operation.ids, status)
	operation.bundle.operationCounter.Add(operation.ctx, 1, metric.WithAttributes(attrs...))
	operation.bundle.operationDurations.Record(
		operation.ctx,
		time.Since(operation.startedAt).Seconds(),
		metric.WithAttributes(attrs...),
	)

	if operation.span != nil {
		if err != nil {
			operation.span.RecordError(err)
			operation.span.SetStatus(codes.Error, err.Error())
		} else {
			operation.span.SetStatus(codes.Ok, "ok")
		}
		operation.span.End()
	}
}

func (bundle *Bundle) attrs(operation string, ids CorrelationIDs, status string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", bundle.serviceName),
		attribute.String("deployment.environment", bundle.environment),
		attribute.String("service.version", bundle.version),
		attribute.String("operation", operation),
		attribute.String("status", status),
	}
	if ids.RunID != "" {
		attrs = append(attrs, attribute.String("run_id", ids.RunID))
	}
	if ids.TaskID != "" {
		attrs = append(attrs, attribute.String("task_id", ids.TaskID))
	}
	if ids.JobID != "" {
		attrs = append(attrs, attribute.String("job_id", ids.JobID))
	}
	return attrs
}
