package copilot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	sdk "github.com/github/copilot-sdk/go"
	"github.com/sirupsen/logrus"
)

type Client struct {
	config ClientConfig
	logger *logrus.Logger
}

func NewClient(config ClientConfig, logger *logrus.Logger) *Client {
	return &Client{config: config.Normalized(), logger: logger}
}

func (client *Client) RunPrompt(ctx context.Context, requestedModel string, resumeSessionID string, workingDirectory string, skillDirectories []string, prompt string) (string, string, string, error) {
	model := strings.TrimSpace(requestedModel)
	if model == "" {
		model = client.config.DefaultModel
	}
	cleanResumeSessionID := strings.TrimSpace(resumeSessionID)

	entry := client.entry().WithFields(logrus.Fields{
		"event":             "copilot.run_prompt",
		"model":             model,
		"working_directory": strings.TrimSpace(workingDirectory),
	})
	entry.WithFields(logrus.Fields{
		"cli_path":         strings.TrimSpace(client.config.CLIPath),
		"cli_url":          strings.TrimSpace(client.config.CLIURL),
		"has_github_token": strings.TrimSpace(client.config.GitHubToken) != "",
		"home":             strings.TrimSpace(os.Getenv("HOME")),
		"pwd":              strings.TrimSpace(os.Getenv("PWD")),
	}).Info("copilot startup context")
	entry.Info("starting copilot client")

	preflight, preflightErr := client.runStartupPreflight(ctx)
	if preflightErr != nil {
		if isSDKCLIIncompatibilityError(preflightErr) {
			entry.WithError(preflightErr).WithField("preflight", preflight).Warn("copilot sdk/cli incompatibility detected; using direct cli fallback")
			return client.runPromptViaCLIFallback(ctx, model, workingDirectory, prompt)
		}
		entry.WithError(preflightErr).WithField("preflight", preflight).Error("copilot startup preflight failed")
		return "", "", "", fmt.Errorf("copilot preflight failed: %w", preflightErr)
	}
	entry.WithField("preflight", preflight).Info("copilot startup preflight succeeded")

	options := &sdk.ClientOptions{
		GitHubToken: client.config.GitHubToken,
		CLIPath:     client.config.CLIPath,
		CLIUrl:      client.config.CLIURL,
		LogLevel:    client.config.LogLevel,
		Cwd:         strings.TrimSpace(workingDirectory),
	}

	sdkClient := sdk.NewClient(options)
	if err := sdkClient.Start(ctx); err != nil {
		if isSDKCLIIncompatibilityError(err) {
			entry.WithError(err).Warn("copilot sdk start incompatible with installed cli; using direct cli fallback")
			return client.runPromptViaCLIFallback(ctx, model, workingDirectory, prompt)
		}
		friendly := explainClientStartError(err)
		entry.WithError(err).Error("failed to start copilot client")
		return "", "", "", fmt.Errorf("start copilot client: %s: %w", friendly, err)
	}
	entry.Info("copilot client started")
	defer sdkClient.Stop()

	authStatus, authErr := sdkClient.GetAuthStatus(ctx)
	if authErr != nil {
		entry.WithError(authErr).Error("failed to get copilot sdk auth status")
		return "", "", "", fmt.Errorf("check copilot auth status from sdk: %w", authErr)
	}
	if authStatus == nil {
		entry.Error("copilot sdk returned nil auth status")
		return "", "", "", fmt.Errorf("copilot sdk auth status unavailable")
	}
	entry.WithFields(logrus.Fields{
		"is_authenticated": authStatus.IsAuthenticated,
		"auth_type":        safeStringPtr(authStatus.AuthType),
		"login":            safeStringPtr(authStatus.Login),
		"host":             safeStringPtr(authStatus.Host),
		"status_message":   safeStringPtr(authStatus.StatusMessage),
	}).Info("copilot sdk auth status")
	if !authStatus.IsAuthenticated {
		return "", "", "", fmt.Errorf("copilot not authenticated: %s", formatSDKAuthStatus(authStatus))
	}

	combinedSkills := skillDirectories
	if len(combinedSkills) == 0 {
		combinedSkills = client.config.SkillDirectories
	}

	var (
		session *sdk.Session
		err     error
	)
	if cleanResumeSessionID != "" {
		session, err = sdkClient.ResumeSession(ctx, cleanResumeSessionID)
		if err != nil {
			entry.WithError(err).WithField("resume_session_id", cleanResumeSessionID).Warn("failed to resume copilot session; creating a fresh session")
		}
	}
	if session == nil {
		session, err = sdkClient.CreateSession(ctx, &sdk.SessionConfig{
			Model:            model,
			WorkingDirectory: workingDirectory,
			SkillDirectories: combinedSkills,
			OnPermissionRequest: func(sdk.PermissionRequest, sdk.PermissionInvocation) (sdk.PermissionRequestResult, error) {
				return sdk.PermissionRequestResult{Kind: "approved"}, nil
			},
		})
		if err != nil {
			entry.WithError(err).Error("failed to create copilot session")
			return "", "", "", fmt.Errorf("create copilot session: %w", err)
		}
	} else {
		entry.WithField("resumed_session_id", cleanResumeSessionID).Info("copilot session resumed")
	}
	entry = entry.WithField("session_id", session.SessionID)
	entry.Info("copilot session created")
	defer session.Destroy()

	responseEvent, err := session.SendAndWait(ctx, sdk.MessageOptions{Prompt: prompt})
	if err != nil {
		entry.WithError(err).Error("failed to send prompt to copilot session")
		return session.SessionID, "", model, fmt.Errorf("send decomposition prompt: %w", err)
	}

	response := ""
	if responseEvent != nil && responseEvent.Data.Content != nil {
		response = *responseEvent.Data.Content
	}
	entry.WithField("response_bytes", len(response)).Info("received copilot response")

	return session.SessionID, response, model, nil
}

