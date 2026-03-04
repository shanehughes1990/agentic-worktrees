package main

import (
	coreworker "agentic-orchestrator/internal/core/worker"
	"fmt"
	"os"
)

func main() {
	app, err := coreworker.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap worker: %v\n", err)
		os.Exit(1)
	}
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run worker: %v\n", err)
		os.Exit(1)
	}
}
