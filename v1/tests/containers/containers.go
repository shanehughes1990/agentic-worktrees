package containers

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

type ResourceBuilder func(ctx context.Context, container testcontainers.Container) (any, error)

type RequestSpec struct {
	Name          string
	Request       testcontainers.ContainerRequest
	BuildResource ResourceBuilder
}

type StartedContainer struct {
	Name      string
	Container testcontainers.Container
	Resource  any
}

func Start(ctx context.Context, requests []RequestSpec) ([]StartedContainer, error) {
	if len(requests) == 0 {
		return []StartedContainer{}, nil
	}

	started := make([]StartedContainer, 0, len(requests))
	for index, request := range requests {
		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: request.Request,
			Started:          true,
		})
		if err != nil {
			_ = Terminate(context.Background(), started)
			return nil, fmt.Errorf("start container %d (%s): %w", index, strings.TrimSpace(request.Name), err)
		}

		item := StartedContainer{Name: request.Name, Container: container}
		if request.BuildResource != nil {
			resource, buildErr := request.BuildResource(ctx, container)
			if buildErr != nil {
				_ = container.Terminate(context.Background())
				_ = Terminate(context.Background(), started)
				return nil, fmt.Errorf("build resource for container %d (%s): %w", index, strings.TrimSpace(request.Name), buildErr)
			}
			item.Resource = resource
		}

		started = append(started, item)
	}

	return started, nil
}

func Terminate(ctx context.Context, containers []StartedContainer) error {
	var terminateErr error
	for index := len(containers) - 1; index >= 0; index-- {
		container := containers[index].Container
		if container == nil {
			continue
		}
		if err := container.Terminate(ctx); err != nil && terminateErr == nil {
			terminateErr = err
		}
	}
	return terminateErr
}

func IsDockerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "cannot connect to the docker daemon") ||
		strings.Contains(message, "docker socket") ||
		strings.Contains(message, "is the docker daemon running") ||
		strings.Contains(message, "permission denied")
}
