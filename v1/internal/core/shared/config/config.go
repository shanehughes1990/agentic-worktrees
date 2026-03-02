package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/xo/dburl"
)

type BaseConfig struct {
	App    AppConfig    `validate:"required"`
	Health HealthConfig `validate:"required"`
	OTEL   OTELConfig   `validate:"required"`
}

type OTELConfig struct {
	OTLPEndpoint       string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTLPHeaders        string `envconfig:"OTEL_EXPORTER_OTLP_HEADERS"`
	ServiceEnvironment string `envconfig:"OTEL_DEPLOYMENT_ENVIRONMENT" default:"local" validate:"required"`
	ServiceName        string `envconfig:"OTEL_SERVICE_NAME" validate:"required"`
}

type AppConfig struct {
	Port            int           `envconfig:"APP_PORT" default:"8080" validate:"required,min=1,max=65535"`
	DatabaseDSN     string        `envconfig:"DATABASE_DSN" validate:"required,database_dsn"`
	RedisURL        string        `envconfig:"REDIS_URL" default:"redis://redis:6379/0" validate:"required,redis_url"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"15s" validate:"required,gt=0"`
}

type HealthConfig struct {
	LivePath  string `envconfig:"HEALTH_LIVE_PATH" default:"/live" validate:"required,startswith=/"`
	ReadyPath string `envconfig:"HEALTH_READY_PATH" default:"/ready" validate:"required,startswith=/"`
}

func LoadConfigFromEnv[T any]() (T, error) {
	var target T
	_ = godotenv.Load()
	if err := envconfig.Process("", &target); err != nil {
		return target, fmt.Errorf("load env config: %w", err)
	}
	validate, err := newValidator()
	if err != nil {
		return target, err
	}
	if err := validate.Struct(target); err != nil {
		return target, fmt.Errorf("validate env config: %w", err)
	}
	return target, nil
}

func newValidator() (*validator.Validate, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("database_dsn", func(level validator.FieldLevel) bool {
		return isValidDatabaseDSN(level.Field().String())
	}); err != nil {
		return nil, fmt.Errorf("register database_dsn validator: %w", err)
	}
	if err := validate.RegisterValidation("redis_url", func(level validator.FieldLevel) bool {
		return isValidRedisURL(level.Field().String())
	}); err != nil {
		return nil, fmt.Errorf("register redis_url validator: %w", err)
	}
	return validate, nil
}

func isValidDatabaseDSN(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	parsedURL, err := dburl.Parse(raw)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(strings.TrimSpace(parsedURL.Scheme))
	if scheme != "postgres" && scheme != "postgresql" {
		return false
	}
	databaseName := strings.TrimSpace(strings.TrimPrefix(parsedURL.EscapedPath(), "/"))
	if databaseName != "" {
		return true
	}
	query := parsedURL.Query()
	if strings.TrimSpace(query.Get("dbname")) != "" {
		return true
	}
	return false
}

func isValidRedisURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	_, err := asynq.ParseRedisURI(raw)
	return err == nil
}
