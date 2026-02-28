package agent

import (
	"agentic-orchestrator/internal/application/taskengine"
	domainagent "agentic-orchestrator/internal/domain/agent"
	"agentic-orchestrator/internal/domain/failures"
	"context"
	"errors"
)

type Service struct {
	scm domainagent.SCMPort
}

var _ domainagent.ExecutionPort = (*Service)(nil)
var _ domainagent.SessionIntrospectionPort = (*Service)(nil)

func NewService(scm domainagent.SCMPort) (*Service, error) {
	if scm == nil {
		return nil, failures.WrapTerminal(errors.New("agent scm port is required"))
	}
	return &Service{scm: scm}, nil
}

func (service *Service) Execute(ctx context.Context, request domainagent.ExecutionRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}
	var checkpoint *taskengine.Checkpoint
	if request.ResumeCheckpoint != nil {
		checkpoint = &taskengine.Checkpoint{Step: request.ResumeCheckpoint.Step, Token: request.ResumeCheckpoint.Token}
	}
	// checkpoint resume: skip source_state if a prior run already completed it.
	if !taskengine.CheckpointMatches(checkpoint, "source_state", request.Metadata.IdempotencyKey) {
		sourceState, err := service.scm.SourceState(ctx, request.Session.Repository)
		if err != nil {
			return ensureClassified(err)
		}
		if err := sourceState.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) IntrospectSession(ctx context.Context, request domainagent.SessionIntrospectionRequest) (domainagent.SessionState, error) {
	if err := request.Validate(); err != nil {
		return domainagent.SessionState{}, err
	}
	sourceState, err := service.scm.SourceState(ctx, request.Session.Repository)
	if err != nil {
		return domainagent.SessionState{}, ensureClassified(err)
	}
	state := domainagent.SessionState{
		SessionID:   request.Session.SessionID,
		Repository:  request.Session.Repository,
		SourceState: sourceState,
	}
	if err := state.Validate(); err != nil {
		return domainagent.SessionState{}, err
	}
	return state, nil
}

func ensureClassified(err error) error {
	if err == nil {
		return nil
	}
	if failures.ClassOf(err) != failures.ClassUnknown {
		return err
	}
	return failures.WrapTransient(err)
}
