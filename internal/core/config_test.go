package core

import "testing"

func TestLoadAppConfigFromEnv(t *testing.T) {
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FILE_PATH", "logs/test.log")
	t.Setenv("REDIS_URI", "redis://localhost:6379/0")
	t.Setenv("TASKBOARD_JSON_DIR", "data/test-taskboards")

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
}

func TestLoadAppConfigFromEnvValidationError(t *testing.T) {
	t.Setenv("REDIS_URI", "")
	_, err := LoadAppConfigFromEnv()
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
