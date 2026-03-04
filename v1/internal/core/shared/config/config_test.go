package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (suite *ConfigSuite) TestLoadConfigFromEnv_SuccessWithDefaults() {
	suite.setMinimalValidEnv()

	loaded, err := LoadConfigFromEnv[BaseConfig]()
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "service-a", loaded.OTEL.ServiceName)
	require.Equal(suite.T(), "local", loaded.OTEL.ServiceEnvironment)
	require.Equal(suite.T(), "/live", loaded.Health.LivePath)
	require.Equal(suite.T(), "/ready", loaded.Health.ReadyPath)
	require.Equal(suite.T(), 15*time.Second, loaded.App.ShutdownTimeout)
}

func (suite *ConfigSuite) TestLoadConfigFromEnv_ProcessError() {
	suite.setMinimalValidEnv()
	suite.T().Setenv("SHUTDOWN_TIMEOUT", "invalid-duration")

	_, err := LoadConfigFromEnv[BaseConfig]()
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "load env config")
}

func (suite *ConfigSuite) TestLoadConfigFromEnv_ValidateError() {
	suite.setMinimalValidEnv()
	suite.T().Setenv("HEALTH_LIVE_PATH", "health")

	_, err := LoadConfigFromEnv[BaseConfig]()
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "validate env config")
}

func (suite *ConfigSuite) TestNewValidator_Success() {
	validate, err := newValidator()
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), validate)
}

func (suite *ConfigSuite) TestIsValidDatabaseDSN_ValidPathDatabase() {
	require.True(suite.T(), isValidDatabaseDSN("postgres://user:pass@localhost:5432/appdb?sslmode=disable"))
}

func (suite *ConfigSuite) TestIsValidDatabaseDSN_ValidQueryDatabase() {
	require.True(suite.T(), isValidDatabaseDSN("postgres://user:pass@localhost:5432?dbname=appdb"))
}

func (suite *ConfigSuite) TestIsValidDatabaseDSN_InvalidEmpty() {
	require.False(suite.T(), isValidDatabaseDSN(""))
}

func (suite *ConfigSuite) TestIsValidDatabaseDSN_InvalidParse() {
	require.False(suite.T(), isValidDatabaseDSN("not a dsn"))
}

func (suite *ConfigSuite) TestIsValidDatabaseDSN_InvalidScheme() {
	require.False(suite.T(), isValidDatabaseDSN("mysql://user:pass@localhost:3306/appdb"))
}

func (suite *ConfigSuite) TestIsValidDatabaseDSN_InvalidNoDatabaseName() {
	require.False(suite.T(), isValidDatabaseDSN("postgres://user:pass@localhost:5432"))
}

func (suite *ConfigSuite) TestIsValidRedisURL_Valid() {
	require.True(suite.T(), isValidRedisURL("redis://localhost:6379/0"))
}

func (suite *ConfigSuite) TestIsValidRedisURL_InvalidEmpty() {
	require.False(suite.T(), isValidRedisURL(""))
}

func (suite *ConfigSuite) TestIsValidRedisURL_InvalidURI() {
	require.False(suite.T(), isValidRedisURL("http://localhost:6379"))
}

func (suite *ConfigSuite) TestLoadConfigFromEnv_RemoteStorageGCSRequiresGoogleCredentials() {
	suite.setMinimalValidEnv()
	suite.T().Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")

	_, err := LoadConfigFromEnv[BaseConfig]()
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "required_if_remote_storage_type")
}

func (suite *ConfigSuite) TestLoadConfigFromEnv_RemoteStorageGCSRequiresBucketAndCDNSettings() {
	suite.setMinimalValidEnv()
	suite.T().Setenv("GOOGLE_CLOUD_STORAGE_BUCKET", "")
	suite.T().Setenv("GOOGLE_CDN_BASE_URL", "")
	suite.T().Setenv("GOOGLE_CDN_KEY_NAME", "")
	suite.T().Setenv("GOOGLE_CDN_KEY_VALUE", "")

	_, err := LoadConfigFromEnv[BaseConfig]()
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "required_if_remote_storage_type")
}

func (suite *ConfigSuite) TestLoadConfigFromEnv_RemoteStorageGCSRequiresProjectID() {
	suite.setMinimalValidEnv()
	suite.T().Setenv("GOOGLE_CLOUD_PROJECT_ID", "")

	_, err := LoadConfigFromEnv[BaseConfig]()
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "ProjectID")
}

func (suite *ConfigSuite) TestIsValidGoogleApplicationCredentialsPath_ValidFile() {
	credentialsPath := filepath.Join(suite.T().TempDir(), "gcp-sa.json")
	require.NoError(suite.T(), os.WriteFile(credentialsPath, []byte("{}"), 0o600))
	require.True(suite.T(), isValidGoogleApplicationCredentialsPath(credentialsPath))
}

func (suite *ConfigSuite) TestIsValidGoogleApplicationCredentialsPath_InvalidMissingFile() {
	require.False(suite.T(), isValidGoogleApplicationCredentialsPath(filepath.Join(suite.T().TempDir(), "missing.json")))
}

func (suite *ConfigSuite) setMinimalValidEnv() {
	suite.T().Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/appdb")
	suite.T().Setenv("REDIS_URL", "redis://localhost:6379/0")
	suite.T().Setenv("OTEL_SERVICE_NAME", "service-a")
	suite.T().Setenv("REMOTE_STORAGE_TYPE", "gcs")
	suite.T().Setenv("REMOTE_STORAGE_BUCKET_PREFIX", "projects")
	suite.T().Setenv("GOOGLE_CLOUD_PROJECT_ID", "project-1")
	suite.T().Setenv("GOOGLE_CLOUD_STORAGE_BUCKET", "bucket-1")
	suite.T().Setenv("GOOGLE_CDN_BASE_URL", "https://cdn.example.com")
	suite.T().Setenv("GOOGLE_CDN_KEY_NAME", "k1")
	suite.T().Setenv("GOOGLE_CDN_KEY_VALUE", "YWJj")
	credentialsPath := filepath.Join(suite.T().TempDir(), "gcp-sa.json")
	require.NoError(suite.T(), os.WriteFile(credentialsPath, []byte("{}"), 0o600))
	suite.T().Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsPath)
}
