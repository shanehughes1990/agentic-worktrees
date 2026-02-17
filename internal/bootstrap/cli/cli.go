package cli

import (
	"context"
	"fmt"
	"os"

	urcli "github.com/urfave/cli/v3"

	interfaceclicommands "github.com/shanehughes1990/agentic-worktrees/internal/interface/cli/commands"
)

type Runtime struct {
	command *urcli.Command
}

func New() (*Runtime, error) {
	return &Runtime{command: interfaceclicommands.NewRootCommand()}, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	if r == nil {
		return fmt.Errorf("runtime cannot be nil")
	}

	if r.command == nil {
		return fmt.Errorf("command cannot be nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return r.command.Run(ctx, os.Args)
}
