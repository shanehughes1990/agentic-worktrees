package main

import (
  coreworker "agentic-orchestrator/internal/core/worker"
  "fmt"
  "os"
)

func main() {
  app, err := coreworker.New()
  if err != nil {
    fmt.Printf("NEW_ERROR: %v\n", err)
    os.Exit(2)
  }
  if err := app.Run(); err != nil {
    fmt.Printf("RUN_ERROR: %v\n", err)
    os.Exit(3)
  }
}
