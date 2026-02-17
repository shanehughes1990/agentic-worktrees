package containers

import (
	"context"
	"fmt"
	"sync"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisContainer struct {
	container tc.Container
	host      string
	port      string
}

func (c *RedisContainer) Addr() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}

func (c *RedisContainer) DSN() string {
	return fmt.Sprintf("redis://%s/0", c.Addr())
}

var (
	redisOnce      sync.Once
	redisShared    *RedisContainer
	redisSharedErr error
	redisReleaseMu sync.Mutex
	redisRefCount  int
)

func AcquireRedis(ctx context.Context) (*RedisContainer, func(context.Context) error, error) {
	redisOnce.Do(func() {
		redisShared, redisSharedErr = startRedis(ctx)
	})
	if redisSharedErr != nil {
		return nil, nil, redisSharedErr
	}

	redisReleaseMu.Lock()
	redisRefCount++
	redisReleaseMu.Unlock()

	release := func(releaseCtx context.Context) error {
		redisReleaseMu.Lock()
		defer redisReleaseMu.Unlock()
		if redisRefCount > 0 {
			redisRefCount--
		}
		if redisRefCount != 0 || redisShared == nil {
			return nil
		}

		err := redisShared.container.Terminate(releaseCtx)
		redisShared = nil
		redisSharedErr = nil
		redisOnce = sync.Once{}
		return err
	}

	return redisShared, release, nil
}

func startRedis(ctx context.Context) (*RedisContainer, error) {
	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(120 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("start redis testcontainer: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve redis host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("resolve redis port: %w", err)
	}

	return &RedisContainer{
		container: container,
		host:      host,
		port:      mappedPort.Port(),
	}, nil
}
