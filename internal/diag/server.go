package diag

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/shanehughes1990/agentic-worktrees/internal/app"
	"github.com/shanehughes1990/agentic-worktrees/internal/queue"
	"github.com/shanehughes1990/agentic-worktrees/internal/runstate"
)

type Server struct {
	server      *http.Server
	metrics     *queue.Metrics
	checkpoints *runstate.Store
}

type healthResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	RedisAddr   string `json:"redis_addr"`
	AuditLog    string `json:"audit_log"`
	RedisOK     bool   `json:"redis_ok"`
}

type statusResponse struct {
	Metrics             queue.Snapshot `json:"metrics"`
	CheckpointByStatus  map[string]int `json:"checkpoint_by_status"`
	CheckpointTotalRuns int            `json:"checkpoint_total_rows"`
}

func NewServer(cfg app.Config, metrics *queue.Metrics, checkpoints *runstate.Store) *Server {
	if metrics == nil {
		metrics = queue.NewMetrics()
	}
	if checkpoints == nil {
		checkpoints = runstate.NewStore(cfg.CheckpointPath)
	}

	s := &Server{
		metrics:     metrics,
		checkpoints: checkpoints,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		defer client.Close()

		err := client.Ping(ctx).Err()
		resp := healthResponse{
			Status:      "ok",
			Environment: cfg.Env,
			RedisAddr:   cfg.RedisAddr,
			AuditLog:    cfg.AuditLogPath,
			RedisOK:     err == nil,
		}
		if err != nil {
			resp.Status = "degraded"
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		summary, err := s.checkpoints.Summary()
		if err != nil {
			http.Error(w, fmt.Sprintf("checkpoint summary error: %v", err), http.StatusInternalServerError)
			return
		}
		rows, err := s.checkpoints.All()
		if err != nil {
			http.Error(w, fmt.Sprintf("checkpoint rows error: %v", err), http.StatusInternalServerError)
			return
		}

		resp := statusResponse{
			Metrics:             s.metrics.Snapshot(),
			CheckpointByStatus:  summary,
			CheckpointTotalRuns: len(rows),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
		}
	})

	s.server = &http.Server{Addr: cfg.DiagListenAddr, Handler: mux}
	return s
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- s.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.server.Shutdown(shutdownCtx)
		return nil
	case err := <-errChan:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
