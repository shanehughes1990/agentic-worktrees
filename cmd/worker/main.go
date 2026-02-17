package main

import (
	"context"
	"log"

	bootstrapworker "github.com/shanehughes1990/agentic-worktrees/internal/bootstrap/worker"
)

func main() {
	runtime, err := bootstrapworker.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := runtime.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
