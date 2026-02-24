package containers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisContainer struct {
	container testcontainers.Container
	URI       string
}

func StartRedis(t *testing.T) *RedisContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	request := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	if err != nil {
		if isDockerUnavailable(err) {
			t.Skipf("docker unavailable for testcontainers redis: %v", err)
		}
		t.Fatalf("start redis container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(context.Background())
		t.Fatalf("resolve redis host: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		_ = container.Terminate(context.Background())
		t.Fatalf("resolve redis port: %v", err)
	}

	return &RedisContainer{
		container: container,
		URI:       fmt.Sprintf("redis://%s:%s/0", host, port.Port()),
	}
}

func (redisContainer *RedisContainer) Terminate(t *testing.T) {
	t.Helper()
	if redisContainer == nil || redisContainer.container == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := redisContainer.container.Terminate(ctx); err != nil {
		t.Fatalf("terminate redis container: %v", err)
	}
}

func isDockerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "cannot connect to the docker daemon") ||
		strings.Contains(message, "docker socket") ||
		strings.Contains(message, "is the docker daemon running") ||
		strings.Contains(message, "permission denied")
}
