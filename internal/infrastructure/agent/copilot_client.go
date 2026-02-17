package agent

import (
	"context"
	"fmt"
	"strings"

	copilot "github.com/github/copilot-sdk/go"
	domainservices "github.com/shanehughes1990/agentic-worktrees/internal/domain/services"
	"github.com/sirupsen/logrus"
)

const DefaultCopilotModel = "gpt-5.3-codex"

type CopilotClient struct {
	client *copilot.Client
}

var _ domainservices.AgentRunner = (*CopilotClient)(nil)

func NewCopilotClient(appLogger *logrus.Logger, cliPath string, cliURL string, githubToken string) (*CopilotClient, error) {
	if appLogger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	options := &copilot.ClientOptions{}
	if strings.TrimSpace(cliPath) != "" {
		options.CLIPath = strings.TrimSpace(cliPath)
	}
	if strings.TrimSpace(cliURL) != "" {
		options.CLIUrl = strings.TrimSpace(cliURL)
	}
	if strings.TrimSpace(githubToken) != "" {
		options.GithubToken = strings.TrimSpace(githubToken)
		options.UseLoggedInUser = copilot.Bool(false)
	}
	options.LogLevel = appLogger.GetLevel().String()

	return &CopilotClient{client: copilot.NewClient(options)}, nil
}

func (c *CopilotClient) Client() *copilot.Client {
	if c == nil {
		return nil
	}
	return c.client
}

func (c *CopilotClient) Start(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("copilot client is not initialized")
	}
	return c.client.Start(ctx)
}

func (c *CopilotClient) Stop() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Stop()
}

func (c *CopilotClient) DoTaskFromTaskBoard(ctx context.Context, request domainservices.DoTaskFromTaskBoardRequest) (domainservices.DoTaskFromTaskBoardResult, error) {
	if strings.TrimSpace(request.TaskID) == "" {
		return domainservices.DoTaskFromTaskBoardResult{}, fmt.Errorf("task id cannot be empty")
	}
	if strings.TrimSpace(request.Prompt) == "" {
		return domainservices.DoTaskFromTaskBoardResult{}, fmt.Errorf("prompt cannot be empty")
	}

	output, err := c.runPrompt(ctx, request.Metadata.Model, request.Metadata.RepositoryPath, request.Prompt)
	if err != nil {
		return domainservices.DoTaskFromTaskBoardResult{}, err
	}
	return domainservices.DoTaskFromTaskBoardResult{Summary: output}, nil
}

func (c *CopilotClient) CreateTaskBoardFromTextFiles(ctx context.Context, request domainservices.CreateTaskBoardFromTextFilesRequest) (domainservices.CreateTaskBoardFromTextFilesResult, error) {
	if len(request.FilePaths) == 0 {
		return domainservices.CreateTaskBoardFromTextFilesResult{}, fmt.Errorf("file paths cannot be empty")
	}
	if strings.TrimSpace(request.Prompt) == "" {
		return domainservices.CreateTaskBoardFromTextFilesResult{}, fmt.Errorf("prompt cannot be empty")
	}

	output, err := c.runPrompt(ctx, request.Metadata.Model, request.Metadata.RepositoryPath, request.Prompt)
	if err != nil {
		return domainservices.CreateTaskBoardFromTextFilesResult{}, err
	}
	return domainservices.CreateTaskBoardFromTextFilesResult{BoardJSON: output}, nil
}

func (c *CopilotClient) ResolveGitConflicts(ctx context.Context, request domainservices.ResolveGitConflictsRequest) (domainservices.ResolveGitConflictsResult, error) {
	if len(request.ConflictFiles) == 0 {
		return domainservices.ResolveGitConflictsResult{}, fmt.Errorf("conflict files cannot be empty")
	}
	if strings.TrimSpace(request.Prompt) == "" {
		return domainservices.ResolveGitConflictsResult{}, fmt.Errorf("prompt cannot be empty")
	}

	output, err := c.runPrompt(ctx, request.Metadata.Model, request.Metadata.RepositoryPath, request.Prompt)
	if err != nil {
		return domainservices.ResolveGitConflictsResult{}, err
	}
	return domainservices.ResolveGitConflictsResult{Summary: output, ResolvedFiles: request.ConflictFiles}, nil
}

func (c *CopilotClient) RunPrompt(ctx context.Context, model string, prompt string) (string, error) {
	return c.runPrompt(ctx, model, "", prompt)
}

func (c *CopilotClient) runPrompt(ctx context.Context, model string, repositoryPath string, prompt string) (string, error) {
	if c == nil || c.client == nil {
		return "", fmt.Errorf("copilot client is not initialized")
	}
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	resolvedModel := strings.TrimSpace(model)
	if resolvedModel == "" {
		resolvedModel = DefaultCopilotModel
	}

	sessionConfig := &copilot.SessionConfig{
		Model: resolvedModel,
	}
	if strings.TrimSpace(repositoryPath) != "" {
		sessionConfig.WorkingDirectory = strings.TrimSpace(repositoryPath)
	}

	session, err := c.client.CreateSession(ctx, sessionConfig)
	if err != nil {
		return "", fmt.Errorf("create copilot session: %w", err)
	}
	defer session.Destroy()

	response, err := session.SendAndWait(ctx, copilot.MessageOptions{Prompt: prompt})
	if err != nil {
		return "", fmt.Errorf("send copilot prompt: %w", err)
	}
	if response == nil || response.Data.Content == nil {
		return "", fmt.Errorf("copilot response content is empty")
	}

	return *response.Data.Content, nil
}
