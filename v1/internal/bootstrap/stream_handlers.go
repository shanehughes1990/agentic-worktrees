package bootstrap

import (
	applicationstream "agentic-orchestrator/internal/application/stream"
	domainstream "agentic-orchestrator/internal/domain/stream"
	infraagent "agentic-orchestrator/internal/infrastructure/agent"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type streamInjectRequest struct {
	SessionID     string `json:"session_id"`
	Prompt        string `json:"prompt"`
	RunID         string `json:"run_id"`
	TaskID        string `json:"task_id"`
	JobID         string `json:"job_id"`
	CorrelationID string `json:"correlation_id"`
}

type streamRecoverRequest struct {
	SessionID     string `json:"session_id"`
	RunID         string `json:"run_id"`
	TaskID        string `json:"task_id"`
	JobID         string `json:"job_id"`
	CorrelationID string `json:"correlation_id"`
	Limit         int    `json:"limit"`
}

func streamReplayHandler(streamService *applicationstream.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if streamService == nil {
			http.Error(writer, "stream service is not configured", http.StatusInternalServerError)
			return
		}
		offsetValue := strings.TrimSpace(request.URL.Query().Get("offset"))
		limitValue := strings.TrimSpace(request.URL.Query().Get("limit"))
		offset := uint64(0)
		if offsetValue != "" {
			parsedOffset, err := strconv.ParseUint(offsetValue, 10, 64)
			if err != nil {
				http.Error(writer, "offset must be an unsigned integer", http.StatusBadRequest)
				return
			}
			offset = parsedOffset
		}
		limit := applicationstream.DefaultReplayLimit
		if limitValue != "" {
			parsedLimit, err := strconv.Atoi(limitValue)
			if err != nil {
				http.Error(writer, "limit must be an integer", http.StatusBadRequest)
				return
			}
			limit = parsedLimit
		}
		events, err := streamService.ReplayFromOffset(request.Context(), offset, limit)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{"events": events}); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func streamLiveHandler(streamService *applicationstream.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if streamService == nil {
			http.Error(writer, "stream service is not configured", http.StatusInternalServerError)
			return
		}
		flusher, ok := writer.(http.Flusher)
		if !ok {
			http.Error(writer, "streaming not supported", http.StatusInternalServerError)
			return
		}
		offset := uint64(0)
		offsetValue := strings.TrimSpace(request.URL.Query().Get("offset"))
		if offsetValue != "" {
			parsedOffset, err := strconv.ParseUint(offsetValue, 10, 64)
			if err != nil {
				http.Error(writer, "offset must be an unsigned integer", http.StatusBadRequest)
				return
			}
			offset = parsedOffset
		}
		writer.Header().Set("Content-Type", "text/event-stream")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Header().Set("Connection", "keep-alive")

		replayEvents, err := streamService.ReplayFromOffset(request.Context(), offset, applicationstream.DefaultReplayLimit)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		for _, replayEvent := range replayEvents {
			if err := writeSSEEvent(writer, replayEvent); err != nil {
				return
			}
			offset = replayEvent.StreamOffset
		}
		flusher.Flush()

		_, channel, cancel := streamService.Subscribe(128)
		defer cancel()
		for {
			select {
			case <-request.Context().Done():
				return
			case event, open := <-channel:
				if !open {
					return
				}
				if event.StreamOffset <= offset {
					continue
				}
				if err := writeSSEEvent(writer, event); err != nil {
					return
				}
				offset = event.StreamOffset
				flusher.Flush()
			}
		}
	}
}

func writeSSEEvent(writer http.ResponseWriter, event domainstream.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := writer.Write([]byte("event: stream_event\n")); err != nil {
		return err
	}
	if _, err := writer.Write([]byte("data: " + string(payload) + "\n\n")); err != nil {
		return err
	}
	return nil
}

func streamInjectHandler(streamService *applicationstream.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if streamService == nil {
			http.Error(writer, "stream service is not configured", http.StatusInternalServerError)
			return
		}
		var payload streamInjectRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			http.Error(writer, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		sessionID := strings.TrimSpace(payload.SessionID)
		prompt := strings.TrimSpace(payload.Prompt)
		if sessionID == "" || prompt == "" {
			http.Error(writer, "session_id and prompt are required", http.StatusBadRequest)
			return
		}
		correlationID := strings.TrimSpace(payload.CorrelationID)
		if correlationID == "" {
			correlationID = sessionID
		}
		event, err := streamService.InjectPrompt(request.Context(), sessionID, prompt, domainstream.CorrelationIDs{
			RunID:         strings.TrimSpace(payload.RunID),
			TaskID:        strings.TrimSpace(payload.TaskID),
			JobID:         strings.TrimSpace(payload.JobID),
			SessionID:     sessionID,
			CorrelationID: correlationID,
		})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{"event": event}); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func streamRecoverHandler(streamService *applicationstream.Service, sessionReader *infraagent.SessionStateReader) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if streamService == nil || sessionReader == nil {
			http.Error(writer, "stream recovery is not configured", http.StatusInternalServerError)
			return
		}
		var payload streamRecoverRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			http.Error(writer, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		sessionID := strings.TrimSpace(payload.SessionID)
		if sessionID == "" {
			http.Error(writer, "session_id is required", http.StatusBadRequest)
			return
		}
		correlationID := strings.TrimSpace(payload.CorrelationID)
		if correlationID == "" {
			correlationID = sessionID
		}
		events, err := sessionReader.ReadSessionEvents(request.Context(), sessionID, domainstream.CorrelationIDs{
			RunID:         strings.TrimSpace(payload.RunID),
			TaskID:        strings.TrimSpace(payload.TaskID),
			JobID:         strings.TrimSpace(payload.JobID),
			SessionID:     sessionID,
			CorrelationID: correlationID,
		}, payload.Limit)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		persistedEvents := make([]domainstream.Event, 0, len(events))
		for _, event := range events {
			persistedEvent, err := streamService.AppendAndPublish(request.Context(), event)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
			persistedEvents = append(persistedEvents, persistedEvent)
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{"recovered_events": persistedEvents}); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func streamHealthHandler(streamService *applicationstream.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if streamService == nil {
			http.Error(writer, "stream service is not configured", http.StatusInternalServerError)
			return
		}
		sessionID := strings.TrimSpace(request.URL.Query().Get("session_id"))
		if sessionID == "" {
			http.Error(writer, "session_id is required", http.StatusBadRequest)
			return
		}
		correlationID := strings.TrimSpace(request.URL.Query().Get("correlation_id"))
		if correlationID == "" {
			correlationID = sessionID
		}
		event, err := streamService.PublishHealth(request.Context(), sessionID, domainstream.CorrelationIDs{SessionID: sessionID, CorrelationID: correlationID})
		if err != nil {
			http.Error(writer, fmt.Sprintf("publish stream health: %v", err), http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{"event": event, "published_at": time.Now().UTC()}); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}
