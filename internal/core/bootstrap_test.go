package core

import (
	"path/filepath"
	"testing"
)

func TestInitRuntime(t *testing.T) {
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FILE_PATH", filepath.Join(t.TempDir(), "app.log"))
	t.Setenv("REDIS_URI", "redis://localhost:6379/0")
	t.Setenv("TASKBOARD_JSON_DIR", filepath.Join(t.TempDir(), "taskboards"))

	runtime, err := Init()
	if err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}
	if runtime == nil {
		t.Fatalf("expected runtime instance")
	}
	if runtime.worker == nil || runtime.ui == nil || runtime.taskboardService == nil {
		t.Fatalf("expected runtime dependencies to be initialized")
	}
}