func (client *Client) runPromptViaCLIFallback(ctx context.Context, model string, workingDirectory string, prompt string) (string, string, string, error) {
	binary := strings.TrimSpace(client.config.CLIPath)
	if binary == "" {
		binary = "copilot"
	}
	args := []string{"-p", prompt, "--allow-all-tools", "--allow-all-paths", "--silent"}
	if strings.TrimSpace(model) != "" {
		args = append(args, "--model", strings.TrimSpace(model))
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	if strings.TrimSpace(workingDirectory) != "" {
		cmd.Dir = strings.TrimSpace(workingDirectory)
	}

	env := os.Environ()
	token := strings.TrimSpace(client.config.GitHubToken)
	if token != "" {
		env = append(env,
			"COPILOT_SDK_AUTH_TOKEN="+token,
			"GITHUB_TOKEN="+token,
			"GH_TOKEN="+token,
		)
	}
	cmd.Env = env

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if stderrText == "" {
			stderrText = strings.TrimSpace(stdout.String())
		}
		return "", "", "", fmt.Errorf("copilot cli fallback prompt failed: %w: %s", err, stderrText)
	}

	response := strings.TrimSpace(stdout.String())
	if response == "" {
		response = strings.TrimSpace(stderr.String())
	}
	if response == "" {
		return "", "", "", fmt.Errorf("copilot cli fallback returned empty response")
	}

	sessionID := fmt.Sprintf("cli-fallback-%d", time.Now().UnixNano())
	client.entry().WithFields(logrus.Fields{
		"event":          "copilot.run_prompt.cli_fallback",
		"session_id":     sessionID,
		"response_bytes": len(response),
		"binary":         binary,
	}).Info("copilot cli fallback completed")
	return sessionID, response, model, nil
}

