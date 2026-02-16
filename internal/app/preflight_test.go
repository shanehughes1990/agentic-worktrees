package app

import (
	"context"
	"testing"
)

func TestRunPreflightReportsUnhealthyForBadRedis(t *testing.T) {
	cfg := Config{
		Env:            "test",
		RedisAddr:      "127.0.0.1:1",
		QueueName:      "default",
		ModelDefault:   "gpt-5.3-codex",
		AuditLogPath:   "logs/audit.log",
		DiagListenAddr: ":8080",
	}

	report := RunPreflight(context.Background(), cfg)
	if report.Healthy {
		t.Fatalf("expected unhealthy preflight report")
	}

	foundRedisFailure := false
	for _, check := range report.Checks {
		if check.Name == "redis" && !check.OK {
			foundRedisFailure = true
			break
		}
	}

	if !foundRedisFailure {
		t.Fatalf("expected redis failure check")
	}

	if report.Error() == nil {
		t.Fatalf("expected preflight report error")
	}
}

func TestCheckModel(t *testing.T) {
	if check := checkModel(""); check.OK {
		t.Fatalf("expected model check to fail for empty value")
	}
	if check := checkModel("gpt-5.3-codex"); !check.OK {
		t.Fatalf("expected model check success")
	}
}
