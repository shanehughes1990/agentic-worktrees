package api

import (
	sharedconfig "agentic-orchestrator/internal/core/shared/config"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestLoadConfigFromEnv_RequiresBucketPrefix(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/appdb")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("OTEL_SERVICE_NAME", "agentic-api")
	t.Setenv("REMOTE_STORAGE_TYPE", "gcs")
	t.Setenv("REMOTE_STORAGE_BUCKET_PREFIX", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT_ID", "project-1")
	t.Setenv("GOOGLE_CLOUD_STORAGE_BUCKET", "bucket-1")
	t.Setenv("GOOGLE_CDN_BASE_URL", "https://cdn.example.com")
	t.Setenv("GOOGLE_CDN_KEY_NAME", "k1")
	t.Setenv("GOOGLE_CDN_KEY_VALUE", "YWJj")
	credentialsPath := filepath.Join(t.TempDir(), "gcp-sa.json")
	if err := os.WriteFile(credentialsPath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("create temp credentials file: %v", err)
	}
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsPath)

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected validation error for missing bucket prefix")
	}
	if !strings.Contains(err.Error(), "BucketPrefix") {
		t.Fatalf("expected bucket prefix validation error, got %v", err)
	}
}

type customValidationProbe struct {
	Value string `envconfig:"CUSTOM_VALIDATION_PROBE" validate:"required,probe_rule"`
}

func TestSharedLoadConfigFromEnv_SupportsInjectedCustomValidators(t *testing.T) {
	t.Setenv("CUSTOM_VALIDATION_PROBE", "ok")

	_, err := sharedconfig.LoadConfigFromEnv[customValidationProbe](func(validate *validator.Validate) error {
		return validate.RegisterValidation("probe_rule", func(level validator.FieldLevel) bool {
			return strings.TrimSpace(level.Field().String()) == "ok"
		})
	})
	if err != nil {
		t.Fatalf("expected custom validator registration to work, got %v", err)
	}
}

func TestLoadConfigFromEnv_ParsesGoogleApplicationCredentialsPath(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/appdb")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("OTEL_SERVICE_NAME", "agentic-api")
	t.Setenv("REMOTE_STORAGE_TYPE", "gcs")
	t.Setenv("REMOTE_STORAGE_BUCKET_PREFIX", "projects")
	t.Setenv("GOOGLE_CLOUD_PROJECT_ID", "project-1")
	t.Setenv("GOOGLE_CLOUD_STORAGE_BUCKET", "bucket-1")
	t.Setenv("GOOGLE_CDN_BASE_URL", "https://cdn.example.com")
	t.Setenv("GOOGLE_CDN_KEY_NAME", "k1")
	t.Setenv("GOOGLE_CDN_KEY_VALUE", "YWJj")
	credentialsPath := filepath.Join(t.TempDir(), "gcp-sa.json")
	if err := os.WriteFile(credentialsPath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("create temp credentials file: %v", err)
	}
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsPath)

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}
	if config.RemoteStorage.GoogleCloudStorage.ApplicationCredentialsFilePath != credentialsPath {
		t.Fatalf("expected google application credentials file path to be parsed")
	}
}

func TestLoadConfigFromEnv_RequiresGoogleApplicationCredentials(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/appdb")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("OTEL_SERVICE_NAME", "agentic-api")
	t.Setenv("REMOTE_STORAGE_TYPE", "gcs")
	t.Setenv("REMOTE_STORAGE_BUCKET_PREFIX", "projects")
	t.Setenv("GOOGLE_CLOUD_PROJECT_ID", "project-1")
	t.Setenv("GOOGLE_CLOUD_STORAGE_BUCKET", "bucket-1")
	t.Setenv("GOOGLE_CDN_BASE_URL", "https://cdn.example.com")
	t.Setenv("GOOGLE_CDN_KEY_NAME", "k1")
	t.Setenv("GOOGLE_CDN_KEY_VALUE", "YWJj")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected validation error when google application credentials is empty")
	}
	if !strings.Contains(err.Error(), "ApplicationCredentialsFilePath") {
		t.Fatalf("expected google application credentials validation error, got %v", err)
	}
}
