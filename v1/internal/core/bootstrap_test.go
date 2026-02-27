package core

import (
	"reflect"
	"testing"
	"unsafe"

	filesystemsource "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/taskboard/filesystemsource"
)

func TestInitRuntime(t *testing.T) {
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_URI", "redis://localhost:6379/0")
	t.Setenv("APP_ROOT_DIR", ".worktree-test")

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

func TestInitRuntimeRegistersFilesystemAsOnlyConcreteSourceProvider(t *testing.T) {
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REDIS_URI", "redis://localhost:6379/0")
	t.Setenv("APP_ROOT_DIR", ".worktree-test")

	runtime, err := Init()
	if err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}
	if runtime == nil || runtime.ingestionCommand == nil {
		t.Fatalf("expected ingestion runtime dependency to be initialized")
	}

	ingestionServiceValue := reflect.ValueOf(runtime.ingestionCommand).Elem()
	sourceLister := unexportedFieldValue(ingestionServiceValue.FieldByName("sourceLister"))
	sourceReader := unexportedFieldValue(ingestionServiceValue.FieldByName("sourceReader"))
	if sourceLister.IsNil() || sourceReader.IsNil() {
		t.Fatalf("expected both source lister and reader dependencies to be configured")
	}

	listerValue := sourceLister.Elem()
	readerValue := sourceReader.Elem()
	expectedType := reflect.TypeOf(&filesystemsource.Adapter{})
	if listerValue.Type() != expectedType {
		t.Fatalf("expected filesystem source lister, got %s", listerValue.Type())
	}
	if readerValue.Type() != expectedType {
		t.Fatalf("expected filesystem source reader, got %s", readerValue.Type())
	}
	if listerValue.Pointer() != readerValue.Pointer() {
		t.Fatalf("expected a single filesystem provider instance wired for both listing and reading")
	}
}

func unexportedFieldValue(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}