func (client *Client) runStartupPreflight(ctx context.Context) (string, error) {
	checkCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	versionCommand := "copilot --version"
	if strings.TrimSpace(client.config.CLIPath) != "" {
		versionCommand = strings.TrimSpace(client.config.CLIPath) + " --version"
	}
	stdout, stderr, err := runCommand(checkCtx, versionCommand, false)
	if err != nil {
		return fmt.Sprintf("version_check_failed command=%q stderr=%q", versionCommand, strings.TrimSpace(stderr)), err
	}
	versionOutput := strings.TrimSpace(joinOutput(stdout, stderr))
	if versionOutput == "" {
		versionOutput = "version check completed with empty output"
	}

	compatCommand, compatErr := runSDKCLICompatibilityCheck(checkCtx, client.config.CLIPath)
	if compatErr != nil {
		return fmt.Sprintf("command=%q output=%q startup_probe_command=%q", versionCommand, versionOutput, compatCommand), compatErr
	}

	return fmt.Sprintf("command=%q output=%q startup_probe_command=%q", versionCommand, versionOutput, compatCommand), nil
}

func runSDKCLICompatibilityCheck(ctx context.Context, cliPath string) (string, error) {
	baseCommand := strings.TrimSpace(cliPath)
	if baseCommand == "" {
		baseCommand = "copilot"
	}
	command := fmt.Sprintf("%s --headless --no-auto-update --log-level error --stdio", baseCommand)
	probeCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	_, stderr, err := runCommand(probeCtx, command, false)
	if err == nil {
		return command, fmt.Errorf("copilot cli startup probe exited unexpectedly without opening stdio transport")
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return command, nil
	}

	stderrText := strings.TrimSpace(stderr)
	lower := strings.ToLower(stderrText)
	if strings.Contains(lower, "unknown option '--headless'") || strings.Contains(lower, "unknown option '--stdio'") || strings.Contains(lower, "unknown option '--no-auto-update'") {
		return command, fmt.Errorf("installed Copilot CLI is incompatible with Go SDK v0.1.25 process launch flags (--headless/--stdio/--no-auto-update). install an SDK-compatible Copilot CLI build or upgrade the SDK. cli_stderr=%q", stderrText)
	}
	return command, fmt.Errorf("copilot cli startup probe failed: %w: %s", err, stderrText)
}

func isSDKCLIIncompatibilityError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "incompatible with go sdk") || strings.Contains(message, "process launch flags") || strings.Contains(message, "unknown option '--headless'") || strings.Contains(message, "unknown option '--stdio'")
}

func (client *Client) entry() *logrus.Entry {
	if client.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(client.logger)
}

func explainClientStartError(err error) string {
	if err == nil {
		return "copilot client startup failed"
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "incompatible with go sdk") || strings.Contains(message, "--headless") || strings.Contains(message, "--stdio") {
		return "installed Copilot CLI does not match SDK launch contract; use an SDK-compatible CLI build or upgrade SDK"
	}
	if strings.Contains(message, "exit status 1") || strings.Contains(message, "cli process exited") {
		return "Copilot CLI process failed during startup; check GitHub auth, Copilot entitlement, CLI path/url, and token configuration"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "Copilot startup timed out or was canceled"
	}
	if strings.Contains(message, "not found") || strings.Contains(message, "executable file") {
		return "Copilot CLI binary was not found; verify the configured CLI path"
	}
	return "Copilot client failed to start"
}

func safeStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func formatSDKAuthStatus(status *sdk.GetAuthStatusResponse) string {
	if status == nil {
		return "auth status unavailable"
	}
	parts := []string{}
	if authType := safeStringPtr(status.AuthType); authType != "" {
		parts = append(parts, "auth_type="+authType)
	}
	if login := safeStringPtr(status.Login); login != "" {
		parts = append(parts, "login="+login)
	}
	if host := safeStringPtr(status.Host); host != "" {
		parts = append(parts, "host="+host)
	}
	if message := safeStringPtr(status.StatusMessage); message != "" {
		parts = append(parts, "message="+message)
	}
	if len(parts) == 0 {
		return "not authenticated"
	}
	return strings.Join(parts, ", ")
}

func runCommand(ctx context.Context, command string, _ bool) (string, string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", "", fmt.Errorf("command is required")
	}
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func joinOutput(stdout string, stderr string) string {
	out := strings.TrimSpace(stdout)
	errOut := strings.TrimSpace(stderr)
	if out == "" {
		return errOut
	}
	if errOut == "" {
		return out
	}
	return out + "\n" + errOut
}
