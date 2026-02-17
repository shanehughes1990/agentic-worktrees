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

func (c *CopilotClient) GenerateTaskBoard(ctx context.Context, request domainservices.GenerateTaskBoardRequest) (domainservices.GenerateTaskBoardResult, error) {
	if err := request.Validate(); err != nil {
		return domainservices.GenerateTaskBoardResult{}, err
	}

	prompt := buildGenerateTaskBoardPrompt(request.Prompt, request.Documents)
	output, err := c.runPrompt(ctx, request.Metadata.Model, prompt)
	if err != nil {
		return domainservices.GenerateTaskBoardResult{}, err
	}
	return domainservices.GenerateTaskBoardResult{BoardJSON: output}, nil
}

func buildGenerateTaskBoardPrompt(basePrompt string, documents []domainservices.DocumentationSourceFile) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(basePrompt))
	builder.WriteString("\n\nAuthoritative document payload follows:\n")

	for _, document := range documents {
		builder.WriteString("\n--- DOCUMENT START: ")
		builder.WriteString(strings.TrimSpace(document.Path))
		builder.WriteString(" ---\n")
		builder.WriteString(document.Content)
		if !strings.HasSuffix(document.Content, "\n") {
			builder.WriteString("\n")
		}
		builder.WriteString("--- DOCUMENT END ---\n")
	}

	builder.WriteString("\nUse only this provided payload to generate the task board output.")
	return builder.String()
}

func (c *CopilotClient) runPrompt(ctx context.Context, model string, prompt string) (string, error) {
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

	session, err := c.client.CreateSession(ctx, &copilot.SessionConfig{Model: resolvedModel})
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
