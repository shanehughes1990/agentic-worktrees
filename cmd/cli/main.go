package main

import (
	"context"
	"log"

	"github.com/shanehughes1990/agentic-worktrees/internal/features/run/service/app"
)

func main() {
	runtime, err := app.Init(app.KindCLI)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(context.Background(), runtime); err != nil {
		log.Fatal(err)
	}
}
