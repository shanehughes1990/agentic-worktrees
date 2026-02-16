package app

import (
	"fmt"
	"os"
	"strings"
)

const defaultModel = "gpt-5.3-codex"

type Config struct {
	Env             string
	RedisAddr       string
	QueueName       string
	ModelDefault    string
	AuditLogPath    string
	DiagListenAddr  string
	BoardPath       string
	CheckpointPath  string
	WorktreeRoot    string
	WatchdogSeconds int
}

func LoadConfig() Config {
	cfg := Config{
		Env:             getEnv("APP_ENV", "development"),
		RedisAddr:       getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		QueueName:       getEnv("ASYNQ_QUEUE", "default"),
		ModelDefault:    getEnv("MODEL_DEFAULT", defaultModel),
		AuditLogPath:    getEnv("AUDIT_LOG_PATH", "logs/audit.log"),
		DiagListenAddr:  getEnv("DIAG_LISTEN_ADDR", ":8080"),
		BoardPath:       getEnv("BOARD_PATH", "state/board.json"),
		CheckpointPath:  getEnv("CHECKPOINT_PATH", "state/checkpoints.json"),
		WorktreeRoot:    getEnv("WORKTREE_ROOT", ".worktrees"),
		WatchdogSeconds: getEnvInt("WATCHDOG_SECONDS", 10),
	}

	cfg.Env = strings.TrimSpace(cfg.Env)
	cfg.RedisAddr = strings.TrimSpace(cfg.RedisAddr)
	cfg.QueueName = strings.TrimSpace(cfg.QueueName)
	cfg.ModelDefault = strings.TrimSpace(cfg.ModelDefault)
	cfg.AuditLogPath = strings.TrimSpace(cfg.AuditLogPath)
	cfg.DiagListenAddr = strings.TrimSpace(cfg.DiagListenAddr)
	cfg.BoardPath = strings.TrimSpace(cfg.BoardPath)
	cfg.CheckpointPath = strings.TrimSpace(cfg.CheckpointPath)
	cfg.WorktreeRoot = strings.TrimSpace(cfg.WorktreeRoot)

	return cfg
}

func (c Config) Validate() error {
	if c.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR cannot be empty")
	}
	if c.QueueName == "" {
		return fmt.Errorf("ASYNQ_QUEUE cannot be empty")
	}
	if c.AuditLogPath == "" {
		return fmt.Errorf("AUDIT_LOG_PATH cannot be empty")
	}
	if c.DiagListenAddr == "" {
		return fmt.Errorf("DIAG_LISTEN_ADDR cannot be empty")
	}
	if c.ModelDefault == "" {
		return fmt.Errorf("MODEL_DEFAULT cannot be empty")
	}
	if c.BoardPath == "" {
		return fmt.Errorf("BOARD_PATH cannot be empty")
	}
	if c.CheckpointPath == "" {
		return fmt.Errorf("CHECKPOINT_PATH cannot be empty")
	}
	if c.WorktreeRoot == "" {
		return fmt.Errorf("WORKTREE_ROOT cannot be empty")
	}
	if c.WatchdogSeconds < 1 {
		return fmt.Errorf("WATCHDOG_SECONDS must be >= 1")
	}
	return nil
}

func getEnv(key string, fallback string) string {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}

func getEnvInt(key string, fallback int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	var parsed int
	_, err := fmt.Sscanf(val, "%d", &parsed)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
