package resolvers

import (
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	TaskScheduler     *taskengine.Scheduler
	SupervisorService *applicationsupervisor.Service
}

func NewResolver(taskScheduler *taskengine.Scheduler, supervisorService *applicationsupervisor.Service) *Resolver {
	return &Resolver{TaskScheduler: taskScheduler, SupervisorService: supervisorService}
}
