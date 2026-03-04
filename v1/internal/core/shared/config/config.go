package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/xo/dburl"
)

type BaseConfig struct {
	App           AppConfig           `validate:"required"`
	Health        HealthConfig        `validate:"required"`
	OTEL          OTELConfig          `validate:"required"`
	RemoteStorage RemoteStorageConfig `validate:"required"`
}

type RemoteStorageConfig struct {
	Type               string                   `envconfig:"REMOTE_STORAGE_TYPE" default:"gcs" validate:"required,oneof=gcs"`
	BucketPrefix       string                   `envconfig:"REMOTE_STORAGE_BUCKET_PREFIX" default:"projects" validate:"required"`
	GoogleCloudStorage GoogleCloudStorageConfig
}

type GoogleCloudStorageConfig struct {
	ProjectID                      string `envconfig:"GOOGLE_CLOUD_PROJECT_ID"`
	Bucket                         string `envconfig:"GOOGLE_CLOUD_STORAGE_BUCKET"`
	ApplicationCredentialsFilePath string `envconfig:"GOOGLE_APPLICATION_CREDENTIALS" validate:"omitempty,google_application_credentials_path"`
	CDNBaseURL                     string `envconfig:"GOOGLE_CDN_BASE_URL"`
	CDNKeyName                     string `envconfig:"GOOGLE_CDN_KEY_NAME"`
	CDNKeyValue                    string `envconfig:"GOOGLE_CDN_KEY_VALUE"`
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
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"info" validate:"required,oneof=debug info warn error"`
	LogType         string        `envconfig:"LOG_TYPE" default:"text" validate:"required,oneof=text json"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"15s" validate:"required,gt=0"`
}

type HealthConfig struct {
	LivePath  string `envconfig:"HEALTH_LIVE_PATH" default:"/live" validate:"required,startswith=/"`
	ReadyPath string `envconfig:"HEALTH_READY_PATH" default:"/ready" validate:"required,startswith=/"`
}

type ValidatorRegistrar func(validate *validator.Validate) error

func LoadConfigFromEnv[T any](registrars ...ValidatorRegistrar) (T, error) {
	var target T
	_ = godotenv.Load()
	if err := envconfig.Process("", &target); err != nil {
		return target, fmt.Errorf("load env config: %w", err)
	}
	validate, err := newValidator(registrars...)
	if err != nil {
		return target, err
	}
	if err := validate.Struct(target); err != nil {
		return target, fmt.Errorf("validate env config: %w", err)
	}
	return target, nil
}

func newValidator(registrars ...ValidatorRegistrar) (*validator.Validate, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterStructValidation(validateRemoteStorageConfig, RemoteStorageConfig{})
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
	if err := validate.RegisterValidation("google_application_credentials_path", func(level validator.FieldLevel) bool {
		return isValidGoogleApplicationCredentialsPath(level.Field().String())
	}); err != nil {
		return nil, fmt.Errorf("register google_application_credentials_path validator: %w", err)
	}
	for _, registrar := range registrars {
		if registrar == nil {
			continue
		}
		if err := registrar(validate); err != nil {
			return nil, fmt.Errorf("register custom validator: %w", err)
		}
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

func isValidGoogleApplicationCredentialsPath(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	info, err := os.Stat(raw)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func validateRemoteStorageConfig(level validator.StructLevel) {
	config, ok := level.Current().Interface().(RemoteStorageConfig)
	if !ok {
		return
	}
	if !strings.EqualFold(strings.TrimSpace(config.Type), "gcs") {
		return
	}
	bucket := strings.TrimSpace(config.GoogleCloudStorage.Bucket)
	projectID := strings.TrimSpace(config.GoogleCloudStorage.ProjectID)
	credentialsPath := strings.TrimSpace(config.GoogleCloudStorage.ApplicationCredentialsFilePath)
	cdnBaseURL := strings.TrimSpace(config.GoogleCloudStorage.CDNBaseURL)
	cdnKeyName := strings.TrimSpace(config.GoogleCloudStorage.CDNKeyName)
	cdnKeyValue := strings.TrimSpace(config.GoogleCloudStorage.CDNKeyValue)
	if projectID == "" {
		level.ReportError(
			config.GoogleCloudStorage.ProjectID,
			"ProjectID",
			"projectID",
			"required_if_remote_storage_type",
			"gcs",
		)
	}
	if bucket == "" {
		level.ReportError(
			config.GoogleCloudStorage.Bucket,
			"Bucket",
			"bucket",
			"required_if_remote_storage_type",
			"gcs",
		)
	}
	if credentialsPath == "" {
		level.ReportError(
			config.GoogleCloudStorage,
			"GoogleCloudStorage",
			"googleCloudStorage",
			"required_if_remote_storage_type",
			"gcs",
		)
		level.ReportError(
			config.GoogleCloudStorage.ApplicationCredentialsFilePath,
			"ApplicationCredentialsFilePath",
			"applicationCredentialsFilePath",
			"required_if_remote_storage_type",
			"gcs",
		)
	}
	if credentialsPath != "" && !isValidGoogleApplicationCredentialsPath(credentialsPath) {
		level.ReportError(
			config.GoogleCloudStorage.ApplicationCredentialsFilePath,
			"ApplicationCredentialsFilePath",
			"applicationCredentialsFilePath",
			"google_application_credentials_path",
			"",
		)
	}
	if cdnBaseURL == "" {
		level.ReportError(
			config.GoogleCloudStorage.CDNBaseURL,
			"CDNBaseURL",
			"cdnBaseURL",
			"required_if_remote_storage_type",
			"gcs",
		)
	}
	if cdnKeyName == "" {
		level.ReportError(
			config.GoogleCloudStorage.CDNKeyName,
			"CDNKeyName",
			"cdnKeyName",
			"required_if_remote_storage_type",
			"gcs",
		)
	}
	if cdnKeyValue == "" {
		level.ReportError(
			config.GoogleCloudStorage.CDNKeyValue,
			"CDNKeyValue",
			"cdnKeyValue",
			"required_if_remote_storage_type",
			"gcs",
		)
	}
}
