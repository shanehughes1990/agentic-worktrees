package stream

const (
	EventSessionStarted       EventType = "stream.session.started"
	EventSessionUpdated       EventType = "stream.session.updated"
	EventAgentChunk           EventType = "stream.agent.chunk"
	EventAgentTurnCompleted   EventType = "stream.agent.turn_completed"
	EventToolStarted          EventType = "stream.tool.started"
	EventToolCompleted        EventType = "stream.tool.completed"
	EventPermissionRequested  EventType = "stream.permission.requested"
	EventPermissionDecided    EventType = "stream.permission.decided"
	EventSessionCheckpointed  EventType = "stream.session.checkpointed"
	EventSessionEnded         EventType = "stream.session.ended"
	EventSessionRecovered     EventType = "stream.session.recovered"
	EventSessionHealth        EventType = "stream.session.health"
	EventSessionInjectedPrompt EventType = "stream.session.injected_prompt"
	EventWorkerHeartbeat      EventType = "stream.worker.heartbeat"
	EventWorkerShutdown       EventType = "stream.worker.shutdown"
	EventWorkerDeregistered   EventType = "stream.worker.deregistered"
	EventWorkerRogueDetected  EventType = "stream.worker.rogue_detected"
)
