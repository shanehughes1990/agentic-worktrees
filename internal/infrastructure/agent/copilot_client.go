package agent

import (
	"context"
	"fmt"
	"strings"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/sirupsen/logrus"
)

const DefaultCopilotModel = "gpt-5.3-codex"

type CopilotClient struct {
	client *copilot.Client
}

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

func (c *CopilotClient) RunPrompt(ctx context.Context, model string, prompt string) (string, error) {
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
