package application

import (
	"context"
	"errors"
	"testing"

	domainservices "github.com/shanehughes1990/agentic-worktrees/internal/domain/services"
)

type generateTaskBoardStubAgentRunner struct {
	result  domainservices.GenerateTaskBoardResult
	err     error
	request domainservices.GenerateTaskBoardRequest
}

func (s *generateTaskBoardStubAgentRunner) GenerateTaskBoard(_ context.Context, request domainservices.GenerateTaskBoardRequest) (domainservices.GenerateTaskBoardResult, error) {
	s.request = request
	return s.result, s.err
}

type generateTaskBoardStubLoader struct {
	documents []domainservices.DocumentationSourceFile
	err       error
}

func (s *generateTaskBoardStubLoader) LoadDocumentationFiles(_ context.Context, _ string, _ int) ([]domainservices.DocumentationSourceFile, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.documents, nil
}

func TestGenerateTaskBoardPrepareCommandExecute(t *testing.T) {
	loader := &generateTaskBoardStubLoader{documents: []domainservices.DocumentationSourceFile{{Path: "docs/a.md", Content: "a"}}}
	command, err := NewPrepareGenerateTaskBoardCommand(loader)
	if err != nil {
		t.Fatalf("new prepare command: %v", err)
	}

	payload, err := command.Execute(context.Background(), GenerateTaskBoardInput{JobID: "job-1", RootDirectory: "docs", MaxDepth: 1, Prompt: "build"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if payload.Metadata.Model != DefaultCreateBoardModel {
		t.Fatalf("expected default model")
	}
	if len(payload.Documents) != 1 {
		t.Fatalf("expected documents")
	}
}

func TestGenerateTaskBoardPrepareCommandErrors(t *testing.T) {
	if _, err := NewPrepareGenerateTaskBoardCommand(nil); err == nil {
		t.Fatalf("expected nil loader error")
	}

	command, err := NewPrepareGenerateTaskBoardCommand(&generateTaskBoardStubLoader{})
	if err != nil {
		t.Fatalf("new prepare command: %v", err)
	}
	if _, err := command.Execute(context.Background(), GenerateTaskBoardInput{}); err == nil {
		t.Fatalf("expected validation error")
	}

	errCommand, err := NewPrepareGenerateTaskBoardCommand(&generateTaskBoardStubLoader{err: errors.New("load fail")})
	if err != nil {
		t.Fatalf("new prepare command: %v", err)
	}
	if _, err := errCommand.Execute(context.Background(), GenerateTaskBoardInput{JobID: "job-1", RootDirectory: "docs", MaxDepth: 1, Prompt: "x"}); err == nil {
		t.Fatalf("expected load error")
	}
}

func TestGenerateTaskBoardExecuteCommandExecute(t *testing.T) {
	agent := &generateTaskBoardStubAgentRunner{result: domainservices.GenerateTaskBoardResult{BoardJSON: "{}"}}
	command, err := NewExecuteGenerateTaskBoardCommand(agent)
	if err != nil {
		t.Fatalf("new execute command: %v", err)
	}

	result, err := command.Execute(context.Background(), GenerateTaskBoardPayload{
		Metadata:      domainservices.AgentRequestMetadata{JobID: "job-1", Model: "gpt-5.3-codex"},
		Prompt:        "build",
		RootDirectory: "docs",
		MaxDepth:      1,
		Documents:     []domainservices.DocumentationSourceFile{{Path: "docs/a.md", Content: "a"}},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result.BoardJSON != "{}" {
		t.Fatalf("unexpected board json")
	}
}
