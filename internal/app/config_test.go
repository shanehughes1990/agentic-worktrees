package app

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("ASYNQ_QUEUE", "")
	t.Setenv("MODEL_DEFAULT", "")
	t.Setenv("AUDIT_LOG_PATH", "")
	t.Setenv("DIAG_LISTEN_ADDR", "")
	t.Setenv("BOARD_PATH", "")
	t.Setenv("CHECKPOINT_PATH", "")
	t.Setenv("WORKTREE_ROOT", "")
	t.Setenv("WATCHDOG_SECONDS", "")

	cfg := LoadConfig()

	if cfg.Env != "development" {
		t.Fatalf("expected default env, got %q", cfg.Env)
	}
	if cfg.RedisAddr != "127.0.0.1:6379" {
		t.Fatalf("expected default redis addr, got %q", cfg.RedisAddr)
	}
	if cfg.QueueName != "default" {
		t.Fatalf("expected default queue, got %q", cfg.QueueName)
	}
	if cfg.ModelDefault != defaultModel {
		t.Fatalf("expected default model %q, got %q", defaultModel, cfg.ModelDefault)
	}
	if cfg.AuditLogPath != "logs/audit.log" {
		t.Fatalf("expected default audit path, got %q", cfg.AuditLogPath)
	}
	if cfg.DiagListenAddr != ":8080" {
		t.Fatalf("expected default diag listen addr, got %q", cfg.DiagListenAddr)
	}
	if cfg.BoardPath != "state/board.json" {
		t.Fatalf("expected default board path, got %q", cfg.BoardPath)
	}
	if cfg.CheckpointPath != "state/checkpoints.json" {
		t.Fatalf("expected default checkpoint path, got %q", cfg.CheckpointPath)
	}
	if cfg.WorktreeRoot != ".worktrees" {
		t.Fatalf("expected default worktree root, got %q", cfg.WorktreeRoot)
	}
	if cfg.WatchdogSeconds != 10 {
		t.Fatalf("expected default watchdog seconds 10, got %d", cfg.WatchdogSeconds)
	}
}

func TestLoadConfigUsesEnv(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("ASYNQ_QUEUE", "critical")
	t.Setenv("MODEL_DEFAULT", "gpt-custom")
	t.Setenv("AUDIT_LOG_PATH", "tmp/audit.log")
	t.Setenv("DIAG_LISTEN_ADDR", ":9090")
	t.Setenv("BOARD_PATH", "tmp/board.json")
	t.Setenv("CHECKPOINT_PATH", "tmp/checkpoints.json")
	t.Setenv("WORKTREE_ROOT", "tmp/worktrees")
	t.Setenv("WATCHDOG_SECONDS", "21")

	cfg := LoadConfig()

	if cfg.Env != "prod" || cfg.RedisAddr != "localhost:6380" || cfg.QueueName != "critical" || cfg.ModelDefault != "gpt-custom" || cfg.AuditLogPath != "tmp/audit.log" || cfg.DiagListenAddr != ":9090" || cfg.BoardPath != "tmp/board.json" || cfg.CheckpointPath != "tmp/checkpoints.json" || cfg.WorktreeRoot != "tmp/worktrees" || cfg.WatchdogSeconds != 21 {
		t.Fatalf("env values were not loaded correctly: %+v", cfg)
	}
}

func TestValidateRejectsEmptyValues(t *testing.T) {
	cfg := Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestGetEnvUsesFallback(t *testing.T) {
	key := "UNLIKELY_KEY_123"
	_ = os.Unsetenv(key)
	if got := getEnv(key, "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}
