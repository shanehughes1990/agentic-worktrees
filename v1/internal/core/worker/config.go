package worker

import (
	sharedconfig "agentic-orchestrator/internal/core/shared/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	sharedconfig.BaseConfig

	ApplicationRootDirectory string `envconfig:"APPLICATION_ROOT_DIRECTORY" default:".agentic-orchestrator" validate:"required"`

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

func LoadConfigFromEnv() (Config, error) {
	if strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")) == "" {
		_ = os.Setenv("OTEL_SERVICE_NAME", "agentic-worker")
	}
	config, err := sharedconfig.LoadConfigFromEnv[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("load worker env config: %w", err)
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

func (config Config) ProjectsPath() string {
	return filepath.Join(config.ApplicationRootPath(), "projects")
}

func (config Config) ProjectPath(projectID string) string {
	trimmedID := strings.TrimSpace(projectID)
	if trimmedID == "" {
		trimmedID = "unscoped"
	}
	return filepath.Join(config.ProjectsPath(), trimmedID)
}

func (config Config) RepositoriesPath() string {
	return filepath.Join(config.ProjectPath("unscoped"), "repositories")
}

func (config Config) RepositorySourcePath() string {
	return filepath.Join(config.RepositoriesPath(), "source")
}

func (config Config) WorktreesPath() string {
	return filepath.Join(config.ProjectPath("unscoped"), "worktrees")
}

func (config Config) LogsPath() string {
	return filepath.Join(config.ApplicationRootPath(), "logs")
}

func (config Config) TrackerPath() string {
	return filepath.Join(config.ProjectPath("unscoped"), "tracker")
}
