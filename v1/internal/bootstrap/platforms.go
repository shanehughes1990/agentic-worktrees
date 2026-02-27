package bootstrap

import (
	"agentic-orchestrator/internal/infrastructure/healthcheck"
	"agentic-orchestrator/internal/infrastructure/observability"
	"context"
	"fmt"
)

func bootstrapPlatforms(ctx context.Context, config BaseConfig) (*observability.Platform, *healthcheck.Platform, error) {
	logFormat := toObservabilityLogFormat(config.LogFormat)
	logLevel := toObservabilityLogLevel(config.LogLevel)

	observabilityPlatform, err := observability.Bootstrap(ctx, observability.Config{
		ServiceName:     config.ServiceName,
		Environment:     config.Environment,
		Version:         config.ServiceVersion,
		LogFormat:       logFormat,
		LogLevel:        logLevel,
		PrettyPrintJSON: config.LogPrettyJSON,
		OTLPEndpoint:    config.OTLPEndpoint,
		OTLPHeaders:     parseOTLPHeaders(config.OTLPHeaders),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap observability: %w", err)
	}

	healthPlatform := healthcheck.Bootstrap(healthcheck.Config{
		LivenessPath:  config.HealthLivePath,
		ReadinessPath: config.HealthReadyPath,
		Metadata: map[string]string{
			"service":     config.ServiceName,
			"environment": config.Environment,
			"version":     config.ServiceVersion,
		},
	})

	return observabilityPlatform, healthPlatform, nil
}

func toObservabilityLogFormat(value string) observability.LogFormat {
	switch value {
	case string(observability.LogFormatText):
		return observability.LogFormatText
	case string(observability.LogFormatJSON):
		return observability.LogFormatJSON
	default:
		return observability.LogFormatText
	}
}

func toObservabilityLogLevel(value string) observability.LogLevel {
	switch value {
	case string(observability.LogLevelDebug):
		return observability.LogLevelDebug
	case string(observability.LogLevelInfo):
		return observability.LogLevelInfo
	case string(observability.LogLevelWarn):
		return observability.LogLevelWarn
	case string(observability.LogLevelError):
		return observability.LogLevelError
	default:
		return observability.LogLevelInfo
	}
}
