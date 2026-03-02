package healthcheck

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	health "github.com/hellofresh/health-go/v5"
)

type runtime struct {
	liveness       *health.Health
	readiness      *health.Health
	livenessPath   string
	readinessPath  string
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc

	mu    sync.Mutex
	hooks []registeredHook
	wg    sync.WaitGroup
}

type registeredHook struct {
	name string
	hook ShutdownFunc
}

type asyncReadinessState struct {
	mu  sync.RWMutex
	err error
}

func (state *asyncReadinessState) set(err error) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.err = err
}

func (state *asyncReadinessState) get() error {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.err
}

func newRuntime(config Config) *runtime {
	livenessPath := config.LivenessPath
	if livenessPath == "" {
		livenessPath = defaultLivenessPath
	}
	readinessPath := config.ReadinessPath
	if readinessPath == "" {
		readinessPath = defaultReadinessPath
	}

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	return &runtime{
		liveness:       newHealthContainer(config.Metadata),
		readiness:      newHealthContainer(config.Metadata),
		livenessPath:   livenessPath,
		readinessPath:  readinessPath,
		shutdownCtx:    shutdownCtx,
		shutdownCancel: shutdownCancel,
		hooks:          make([]registeredHook, 0),
	}
}

func newHealthContainer(metadata map[string]string) *health.Health {
	component := health.Component{}
	if metadata != nil {
		component.Name = metadata["service"]
		if component.Name == "" {
			component.Name = metadata["name"]
		}
		component.Version = metadata["version"]
	}

	h, err := health.New(health.WithComponent(component))
	if err == nil {
		return h
	}

	h, fallbackErr := health.New()
	if fallbackErr == nil {
		return h
	}

	return &health.Health{}
}

func (runtime *runtime) addLivenessCheck(name string, check CheckFunc) {
	if runtime == nil || runtime.liveness == nil || check == nil || name == "" {
		return
	}
	_ = runtime.liveness.Register(health.Config{
		Name: name,
		Check: func(context.Context) error {
			return check()
		},
	})
}

func (runtime *runtime) addReadinessCheck(name string, check CheckFunc) {
	if runtime == nil || runtime.readiness == nil || check == nil || name == "" {
		return
	}
	_ = runtime.readiness.Register(health.Config{
		Name: name,
		Check: func(context.Context) error {
			return check()
		},
	})
}

func (runtime *runtime) addAsyncReadinessCheck(name string, check CheckFunc, interval time.Duration) {
	if runtime == nil || runtime.readiness == nil || check == nil || name == "" {
		return
	}
	if interval <= 0 {
		interval = time.Second
	}

	state := &asyncReadinessState{}
	state.set(check())

	if err := runtime.readiness.Register(health.Config{
		Name: name,
		Check: func(context.Context) error {
			return state.get()
		},
	}); err != nil {
		return
	}

	runtime.wg.Add(1)
	go func() {
		defer runtime.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-runtime.shutdownCtx.Done():
				return
			case <-ticker.C:
				state.set(check())
			}
		}
	}()
}

func (runtime *runtime) registerShutdownHook(name string, hook ShutdownFunc) {
	if runtime == nil || hook == nil || name == "" {
		return
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	runtime.hooks = append(runtime.hooks, registeredHook{name: name, hook: hook})
}

func (runtime *runtime) mount(mux *http.ServeMux) {
	if runtime == nil || mux == nil {
		return
	}
	mux.Handle(runtime.livenessPath, runtime.liveness.Handler())
	mux.Handle(runtime.readinessPath, runtime.readiness.Handler())
}

func (runtime *runtime) liveEndpoint(w http.ResponseWriter, r *http.Request) {
	if runtime == nil || runtime.liveness == nil {
		http.NotFound(w, r)
		return
	}
	runtime.liveness.Handler().ServeHTTP(w, r)
}

func (runtime *runtime) readyEndpoint(w http.ResponseWriter, r *http.Request) {
	if runtime == nil || runtime.readiness == nil {
		http.NotFound(w, r)
		return
	}
	runtime.readiness.Handler().ServeHTTP(w, r)
}

func (runtime *runtime) shutdown(ctx context.Context) error {
	if runtime == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if runtime.shutdownCancel != nil {
		runtime.shutdownCancel()
	}
	runtime.wg.Wait()

	runtime.mu.Lock()
	hooks := make([]registeredHook, len(runtime.hooks))
	copy(hooks, runtime.hooks)
	runtime.mu.Unlock()

	var shutdownErr error
	for index := len(hooks) - 1; index >= 0; index-- {
		if err := hooks[index].hook(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}
	return shutdownErr
}
