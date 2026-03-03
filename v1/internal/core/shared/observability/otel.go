package observability

import (
	domainobservability "agentic-orchestrator/internal/domain/shared/observability"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type LogFormat = domainobservability.LogFormat

type LogLevel = domainobservability.LogLevel

type Config = domainobservability.Config

const (
	LogFormatText = domainobservability.LogFormatText
	LogFormatJSON = domainobservability.LogFormatJSON

	LogLevelDebug = domainobservability.LogLevelDebug
	LogLevelInfo  = domainobservability.LogLevelInfo
	LogLevelWarn  = domainobservability.LogLevelWarn
	LogLevelError = domainobservability.LogLevelError
)

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (factory *Factory) Bootstrap(ctx context.Context, config domainobservability.Config) (domainobservability.Platform, error) {
	platform, err := Bootstrap(ctx, config)
	if err != nil {
		return nil, err
	}
	return platform, nil
}

type Platform struct {
	serviceEntry   *Entry
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
}

func Bootstrap(ctx context.Context, config Config) (*Platform, error) {
	logger := logrus.New()
	if config.LogFormat == domainobservability.LogFormatJSON {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}
	logger.SetLevel(toLogrusLevel(config.LogLevel))

	resourceInstance, _ := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.DeploymentEnvironment(config.Environment),
			semconv.ServiceVersionKey.String(config.Version),
		),
	)
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithResource(resourceInstance))
	tracer := tracerProvider.Tracer(config.ServiceName)
	gootel.SetTracerProvider(tracerProvider)

	return &Platform{
		serviceEntry:   newEntry(logrus.NewEntry(logger)),
		tracerProvider: tracerProvider,
		tracer:         tracer,
	}, nil
}

func (platform *Platform) ServiceEntry() domainobservability.Entry {
	if platform == nil || platform.serviceEntry == nil {
		return newEntry(nil)
	}
	return platform.serviceEntry
}

func (platform *Platform) StartOperation(ctx context.Context, name string) domainobservability.Operation {
	if platform == nil {
		return newOperation(ctx, nil, newEntry(nil))
	}
	if platform.tracer == nil {
		return newOperation(ctx, nil, platform.serviceEntry)
	}
	contextWithSpan, span := platform.tracer.Start(ctx, name)
	entry := platform.serviceEntry
	if entry == nil {
		entry = newEntry(nil)
	}
	entry = newEntry(entry.entry.WithField("operation", name))
	return newOperation(contextWithSpan, span, entry)
}

func (platform *Platform) Shutdown(ctx context.Context) error {
	if platform == nil || platform.tracerProvider == nil {
		return nil
	}
	return platform.tracerProvider.Shutdown(ctx)
}

type Operation struct {
	ctx   context.Context
	span  trace.Span
	entry *Entry
}

func newOperation(ctx context.Context, span trace.Span, entry *Entry) *Operation {
	if ctx == nil {
		ctx = context.Background()
	}
	if entry == nil {
		entry = newEntry(nil)
	}
	return &Operation{ctx: ctx, span: span, entry: entry}
}

func (operation *Operation) Context() context.Context {
	if operation == nil || operation.ctx == nil {
		return context.Background()
	}
	return operation.ctx
}

func (operation *Operation) WithField(key string, value any) domainobservability.Operation {
	if operation == nil {
		return operation
	}
	next := *operation
	next.entry = newEntry(operation.entry.entry.WithField(key, value))
	if operation.span != nil {
		operation.span.SetAttributes(attribute.String(key, toString(value)))
	}
	return &next
}

func (operation *Operation) WithFields(fields map[string]any) domainobservability.Operation {
	if operation == nil {
		return operation
	}
	next := *operation
	next.entry = newEntry(operation.entry.entry.WithFields(logrus.Fields(fields)))
	if operation.span != nil {
		attributes := make([]attribute.KeyValue, 0, len(fields))
		for key, value := range fields {
			attributes = append(attributes, attribute.String(key, toString(value)))
		}
		operation.span.SetAttributes(attributes...)
	}
	return &next
}

func (operation *Operation) WithError(err error) domainobservability.Operation {
	if operation == nil {
		return operation
	}
	next := *operation
	next.entry = newEntry(operation.entry.entry.WithError(err))
	if operation.span != nil && err != nil {
		operation.span.RecordError(err)
		operation.span.SetStatus(codes.Error, err.Error())
	}
	return &next
}

func (operation *Operation) End(err error) {
	if operation == nil {
		return
	}
	if err != nil && operation.span != nil {
		operation.span.RecordError(err)
		operation.span.SetStatus(codes.Error, err.Error())
	}
	if operation.span != nil {
		operation.span.End()
	}
}

func toLogrusLevel(level domainobservability.LogLevel) logrus.Level {
	switch level {
	case domainobservability.LogLevelDebug:
		return logrus.DebugLevel
	case domainobservability.LogLevelWarn:
		return logrus.WarnLevel
	case domainobservability.LogLevelError:
		return logrus.ErrorLevel
	case domainobservability.LogLevelInfo:
		return logrus.InfoLevel
	default:
		return logrus.InfoLevel
	}
}

func toString(value any) string {
	return fmt.Sprintf("%v", value)
}

var _ domainobservability.Factory = (*Factory)(nil)
var _ domainobservability.Platform = (*Platform)(nil)
var _ domainobservability.Operation = (*Operation)(nil)
