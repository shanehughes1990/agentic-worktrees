package main

import (
	"log"

	"github.com/shanehughes1990/agentic-worktrees/internal/core"
)

func main() {
	runtime, err := core.Init()
	if err != nil {
		log.Fatal(err)
	}

	if err := runtime.Run(); err != nil {
		log.Fatal(err)
	}
}
