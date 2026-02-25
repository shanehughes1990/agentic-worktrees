package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	Application ApplicationConfig
	Logging     LoggingConfig
	Redis       RedisConfig
	Copilot     CopilotConfig
}

type ApplicationConfig struct {
	RootDirectory string `envconfig:"APP_ROOT_DIR" default:".worktree" validate:"required"`
}

type LoggingConfig struct {
	Format   string `envconfig:"LOG_FORMAT" default:"text" validate:"required,oneof=text json"`
	Level    string `envconfig:"LOG_LEVEL" default:"info" validate:"required,oneof=debug info warn error fatal panic"`
}

type RedisConfig struct {
	URI string `envconfig:"REDIS_URI" validate:"required"`
}

type CopilotConfig struct {
	Model             string   `envconfig:"COPILOT_MODEL"`
	GitHubToken       string   `envconfig:"GITHUB_TOKEN"`
	CLIPath           string   `envconfig:"COPILOT_CLI_PATH"`
	CLIURL            string   `envconfig:"COPILOT_CLI_URL"`
	AuthStatusCommand string   `envconfig:"COPILOT_AUTH_STATUS_COMMAND" default:"copilot auth status"`
	AuthLoginCommand  string   `envconfig:"COPILOT_AUTH_LOGIN_COMMAND" default:"copilot auth login"`
	SkillDirectories  []string `envconfig:"COPILOT_SKILL_DIRECTORIES"`
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
	if err := validateRuntimePaths(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateRuntimePaths(cfg *AppConfig) error {
	cleanRootDir := runtimeRootDirectory(cfg)
	if cleanRootDir == "" {
		return fmt.Errorf("validate app config: APP_ROOT_DIR is required")
	}
	if strings.HasPrefix(cleanRootDir, "/") {
		return fmt.Errorf("validate app config: APP_ROOT_DIR must be repo-relative")
	}
	if cleanRootDir == "." || cleanRootDir == ".." || strings.HasPrefix(cleanRootDir, "../") {
		return fmt.Errorf("validate app config: APP_ROOT_DIR must be a subdirectory under repository root")
	}

	return nil
}

func runtimeRootDirectory(cfg *AppConfig) string {
	if cfg == nil {
		return ".worktree"
	}
	cleanRootDir := filepath.ToSlash(filepath.Clean(strings.TrimSpace(cfg.Application.RootDirectory)))
	if cleanRootDir == "" {
		return ".worktree"
	}
	return strings.TrimSuffix(cleanRootDir, "/")
}

func runtimeLogsDirectory(cfg *AppConfig) string {
	return filepath.ToSlash(filepath.Join(runtimeRootDirectory(cfg), "logs"))
}

func runtimeTaskboardsDirectory(cfg *AppConfig) string {
	return filepath.ToSlash(filepath.Join(runtimeRootDirectory(cfg), "taskboards"))
}

func runtimeWorkflowsDirectory(cfg *AppConfig) string {
	return filepath.ToSlash(filepath.Join(runtimeRootDirectory(cfg), "workflows"))
}

func runtimeWorktreesDirectory(cfg *AppConfig) string {
	return filepath.ToSlash(filepath.Join(runtimeRootDirectory(cfg), "worktrees"))
}

func defaultRuntimeLogFilePath(cfg *AppConfig) string {
	return filepath.ToSlash(filepath.Join(runtimeLogsDirectory(cfg), "app.log"))
}
