package api

import (
	sharedconfig "agentic-orchestrator/internal/core/shared/config"
	"fmt"
	"path/filepath"
	"strings"
)

type Config struct {
	sharedconfig.BaseConfig

	Environment              string `envconfig:"APP_ENV" default:"local" validate:"required,oneof=local development test staging production"`
	ServiceVersion           string `envconfig:"SERVICE_VERSION" default:"development" validate:"required"`
	ApplicationRootDirectory string `envconfig:"APPLICATION_ROOT_DIRECTORY" default:".agentic-orchestrator" validate:"required"`
	LogPrettyJSON            bool   `envconfig:"LOG_PRETTY_JSON" default:"false"`

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

	APIPort          int    `envconfig:"API_PORT" default:"8080" validate:"required,min=1,max=65535"`
	GraphQLPath      string `envconfig:"API_GRAPHQL_PATH" default:"/query" validate:"required,startswith=/"`
	PlaygroundPath   string `envconfig:"API_PLAYGROUND_PATH" default:"/" validate:"required,startswith=/"`
	EnablePlayground bool   `envconfig:"API_ENABLE_PLAYGROUND" default:"true"`
}

func LoadConfigFromEnv() (Config, error) {
	config, err := sharedconfig.LoadConfigFromEnv[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("load api env config: %w", err)
	}
	return config, nil
}

func (config Config) ApplicationRootPath() string {
	cleanPath := filepath.Clean(strings.TrimSpace(config.ApplicationRootDirectory))
	if cleanPath == "." || cleanPath == "" {
		return ".agentic-orchestrator"
	}
	return cleanPath
}

func (config Config) RepositoriesPath() string {
	return filepath.Join(config.ApplicationRootPath(), "repositories")
}

func (config Config) RepositorySourcePath() string {
	return filepath.Join(config.RepositoriesPath(), "source")
}

func (config Config) WorktreesPath() string {
	return filepath.Join(config.ApplicationRootPath(), "worktrees")
}

func (config Config) LogsPath() string {
	return filepath.Join(config.ApplicationRootPath(), "logs")
}

func (config Config) TrackerPath() string {
	return filepath.Join(config.ApplicationRootPath(), "tracker")
}
