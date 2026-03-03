package api

import (
	sharedconfig "agentic-orchestrator/internal/core/shared/config"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	sharedconfig.BaseConfig

	API ApiConfig `validate:"required"`
}

type ApiConfig struct {
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
