package main

import (
	"agentic-orchestrator/internal/bootstrap"
	"fmt"
	"os"
)

func main() {
	app, err := bootstrap.InitAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap api: %v\n", err)
		os.Exit(1)
	}
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
