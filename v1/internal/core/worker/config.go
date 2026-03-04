package worker

import (
	sharedconfig "agentic-orchestrator/internal/core/shared/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	sharedconfig.BaseConfig

	Worker WorkerConfig `validate:"required"`
}

type WorkerConfig struct {
	ArtifactRootDirectory string `envconfig:"WORKER_ARTIFACT_ROOT_DIRECTORY" default:".agentic-orchestrator" validate:"required"`
	TaskConcurrencyLimit int `envconfig:"WORKER_TASK_CONCURRENCY_LIMIT" default:"10" validate:"gte=1,lte=1024"`
	ProjectSourceReconcileInterval time.Duration `envconfig:"WORKER_PROJECT_SOURCE_RECONCILE_INTERVAL" default:"15m" validate:"required,gt=0"`
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
	cleanPath := filepath.Clean(strings.TrimSpace(config.Worker.ArtifactRootDirectory))
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

func (config Config) RepositorysPath() string {
	return filepath.Join(config.ProjectPath("unscoped"), "repositories")
}

func (config Config) LogsPath() string {
	return filepath.Join(config.ApplicationRootPath(), "logs")
}

func (config Config) TrackerPath() string {
	return filepath.Join(config.ProjectPath("unscoped"), "tracker")
}
