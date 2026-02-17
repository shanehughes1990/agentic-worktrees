package containers

import (
	"context"
	"fmt"
	"sync"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MySQLContainer struct {
	container tc.Container
	host      string
	port      string
	dbName    string
	user      string
	password  string
}

func (c *MySQLContainer) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", c.user, c.password, c.host, c.port, c.dbName)
}

var (
	mysqlOnce       sync.Once
	mysqlShared     *MySQLContainer
	mysqlSharedErr  error
	mysqlReleaseMu  sync.Mutex
	mysqlRefCount   int
)

func AcquireMySQL(ctx context.Context) (*MySQLContainer, func(context.Context) error, error) {
	mysqlOnce.Do(func() {
		mysqlShared, mysqlSharedErr = startMySQL(ctx)
	})
	if mysqlSharedErr != nil {
		return nil, nil, mysqlSharedErr
	}

	mysqlReleaseMu.Lock()
	mysqlRefCount++
	mysqlReleaseMu.Unlock()

	release := func(releaseCtx context.Context) error {
		mysqlReleaseMu.Lock()
		defer mysqlReleaseMu.Unlock()
		if mysqlRefCount > 0 {
			mysqlRefCount--
		}
		if mysqlRefCount != 0 || mysqlShared == nil {
			return nil
		}

		err := mysqlShared.container.Terminate(releaseCtx)
		mysqlShared = nil
		mysqlSharedErr = nil
		mysqlOnce = sync.Once{}
		return err
	}

	return mysqlShared, release, nil
}

func startMySQL(ctx context.Context) (*MySQLContainer, error) {
	const (
		image    = "mysql:8.0"
		dbName   = "agentic_worktrees_test"
		user     = "agentic"
		password = "agentic_pwd"
	)

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_DATABASE":      dbName,
				"MYSQL_USER":          user,
				"MYSQL_PASSWORD":      password,
				"MYSQL_ROOT_PASSWORD": "root_pwd",
			},
			WaitingFor: wait.ForListeningPort("3306/tcp").WithStartupTimeout(120 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("start mysql testcontainer: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve mysql host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve mysql port: %w", err)
	}

	return &MySQLContainer{
		container: container,
		host:      host,
		port:      mappedPort.Port(),
		dbName:    dbName,
		user:      user,
		password:  password,
	}, nil
}
