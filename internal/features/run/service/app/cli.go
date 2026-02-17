package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"

	queuedomain "github.com/shanehughes1990/agentic-worktrees/internal/features/queue/domain"
	"github.com/shanehughes1990/agentic-worktrees/internal/features/queue/producer"
)

var version = "dev"

func (runtime *Runtime) runCLI(ctx context.Context) error {
	command := &cli.Command{
		Name:  runtime.config.Name,
		Usage: "Agentic worktrees CLI",
		Commands: []*cli.Command{
			{
				Name:  "preflight",
				Usage: "Validate runtime dependencies and configuration",
				Action: func(cmdCtx context.Context, _ *cli.Command) error {
					report := runPreflight(cmdCtx, runtime.config)
					payload, _ := json.MarshalIndent(report, "", "  ")
					fmt.Println(string(payload))
					return report.Error()
				},
			},
			{
				Name:  "ingest",
				Usage: "Enqueue ADK board planning task for a scope file/folder",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "scope", Required: true},
					&cli.StringFlag{Name: "out", Required: false},
					&cli.StringFlag{Name: "run-id", Required: false},
					&cli.StringFlag{Name: "task-id", Required: false},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					outPath := strings.TrimSpace(c.String("out"))
					if outPath == "" {
						outPath = runtime.config.BoardPath
					}

					runID := strings.TrimSpace(c.String("run-id"))
					if runID == "" {
						runID = uuid.NewString()
					}

					taskID := strings.TrimSpace(c.String("task-id"))
					if taskID == "" {
						taskID = "task-ingest-board"
					}

					client, err := producer.NewClient(runtime.config.RedisAddr, runtime.config.QueueName)
					if err != nil {
						return err
					}
					defer client.Close()

					scopePath := c.String("scope")
					payload := queuedomain.PlanBoardPayload{
						RunID:          runID,
						TaskID:         taskID,
						ScopePath:      scopePath,
						OutPath:        outPath,
						IdempotencyKey: fmt.Sprintf("%s:%s:%s:%s", runID, taskID, scopePath, outPath),
					}

					info, err := client.EnqueuePlanBoard(payload)
					if err != nil {
						return err
					}

					fmt.Printf("enqueued task: id=%s type=%s queue=%s\n", info.ID, info.Type, info.Queue)
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "Print CLI version",
				Action: func(_ context.Context, _ *cli.Command) error {
					fmt.Println(version)
					return nil
				},
			},
		},
	}

	return command.Run(ctx, os.Args)
}
