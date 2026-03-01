package agent

import (
	domainstream "agentic-orchestrator/internal/domain/stream"
	"testing"
)

func TestNormalizeACPMessageChunk(t *testing.T) {
	event, ok := normalizeACPMessage(map[string]any{
		"method": "sessionUpdate",
		"params": map[string]any{
			"sessionId": "session-123",
			"update": map[string]any{
				"sessionUpdate": "agent_message_chunk",
			},
		},
	})
	if !ok {
		t.Fatalf("expected normalized event")
	}
	if event.EventType != domainstream.EventAgentChunk {
		t.Fatalf("unexpected event type: %s", event.EventType)
	}
	if event.CorrelationIDs.SessionID != "session-123" {
		t.Fatalf("unexpected session id: %s", event.CorrelationIDs.SessionID)
	}
}
