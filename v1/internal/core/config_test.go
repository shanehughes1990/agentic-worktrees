package core

import "testing"

func TestLoadAppConfigFromEnv(t *testing.T) {
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("REDIS_URI", "redis://localhost:6379/0")
	t.Setenv("APP_ROOT_DIR", ".runtime")

	cfg, err := LoadAppConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected load config error: %v", err)
	}

	if cfg.Logging.Format != "json" {
		t.Fatalf("unexpected log format: %s", cfg.Logging.Format)
	}
	if cfg.Redis.URI == "" {
		t.Fatalf("expected redis uri to be set")
	}
	if root := runtimeRootDirectory(cfg); root != ".runtime" {
		t.Fatalf("unexpected runtime root directory: %s", root)
	}
}

func TestLoadAppConfigFromEnvValidationError(t *testing.T) {
	t.Setenv("REDIS_URI", "")
	_, err := LoadAppConfigFromEnv()
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestLoadAppConfigFromEnvIgnoresLogPathEnvForRuntimeLogDestination(t *testing.T) {
	t.Setenv("LOG_FILE_PATH", "logs/test.log")
	t.Setenv("REDIS_URI", "redis://localhost:6379/0")
	t.Setenv("APP_ROOT_DIR", ".custom-root")

	cfg, err := LoadAppConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected load config error: %v", err)
	}

	if got := defaultRuntimeLogFilePath(cfg); got != ".custom-root/logs/app.log" {
		t.Fatalf("unexpected runtime log path: %s", got)
	}

	if got := runtimeTaskboardsDirectory(cfg); got != ".custom-root/taskboards" {
		t.Fatalf("unexpected runtime taskboard directory: %s", got)
	}
}

func TestLoadAppConfigFromEnvRejectsAbsoluteAppRootDirectory(t *testing.T) {
	t.Setenv("REDIS_URI", "redis://localhost:6379/0")
	t.Setenv("APP_ROOT_DIR", "/tmp/runtime")

	_, err := LoadAppConfigFromEnv()
	if err == nil {
		t.Fatalf("expected runtime path validation error")
	}
}
