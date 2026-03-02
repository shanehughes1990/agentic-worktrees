package api

import (
	sharedconfig "agentic-orchestrator/internal/core/shared/config"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	sharedconfig.BaseConfig

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

	GraphQLPath      string `envconfig:"API_GRAPHQL_PATH" default:"/query" validate:"required,startswith=/"`
	PlaygroundPath   string `envconfig:"API_PLAYGROUND_PATH" default:"/" validate:"required,startswith=/"`
	EnablePlayground bool   `envconfig:"API_ENABLE_PLAYGROUND" default:"true"`
}

func LoadConfigFromEnv() (Config, error) {
	if strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")) == "" {
		_ = os.Setenv("OTEL_SERVICE_NAME", "agentic-api")
	}
	config, err := sharedconfig.LoadConfigFromEnv[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("load api env config: %w", err)
	}
	return config, nil
}
