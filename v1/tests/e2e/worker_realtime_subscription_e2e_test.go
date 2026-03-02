//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type websocketMessage struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type graphqlSubscribePayload struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func TestWorkerSessionStreamRealtimeDeliversHeartbeatFiveTimes(t *testing.T) {
	endpoint := strings.TrimSpace(os.Getenv("E2E_GRAPHQL_WS_ENDPOINT"))
	if endpoint == "" {
		endpoint = "ws://localhost:8080/query"
	}

	for attempt := 1; attempt <= 5; attempt++ {
		if err := receiveHeartbeatEvent(endpoint, 20*time.Second); err != nil {
			t.Fatalf("attempt %d failed: %v", attempt, err)
		}
	}
}

func receiveHeartbeatEvent(endpoint string, timeout time.Duration) error {
	header := http.Header{}
	header["Sec-WebSocket-Protocol"] = []string{"graphql-transport-ws", "graphql-ws"}

	connection, _, dialErr := websocket.DefaultDialer.Dial(endpoint, header)
	if dialErr != nil {
		return fmt.Errorf("dial websocket: %w", dialErr)
	}
	defer connection.Close()

	if setReadDeadlineErr := connection.SetReadDeadline(time.Now().Add(timeout)); setReadDeadlineErr != nil {
		return fmt.Errorf("set read deadline: %w", setReadDeadlineErr)
	}
	if setWriteDeadlineErr := connection.SetWriteDeadline(time.Now().Add(timeout)); setWriteDeadlineErr != nil {
		return fmt.Errorf("set write deadline: %w", setWriteDeadlineErr)
	}

	if err := connection.WriteJSON(websocketMessage{Type: "connection_init", Payload: json.RawMessage(`{}`)}); err != nil {
		return fmt.Errorf("write connection_init: %w", err)
	}

	ack := websocketMessage{}
	if err := connection.ReadJSON(&ack); err != nil {
		return fmt.Errorf("read connection_ack: %w", err)
	}
	if ack.Type != "connection_ack" {
		return fmt.Errorf("expected connection_ack, got %q", ack.Type)
	}

	query := "subscription WorkerSessionRealtime($correlation: SupervisorCorrelationInput!, $fromOffset: Int) {" +
		" workerSessionStream(correlation: $correlation, fromOffset: $fromOffset) {" +
		" __typename" +
		" ... on StreamEventSuccess { event { eventType } }" +
		" ... on GraphError { code message field }" +
		" }" +
		" }"
	payloadBytes, marshalErr := json.Marshal(graphqlSubscribePayload{
		Query: query,
		Variables: map[string]any{
			"correlation": map[string]any{"runID": "", "taskID": "", "jobID": ""},
			"fromOffset":  0,
		},
	})
	if marshalErr != nil {
		return fmt.Errorf("marshal subscribe payload: %w", marshalErr)
	}

	protocol := connection.Subprotocol()
	subscribeType := "subscribe"
	if protocol == "graphql-ws" {
		subscribeType = "start"
	}
	if err := connection.WriteJSON(websocketMessage{ID: "worker-stream", Type: subscribeType, Payload: payloadBytes}); err != nil {
		return fmt.Errorf("write subscribe frame: %w", err)
	}

	for {
		frame := websocketMessage{}
		if err := connection.ReadJSON(&frame); err != nil {
			return fmt.Errorf("read subscription frame: %w", err)
		}
		if frame.Type != "next" && frame.Type != "data" {
			continue
		}

		nextPayload := map[string]any{}
		if err := json.Unmarshal(frame.Payload, &nextPayload); err != nil {
			return fmt.Errorf("decode next payload: %w", err)
		}
		dataPayload, dataOK := nextPayload["data"].(map[string]any)
		if !dataOK {
			continue
		}
		streamPayload, streamOK := dataPayload["workerSessionStream"].(map[string]any)
		if !streamOK {
			continue
		}
		typename, _ := streamPayload["__typename"].(string)
		if typename != "StreamEventSuccess" {
			continue
		}
		eventPayload, eventOK := streamPayload["event"].(map[string]any)
		if !eventOK {
			continue
		}
		eventType, _ := eventPayload["eventType"].(string)
		if eventType == "stream.worker.heartbeat" {
			return nil
		}
	}
}
