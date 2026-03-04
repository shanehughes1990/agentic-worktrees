package resolvers

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationstream "agentic-orchestrator/internal/application/stream"
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	applicationworker "agentic-orchestrator/internal/application/worker"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	TaskScheduler        *taskengine.Scheduler
	SupervisorService    *applicationsupervisor.Service
	ControlPlaneService  *applicationcontrolplane.Service
	StreamService        *applicationstream.Service
	TrackerService       *applicationtracker.Service
	WorkerService        *applicationworker.Service
}

func NewResolver(taskScheduler *taskengine.Scheduler, supervisorService *applicationsupervisor.Service, controlPlaneService *applicationcontrolplane.Service, streamService *applicationstream.Service, workerService *applicationworker.Service, trackerService ...*applicationtracker.Service) *Resolver {
	var tracker *applicationtracker.Service
	if len(trackerService) > 0 {
		tracker = trackerService[0]
	}
	return &Resolver{TaskScheduler: taskScheduler, SupervisorService: supervisorService, ControlPlaneService: controlPlaneService, StreamService: streamService, TrackerService: tracker, WorkerService: workerService}
}
