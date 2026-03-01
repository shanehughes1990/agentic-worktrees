package containers

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPostgresImage    = "postgres:16-alpine"
	defaultPostgresUser     = "postgres"
	defaultPostgresPassword = "postgres"
	defaultPostgresDatabase = "agentic_orchestrator"
)

type PostgresResource struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	DSN      string
}

func NewPostgresRequest(name string) RequestSpec {
	return RequestSpec{
		Name: name,
		Request: testcontainers.ContainerRequest{
			Image:        defaultPostgresImage,
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     defaultPostgresUser,
				"POSTGRES_PASSWORD": defaultPostgresPassword,
				"POSTGRES_DB":       defaultPostgresDatabase,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections"),
		},
		BuildResource: func(ctx context.Context, container testcontainers.Container) (any, error) {
			host, err := container.Host(ctx)
			if err != nil {
				return nil, fmt.Errorf("resolve postgres host: %w", err)
			}
			port, err := container.MappedPort(ctx, "5432/tcp")
			if err != nil {
				return nil, fmt.Errorf("resolve postgres port: %w", err)
			}
			portValue := port.Port()
			dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", defaultPostgresUser, defaultPostgresPassword, host, portValue, defaultPostgresDatabase)
			return PostgresResource{
				Host:     host,
				Port:     portValue,
				User:     defaultPostgresUser,
				Password: defaultPostgresPassword,
				Database: defaultPostgresDatabase,
				DSN:      dsn,
			}, nil
		},
	}
}
