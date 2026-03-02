package healthcheck

import (
	"context"
	"net/http"
	"time"
)

const (
	defaultLivenessPath  = "/live"
	defaultReadinessPath = "/ready"
)

// CheckFunc validates a liveness/readiness dependency.
type CheckFunc func() error

// ShutdownFunc performs resource shutdown work.
type ShutdownFunc func(context.Context) error

// Config defines the exposed bootstrap contract for health checks.
type Config struct {
	LivenessPath  string
	ReadinessPath string
	Metadata      map[string]string
}

// Platform is the exposed healthcheck/shutdown surface.
type Platform struct {
	runtime *runtime
}

// Bootstrap initializes the healthcheck platform.
func Bootstrap(config Config) *Platform {
	return &Platform{runtime: newRuntime(config)}
}

// AddLivenessCheck registers a liveness check.
func (platform *Platform) AddLivenessCheck(name string, check CheckFunc) {
	if platform == nil || platform.runtime == nil {
		return
	}
	platform.runtime.addLivenessCheck(name, check)
}

// AddReadinessCheck registers a readiness check.
func (platform *Platform) AddReadinessCheck(name string, check CheckFunc) {
	if platform == nil || platform.runtime == nil {
		return
	}
	platform.runtime.addReadinessCheck(name, check)
}

// AddAsyncReadinessCheck registers a readiness check that runs asynchronously.
func (platform *Platform) AddAsyncReadinessCheck(name string, check CheckFunc, interval time.Duration) {
	if platform == nil || platform.runtime == nil {
		return
	}
	platform.runtime.addAsyncReadinessCheck(name, check, interval)
}

// RegisterShutdownHook registers a shutdown callback.
func (platform *Platform) RegisterShutdownHook(name string, hook ShutdownFunc) {
	if platform == nil || platform.runtime == nil {
		return
	}
	platform.runtime.registerShutdownHook(name, hook)
}

// RegisterDependency registers both readiness check and shutdown hook for one dependency.
func (platform *Platform) RegisterDependency(name string, readiness CheckFunc, shutdown ShutdownFunc) {
	if platform == nil || platform.runtime == nil {
		return
	}
	platform.runtime.addReadinessCheck(name, readiness)
	platform.runtime.registerShutdownHook(name, shutdown)
}

// Mount registers live/ready endpoints on the provided mux.
func (platform *Platform) Mount(mux *http.ServeMux) {
	if platform == nil || platform.runtime == nil || mux == nil {
		return
	}
	platform.runtime.mount(mux)
}

// LiveEndpoint serves liveness checks.
func (platform *Platform) LiveEndpoint(w http.ResponseWriter, r *http.Request) {
	if platform == nil || platform.runtime == nil {
		http.NotFound(w, r)
		return
	}
	platform.runtime.liveEndpoint(w, r)
}

// ReadyEndpoint serves readiness checks.
func (platform *Platform) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {
	if platform == nil || platform.runtime == nil {
		http.NotFound(w, r)
		return
	}
	platform.runtime.readyEndpoint(w, r)
}

// Shutdown executes registered shutdown hooks in reverse registration order.
func (platform *Platform) Shutdown(ctx context.Context) error {
	if platform == nil || platform.runtime == nil {
		return nil
	}
	return platform.runtime.shutdown(ctx)
}
