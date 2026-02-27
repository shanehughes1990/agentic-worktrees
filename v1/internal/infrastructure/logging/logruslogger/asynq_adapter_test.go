package logruslogger

import (
	"path/filepath"
	"testing"
)

func TestAsynqAdapterMethods(t *testing.T) {
	logger, err := New("text", "debug", filepath.Join(t.TempDir(), "app.log"))
	if err != nil {
		t.Fatalf("unexpected logger error: %v", err)
	}

	adapter := NewAsynqAdapter(logger)
	adapter.Debug("debug")
	adapter.Info("info")
	adapter.Warn("warn")
	adapter.Error("error")
}
