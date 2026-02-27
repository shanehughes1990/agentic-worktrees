package resolvers

import "agentic-orchestrator/internal/application/taskengine"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	TaskScheduler *taskengine.Scheduler
}

func NewResolver(taskScheduler *taskengine.Scheduler) *Resolver {
	return &Resolver{TaskScheduler: taskScheduler}
}
