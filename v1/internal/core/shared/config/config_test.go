package config

import (
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
	require.Equal(suite.T(), "info", loaded.Logger.Level)
	require.Equal(suite.T(), "text", loaded.Logger.Type)
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
	suite.T().Setenv("LOG_LEVEL", "verbose")

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

func (suite *ConfigSuite) setMinimalValidEnv() {
	suite.T().Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/appdb")
	suite.T().Setenv("REDIS_URL", "redis://localhost:6379/0")
	suite.T().Setenv("OTEL_SERVICE_NAME", "service-a")
}
