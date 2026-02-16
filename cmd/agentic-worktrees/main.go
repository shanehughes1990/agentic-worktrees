package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"

	"github.com/shanehughes1990/agentic-worktrees/internal/app"
	"github.com/shanehughes1990/agentic-worktrees/internal/board"
	"github.com/shanehughes1990/agentic-worktrees/internal/diag"
	"github.com/shanehughes1990/agentic-worktrees/internal/gitops"
	"github.com/shanehughes1990/agentic-worktrees/internal/mcp"
	"github.com/shanehughes1990/agentic-worktrees/internal/queue"
	"github.com/shanehughes1990/agentic-worktrees/internal/runstate"
	"github.com/shanehughes1990/agentic-worktrees/internal/watchdog"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:  "agentic-worktrees",
		Usage: "Autonomous queue/worktree orchestrator prototype",
		Commands: []*cli.Command{
			{
				Name:  "preflight",
				Usage: "Run dependency and environment checks",
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg := app.LoadConfig()
					if err := cfg.Validate(); err != nil {
						return err
					}
					report := app.RunPreflight(ctx, cfg)
					payload, _ := json.MarshalIndent(report, "", "  ")
					fmt.Println(string(payload))
					return report.Error()
				},
			},
			{
				Name:  "ingest",
				Usage: "Build board JSON from scope file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "scope", Required: true},
					&cli.StringFlag{Name: "out", Required: false},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg := app.LoadConfig()
					if err := cfg.Validate(); err != nil {
						return err
					}
					scopePath := c.String("scope")
					outPath := c.String("out")
					if outPath == "" {
						outPath = cfg.BoardPath
					}
					built, err := board.BuildBoardFromFile(scopePath)
					if err != nil {
						return err
					}
					repo := board.NewRepository(outPath)
					if err := repo.Write(built); err != nil {
						return err
					}
					fmt.Printf("wrote board: path=%s tasks=%d\n", outPath, len(built.Tasks))
					return nil
				},
			},
			{
				Name:  "start",
				Usage: "Start lifecycle for a task by enqueueing prepare_worktree",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "run-id", Required: false},
					&cli.StringFlag{Name: "task-id", Required: true},
					&cli.StringFlag{Name: "worktree", Required: false},
					&cli.StringFlag{Name: "prompt", Required: false},
					&cli.StringFlag{Name: "origin-branch", Required: false, Value: "main"},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg := app.LoadConfig()
					if err := cfg.Validate(); err != nil {
						return err
					}

					runID := c.String("run-id")
					if runID == "" {
						runID = uuid.NewString()
					}

					task, err := queue.NewLifecycleTask(queue.TypePrepareWorktree, queue.LifecyclePayload{
						RunID:        runID,
						TaskID:       c.String("task-id"),
						WorktreeName: c.String("worktree"),
						Prompt:       c.String("prompt"),
						OriginBranch: c.String("origin-branch"),
					}, cfg.QueueName)
					if err != nil {
						return err
					}

					client := queue.NewClient(cfg.RedisAddr)
					defer client.Close()
					info, err := client.Enqueue(task)
					if err != nil {
						return err
					}
					fmt.Printf("started run: run_id=%s job_id=%s type=%s queue=%s\n", runID, info.ID, info.Type, info.Queue)
					return nil
				},
			},
			{
				Name:  "run",
				Usage: "Run worker, diagnostics server, and watchdog",
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg := app.LoadConfig()
					if err := cfg.Validate(); err != nil {
						return err
					}

					report := app.RunPreflight(ctx, cfg)
					if err := report.Error(); err != nil {
						return err
					}

					auditSink, err := app.NewAuditSink(cfg.AuditLogPath)
					if err != nil {
						return err
					}
					defer auditSink.Close()

					checkpointStore := runstate.NewStore(cfg.CheckpointPath)
					metrics := queue.NewMetrics()
					gitManager := gitops.NewManager(cfg.WorktreeRoot)

					runCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
					defer cancel()

					worker := queue.NewWorker(cfg.RedisAddr, cfg.QueueName, 5, auditSink, checkpointStore, gitManager, metrics)
					diagServer := diag.NewServer(cfg, metrics, checkpointStore)

					go watchdog.Run(runCtx, cfg.RedisAddr, time.Duration(cfg.WatchdogSeconds)*time.Second, auditSink, metrics)

					var wg sync.WaitGroup
					errCh := make(chan error, 2)

					wg.Add(1)
					go func() {
						defer wg.Done()
						errCh <- worker.Run(runCtx)
					}()

					wg.Add(1)
					go func() {
						defer wg.Done()
						errCh <- diagServer.Run(runCtx)
					}()

					select {
					case <-runCtx.Done():
						wg.Wait()
						return nil
					case err = <-errCh:
						cancel()
						wg.Wait()
						if errors.Is(err, context.Canceled) || err == nil {
							return nil
						}
						return err
					}
				},
			},
			{
				Name:  "status",
				Usage: "Print checkpoint summary from persisted state",
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg := app.LoadConfig()
					if err := cfg.Validate(); err != nil {
						return err
					}
					store := runstate.NewStore(cfg.CheckpointPath)
					summary, err := store.Summary()
					if err != nil {
						return err
					}
					rows, err := store.All()
					if err != nil {
						return err
					}
					payload, _ := json.MarshalIndent(map[string]any{"summary": summary, "rows": len(rows)}, "", "  ")
					fmt.Println(string(payload))
					return nil
				},
			},
			{
				Name:  "mcp-call",
				Usage: "Call MCP-style tool adapter for local orchestration",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "tool", Required: true},
					&cli.StringFlag{Name: "input-json", Required: false, Value: "{}"},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg := app.LoadConfig()
					if err := cfg.Validate(); err != nil {
						return err
					}

					input := map[string]any{}
					if err := json.Unmarshal([]byte(c.String("input-json")), &input); err != nil {
						return fmt.Errorf("parse input-json: %w", err)
					}

					client := queue.NewClient(cfg.RedisAddr)
					defer client.Close()

					adapter := mcp.Adapter{
						QueueClient: client,
						QueueName:   cfg.QueueName,
						Metrics:     queue.NewMetrics(),
						Checkpoints: runstate.NewStore(cfg.CheckpointPath),
					}
					result, err := adapter.Call(ctx, c.String("tool"), input)
					if err != nil {
						return err
					}
					payload, _ := json.MarshalIndent(result, "", "  ")
					fmt.Println(string(payload))
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "Print version",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println(version)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
