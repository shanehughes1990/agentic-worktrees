package agent

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ACPClient struct {
	binaryPath   string
	workingDir   string
	process      *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	mu           sync.Mutex
	connected    atomic.Bool
	lastEventAt  atomic.Int64
}

func NewACPClient(binaryPath string, workingDir string) (*ACPClient, error) {
	path := strings.TrimSpace(binaryPath)
	if path == "" {
		path = "copilot"
	}
	return &ACPClient{binaryPath: path, workingDir: strings.TrimSpace(workingDir)}, nil
}

func (client *ACPClient) Start(ctx context.Context, publish func(ctx context.Context, event domainstream.Event) error) error {
	if publish == nil {
		return fmt.Errorf("acp client: publish callback is required")
	}
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.process != nil {
		return fmt.Errorf("acp client: process already started")
	}
	command := exec.CommandContext(ctx, client.binaryPath, "--acp", "--stdio")
	if client.workingDir != "" {
		command.Dir = client.workingDir
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		return fmt.Errorf("acp client: stdin pipe: %w", err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("acp client: stdout pipe: %w", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return fmt.Errorf("acp client: stderr pipe: %w", err)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("acp client: start: %w", err)
	}
	client.process = command
	client.stdin = stdin
	client.stdout = stdout
	client.stderr = stderr
	client.connected.Store(true)
	client.lastEventAt.Store(time.Now().UTC().Unix())
	go client.consumeStdout(ctx, publish)
	go client.consumeStderr()
	go client.waitForExit()
	return nil
}

func (client *ACPClient) InjectPrompt(ctx context.Context, sessionID string, prompt string) error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.stdin == nil {
		return fmt.Errorf("acp client: stdin is not initialized")
	}
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      fmt.Sprintf("inject-%d", time.Now().UTC().UnixNano()),
		"method":  "prompt",
		"params": map[string]any{
			"sessionId": strings.TrimSpace(sessionID),
			"prompt": []map[string]any{{
				"type": "text",
				"text": strings.TrimSpace(prompt),
			}},
		},
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("acp client: marshal prompt request: %w", err)
	}
	if _, err := client.stdin.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("acp client: write prompt request: %w", err)
	}
	_ = ctx
	return nil
}

func (client *ACPClient) Evaluate(ctx context.Context, sessionID string) (map[string]any, error) {
	lastEventUnix := client.lastEventAt.Load()
	lastEventAt := time.Unix(lastEventUnix, 0).UTC()
	status := map[string]any{
		"session_id":        strings.TrimSpace(sessionID),
		"connected":         client.connected.Load(),
		"last_event_at":     lastEventAt,
		"seconds_since_event": int(time.Since(lastEventAt).Seconds()),
	}
	_ = ctx
	return status, nil
}

func (client *ACPClient) Shutdown() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.process == nil {
		return nil
	}
	client.connected.Store(false)
	if client.stdin != nil {
		_ = client.stdin.Close()
	}
	if client.process.Process != nil {
		if err := client.process.Process.Kill(); err != nil {
			return fmt.Errorf("acp client: kill process: %w", err)
		}
	}
	client.process = nil
	client.stdin = nil
	client.stdout = nil
	client.stderr = nil
	return nil
}

func (client *ACPClient) consumeStdout(ctx context.Context, publish func(ctx context.Context, event domainstream.Event) error) {
	client.mu.Lock()
	stdout := client.stdout
	client.mu.Unlock()
	if stdout == nil {
		return
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rawMessage := map[string]any{}
		if err := json.Unmarshal([]byte(line), &rawMessage); err != nil {
			continue
		}
		event, ok := normalizeACPMessage(rawMessage)
		if !ok {
			continue
		}
		client.lastEventAt.Store(time.Now().UTC().Unix())
		if err := publish(ctx, event); err != nil {
			continue
		}
	}
}

func (client *ACPClient) consumeStderr() {
	client.mu.Lock()
	stderr := client.stderr
	client.mu.Unlock()
	if stderr == nil {
		return
	}
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		_ = scanner.Text()
	}
}

func (client *ACPClient) waitForExit() {
	client.mu.Lock()
	process := client.process
	client.mu.Unlock()
	if process == nil {
		return
	}
	_ = process.Wait()
	client.connected.Store(false)
}

func normalizeACPMessage(raw map[string]any) (domainstream.Event, bool) {
	now := time.Now().UTC()
	event := domainstream.Event{
		EventID:    fmt.Sprintf("acp-%d", now.UnixNano()),
		OccurredAt: now,
		Source:     domainstream.SourceACP,
		EventType:  domainstream.EventSessionUpdated,
		CorrelationIDs: domainstream.CorrelationIDs{
			CorrelationID: fmt.Sprintf("acp-correlation-%d", now.UnixNano()),
		},
		Payload: map[string]any{"raw": raw},
	}
	if methodValue, ok := raw["method"].(string); ok && methodValue == "sessionUpdate" {
		if paramsMap, ok := raw["params"].(map[string]any); ok {
			if updateMap, ok := paramsMap["update"].(map[string]any); ok {
				if updateType, ok := updateMap["sessionUpdate"].(string); ok {
					switch strings.TrimSpace(updateType) {
					case "agent_message_chunk":
						event.EventType = domainstream.EventAgentChunk
					case "agent_turn_completed":
						event.EventType = domainstream.EventAgentTurnCompleted
					}
				}
			}
			if sessionID, ok := paramsMap["sessionId"].(string); ok {
				event.CorrelationIDs.SessionID = strings.TrimSpace(sessionID)
				event.CorrelationIDs.CorrelationID = strings.TrimSpace(sessionID)
			}
		}
	}
	if err := event.Validate(); err != nil {
		return domainstream.Event{}, false
	}
	return event, true
}
