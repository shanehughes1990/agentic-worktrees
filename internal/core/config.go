package core

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	Logging   LoggingConfig
	Redis     RedisConfig
	Taskboard TaskboardConfig
}

type LoggingConfig struct {
	Format   string `envconfig:"LOG_FORMAT" default:"text" validate:"required,oneof=text json"`
	Level    string `envconfig:"LOG_LEVEL" default:"info" validate:"required,oneof=debug info warn error fatal panic"`
	FilePath string `envconfig:"LOG_FILE_PATH" default:"logs/app.log" validate:"required"`
}

type RedisConfig struct {
	URI string `envconfig:"REDIS_URI" validate:"required"`
}

type TaskboardConfig struct {
	JSONDirectory string `envconfig:"TASKBOARD_JSON_DIR" default:"data/taskboards" validate:"required"`
}

func LoadAppConfigFromEnv() (*AppConfig, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env file: %w", err)
	}

	cfg := &AppConfig{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("load app config from environment: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("validate app config: %w", err)
	}

	return cfg, nil
}
