package containers

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultRedisImage = "redis:7-alpine"
)

type RedisResource struct {
	Host string
	Port string
	Addr string
	URI  string
}

func NewRedisRequest(name string) RequestSpec {
	return RequestSpec{
		Name: name,
		Request: testcontainers.ContainerRequest{
			Image:        defaultRedisImage,
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		BuildResource: func(ctx context.Context, container testcontainers.Container) (any, error) {
			host, err := container.Host(ctx)
			if err != nil {
				return nil, fmt.Errorf("resolve redis host: %w", err)
			}
			port, err := container.MappedPort(ctx, "6379/tcp")
			if err != nil {
				return nil, fmt.Errorf("resolve redis port: %w", err)
			}
			portValue := port.Port()
			addr := fmt.Sprintf("%s:%s", host, portValue)
			return RedisResource{
				Host: host,
				Port: portValue,
				Addr: addr,
				URI:  fmt.Sprintf("redis://%s/0", addr),
			}, nil
		},
	}
}
