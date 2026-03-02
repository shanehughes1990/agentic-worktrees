package main

import (
	coreapi "agentic-orchestrator/internal/core/api"
	"fmt"
	"os"
)

func main() {
	app, err := coreapi.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap api: %v\n", err)
		os.Exit(1)
	}
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
