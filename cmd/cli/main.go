package main

import (
	"context"
	"log"

	bootstrapcli "github.com/shanehughes1990/agentic-worktrees/internal/bootstrap/cli"
)

func main() {
	runtime, err := bootstrapcli.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := runtime.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
