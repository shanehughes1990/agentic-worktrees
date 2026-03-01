package bootstrap

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/xo/dburl"
)

type BaseConfig struct {
	ServiceName              string        `envconfig:"SERVICE_NAME" default:"agentic-orchestrator" validate:"required"`
	Environment              string        `envconfig:"APP_ENV" default:"local" validate:"required,oneof=local development test staging production"`
	ServiceVersion           string        `envconfig:"SERVICE_VERSION" default:"development" validate:"required"`
	ApplicationRootDirectory string        `envconfig:"APPLICATION_ROOT_DIRECTORY" default:".agentic-orchestrator" validate:"required"`
	LogFormat                string        `envconfig:"LOG_FORMAT" default:"text" validate:"required,oneof=text json"`
	LogLevel                 string        `envconfig:"LOG_LEVEL" default:"info" validate:"required,oneof=debug info warn error"`
	LogPrettyJSON            bool          `envconfig:"LOG_PRETTY_JSON" default:"false"`
	OTLPEndpoint             string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTLPHeaders              string        `envconfig:"OTEL_EXPORTER_OTLP_HEADERS"`
	HealthLivePath           string        `envconfig:"HEALTH_LIVE_PATH" default:"/live" validate:"required,startswith=/"`
	HealthReadyPath          string        `envconfig:"HEALTH_READY_PATH" default:"/ready" validate:"required,startswith=/"`
	ShutdownTimeout          time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"15s" validate:"required,gt=0"`
	DatabaseDSN              string        `envconfig:"DATABASE_DSN" validate:"required"`

	TaskEngineBackend        string `envconfig:"TASK_ENGINE_BACKEND" default:"asynq" validate:"required,oneof=asynq"`
	TaskEngineRedisAddress   string `envconfig:"TASK_ENGINE_REDIS_ADDRESS" default:"127.0.0.1:6379" validate:"required"`
	TaskEngineRedisPassword  string `envconfig:"TASK_ENGINE_REDIS_PASSWORD"`
	TaskEngineRedisDatabase  int    `envconfig:"TASK_ENGINE_REDIS_DATABASE" default:"0" validate:"gte=0"`
	TaskEngineConcurrency    int    `envconfig:"TASK_ENGINE_CONCURRENCY" default:"10" validate:"gte=1,lte=1024"`
	TaskEngineIngestionQueue string `envconfig:"TASK_ENGINE_INGESTION_QUEUE" default:"ingestion" validate:"required"`
	TaskEngineSCMQueue       string `envconfig:"TASK_ENGINE_SCM_QUEUE" default:"scm" validate:"required"`

	SCMProvider         string `envconfig:"SCM_PROVIDER" default:"github" validate:"required,oneof=github"`
	SCMGitHubToken      string `envconfig:"SCM_GITHUB_TOKEN"`
	SCMGitHubAPIBaseURL string `envconfig:"SCM_GITHUB_API_BASE_URL" default:"https://api.github.com" validate:"required,url"`
}

func (config BaseConfig) ApplicationRootPath() string {
	cleanPath := filepath.Clean(strings.TrimSpace(config.ApplicationRootDirectory))
	if cleanPath == "." || cleanPath == "" {
		return ".agentic-orchestrator"
	}
	return cleanPath
}

func (config BaseConfig) RepositoriesPath() string {
	return filepath.Join(config.ApplicationRootPath(), "repositories")
}

func (config BaseConfig) RepositorySourcePath() string {
	return filepath.Join(config.RepositoriesPath(), "source")
}

func (config BaseConfig) WorktreesPath() string {
	return filepath.Join(config.ApplicationRootPath(), "worktrees")
}

func (config BaseConfig) LogsPath() string {
	return filepath.Join(config.ApplicationRootPath(), "logs")
}

func (config BaseConfig) TrackerPath() string {
	return filepath.Join(config.ApplicationRootPath(), "tracker")
}

type APIConfig struct {
	BaseConfig
	APIPort          int    `envconfig:"API_PORT" default:"8080" validate:"required,min=1,max=65535"`
	GraphQLPath      string `envconfig:"API_GRAPHQL_PATH" default:"/query" validate:"required,startswith=/"`
	PlaygroundPath   string `envconfig:"API_PLAYGROUND_PATH" default:"/" validate:"required,startswith=/"`
	EnablePlayground bool   `envconfig:"API_ENABLE_PLAYGROUND" default:"true"`
}

type WorkerConfig struct {
	BaseConfig
	WorkerPort              int           `envconfig:"WORKER_PORT" default:"8081" validate:"required,min=1,max=65535"`
	WorkerHeartbeatInterval time.Duration `envconfig:"WORKER_HEARTBEAT_INTERVAL" default:"15s" validate:"required,gt=0"`
}

func LoadAPIConfigFromEnv() (APIConfig, error) {
	var config APIConfig
	if err := envconfig.Process("", &config); err != nil {
		return APIConfig{}, fmt.Errorf("load api env config: %w", err)
	}
	if err := validator.New().Struct(config); err != nil {
		return APIConfig{}, fmt.Errorf("validate api env config: %w", err)
	}
	if err := validateDatabaseDSN(config.DatabaseDSN); err != nil {
		return APIConfig{}, err
	}
	return config, nil
}

func LoadWorkerConfigFromEnv() (WorkerConfig, error) {
	var config WorkerConfig
	if err := envconfig.Process("", &config); err != nil {
		return WorkerConfig{}, fmt.Errorf("load worker env config: %w", err)
	}
	if err := validator.New().Struct(config); err != nil {
		return WorkerConfig{}, fmt.Errorf("validate worker env config: %w", err)
	}
	if err := validateDatabaseDSN(config.DatabaseDSN); err != nil {
		return WorkerConfig{}, err
	}
	return config, nil
}

func validateDatabaseDSN(raw string) error {
	parsedURL, err := dburl.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("validate database dsn: %w", err)
	}
	scheme := strings.ToLower(strings.TrimSpace(parsedURL.Scheme))
	if scheme != "postgres" && scheme != "postgresql" {
		return fmt.Errorf("validate database dsn: expected postgres scheme, got %q", parsedURL.Scheme)
	}
	return nil
}

func parseOTLPHeaders(raw string) map[string]string {
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}
	}
	headers := map[string]string{}
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		if key == "" {
			continue
		}
		headers[key] = value
	}
	return headers
}
