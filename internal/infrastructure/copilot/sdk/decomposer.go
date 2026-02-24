package sdk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	copilot "github.com/github/copilot-sdk/go"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
)

type Decomposer struct{}

func NewDecomposer() *Decomposer {
	return &Decomposer{}
}

func (decomposer *Decomposer) Decompose(ctx context.Context, request appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	prompt := strings.TrimSpace(request.Prompt)
	if prompt == "" {
		return appcopilot.DecomposeResult{}, fmt.Errorf("prompt is required")
	}

	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = "gpt-5"
	}

	clientOptions := &copilot.ClientOptions{
		GithubToken: request.GitHubToken,
		CLIPath:     request.CLIPath,
		CLIUrl:      request.CLIURL,
		LogLevel:    "error",
	}

	client := copilot.NewClient(clientOptions)
	if err := client.Start(ctx); err != nil {
		return appcopilot.DecomposeResult{}, fmt.Errorf("start copilot client: %w", err)
	}
	defer client.Stop()

	sessionConfig := &copilot.SessionConfig{
		Model:            model,
		WorkingDirectory: request.WorkingDirectory,
		SkillDirectories: request.SkillDirectories,
		OnPermissionRequest: func(copilot.PermissionRequest, copilot.PermissionInvocation) (copilot.PermissionRequestResult, error) {
			return copilot.PermissionRequestResult{Kind: "approved"}, nil
		},
	}

	session, err := client.CreateSession(ctx, sessionConfig)
	if err != nil {
		return appcopilot.DecomposeResult{}, fmt.Errorf("create copilot session: %w", err)
	}
	defer session.Destroy()

	responseEvent, err := session.SendAndWait(ctx, copilot.MessageOptions{Prompt: prompt})
	if err != nil {
		return appcopilot.DecomposeResult{}, fmt.Errorf("send decomposition prompt: %w", err)
	}

	response := ""
	if responseEvent != nil && responseEvent.Data.Content != nil {
		response = *responseEvent.Data.Content
	}

	hash := sha256.Sum256([]byte(prompt))
	return appcopilot.DecomposeResult{
		RunID:      request.RunID,
		SessionID:  session.SessionID,
		Response:   response,
		Model:      model,
		PromptHash: hex.EncodeToString(hash[:]),
	}, nil
}
