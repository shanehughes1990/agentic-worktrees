package observability

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type platformRuntime struct {
	serviceEntry      *entryRuntime
	tracer            trace.Tracer
	traceProvider     *sdktrace.TracerProvider
	metricProvider    *sdkmetric.MeterProvider
	operationCounter  otelmetric.Int64Counter
	durationHistogram otelmetric.Float64Histogram
	serviceName       string
	environment       string
	version           string
}

type operationRuntime struct {
	ctx         context.Context
	entry       *entryRuntime
	span        trace.Span
	startedAt   time.Time
	platform    *platformRuntime
	name        string
	correlation CorrelationIDs
}

func newPlatformRuntime(ctx context.Context, config Config) (*platformRuntime, error) {
	if strings.TrimSpace(config.ServiceName) == "" {
		return nil, errors.New("service name is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	providers, err := newProviders(config)
	if err != nil {
		return nil, err
	}

	logger := buildLogger(loggerOptionsFromConfig(config))
	serviceEntry := buildServiceEntry(logger, config.ServiceName, config.Environment, config.Version)
	if serviceEntry == nil {
		_ = providers.TraceProvider.Shutdown(ctx)
		_ = providers.MetricProvider.Shutdown(ctx)
		return nil, errors.New("create service log entry")
	}

	tracer := providers.TraceProvider.Tracer(config.ServiceName)
	meter := providers.MetricProvider.Meter(config.ServiceName)

	operationCounter, err := meter.Int64Counter(
		"taskboard.operation.count",
		otelmetric.WithDescription("Counts taskboard operations"),
	)
	if err != nil {
		_ = providers.TraceProvider.Shutdown(ctx)
		_ = providers.MetricProvider.Shutdown(ctx)
		return nil, fmt.Errorf("create operation counter: %w", err)
	}

	durationHistogram, err := meter.Float64Histogram(
		"taskboard.operation.duration_ms",
		otelmetric.WithDescription("Records taskboard operation duration in milliseconds"),
		otelmetric.WithUnit("ms"),
	)
	if err != nil {
		_ = providers.TraceProvider.Shutdown(ctx)
		_ = providers.MetricProvider.Shutdown(ctx)
		return nil, fmt.Errorf("create duration histogram: %w", err)
	}

	return &platformRuntime{
		serviceEntry:      serviceEntry,
		tracer:            tracer,
		traceProvider:     providers.TraceProvider,
		metricProvider:    providers.MetricProvider,
		operationCounter:  operationCounter,
		durationHistogram: durationHistogram,
		serviceName:       config.ServiceName,
		environment:       config.Environment,
		version:           config.Version,
	}, nil
}

func (platform *platformRuntime) startOperation(ctx context.Context, name string) *operationRuntime {
	if platform == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	correlation := CorrelationIDsFromContext(ctx)
	ctx = WithCorrelationIDs(ctx, correlation)
	ctx, span := platform.tracer.Start(ctx, name)

	entry := platform.serviceEntry.withContext(ctx)
	if entry != nil {
		entry = entry.withField("operation", name)
	}

	return &operationRuntime{
		ctx:         ctx,
		entry:       entry,
		span:        span,
		startedAt:   time.Now(),
		platform:    platform,
		name:        name,
		correlation: correlation,
	}
}

func (platform *platformRuntime) shutdown(ctx context.Context) error {
	if platform == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var shutdownErr error
	if platform.metricProvider != nil {
		if err := platform.metricProvider.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown metric provider: %w", err))
		}
	}
	if platform.traceProvider != nil {
		if err := platform.traceProvider.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown trace provider: %w", err))
		}
	}
	return shutdownErr
}

func (operation *operationRuntime) end(err error) {
	if operation == nil || operation.span == nil || operation.platform == nil {
		return
	}

	status := "ok"
	if err != nil {
		status = "error"
		operation.span.RecordError(err)
		operation.span.SetStatus(codes.Error, err.Error())
	}

	attributes := operation.platform.attributes(operation.name, operation.correlation, status)
	operation.span.SetAttributes(attributes...)

	duration := float64(time.Since(operation.startedAt).Milliseconds())
	operation.platform.recordMetrics(operation.ctx, duration, attributes)
	operation.span.End()
}

func (platform *platformRuntime) attributes(operation string, correlation CorrelationIDs, status string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("status", status),
	}
	if correlation.RunID != "" {
		attrs = append(attrs, attribute.String("run_id", correlation.RunID))
	}
	if correlation.TaskID != "" {
		attrs = append(attrs, attribute.String("task_id", correlation.TaskID))
	}
	if correlation.JobID != "" {
		attrs = append(attrs, attribute.String("job_id", correlation.JobID))
	}
	if platform != nil {
		if platform.serviceName != "" {
			attrs = append(attrs, semconv.ServiceName(platform.serviceName))
		}
		if platform.environment != "" {
			attrs = append(attrs, semconv.DeploymentEnvironment(platform.environment))
		}
		if platform.version != "" {
			attrs = append(attrs, semconv.ServiceVersion(platform.version))
		}
	}
	return attrs
}

func (platform *platformRuntime) recordMetrics(ctx context.Context, durationMs float64, attributes []attribute.KeyValue) {
	if platform == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if platform.operationCounter != nil {
		platform.operationCounter.Add(ctx, 1, otelmetric.WithAttributes(attributes...))
	}
	if platform.durationHistogram != nil {
		platform.durationHistogram.Record(ctx, durationMs, otelmetric.WithAttributes(attributes...))
	}
}

type providers struct {
	TraceProvider  *sdktrace.TracerProvider
	MetricProvider *sdkmetric.MeterProvider
}

func newProviders(cfg Config) (*providers, error) {
	resAttrs := []attribute.KeyValue{semconv.ServiceName(cfg.ServiceName)}
	if cfg.Environment != "" {
		resAttrs = append(resAttrs, semconv.DeploymentEnvironment(cfg.Environment))
	}
	if cfg.Version != "" {
		resAttrs = append(resAttrs, semconv.ServiceVersion(cfg.Version))
	}

	res, err := resource.New(
		context.Background(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(resAttrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	endpoint := strings.TrimSpace(cfg.OTLPEndpoint)
	if endpoint == "" {
		traceExporter := tracetest.NewInMemoryExporter()
		traceProvider := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(res),
			sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(traceExporter)),
		)

		metricReader := sdkmetric.NewManualReader()
		metricProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(metricReader),
		)

		return &providers{TraceProvider: traceProvider, MetricProvider: metricProvider}, nil
	}

	traceOptions := []otlptracehttp.Option{otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithInsecure()}
	metricOptions := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(endpoint), otlpmetrichttp.WithInsecure()}
	if len(cfg.OTLPHeaders) > 0 {
		traceOptions = append(traceOptions, otlptracehttp.WithHeaders(cfg.OTLPHeaders))
		metricOptions = append(metricOptions, otlpmetrichttp.WithHeaders(cfg.OTLPHeaders))
	}

	traceExporter, err := otlptracehttp.New(context.Background(), traceOptions...)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}
	metricExporter, err := otlpmetrichttp.New(context.Background(), metricOptions...)
	if err != nil {
		return nil, fmt.Errorf("create otlp metric exporter: %w", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	metricProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(2*time.Second), sdkmetric.WithTimeout(5*time.Second))),
	)

	return &providers{TraceProvider: traceProvider, MetricProvider: metricProvider}, nil
}
