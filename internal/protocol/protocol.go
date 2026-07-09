package protocol

type Action string

const (
	ActionStartDebate Action = "START_DEBATE"
)

type EventType string

const (
	EventSessionStarted EventType = "SESSION_STARTED"
	EventTurnStarted    EventType = "TURN_STARTED"
	EventToken          EventType = "TOKEN"
	EventTurnEnded      EventType = "TURN_ENDED"
	EventDebateEnded    EventType = "DEBATE_ENDED"
	EventError          EventType = "ERROR"
)

type Agent string

const (
	AgentVisionary Agent = "visionary"
	AgentCFO       Agent = "cfo"
)

type InboundEvent struct {
	Action Action `json:"action"`
	Topic  string `json:"topic"`
}

type OutboundEvent struct {
	Type       EventType `json:"type"`
	SessionID  string    `json:"sessionId,omitempty"`
	Topic      string    `json:"topic,omitempty"`
	TotalTurns int       `json:"totalTurns,omitempty"`
	Turn       int       `json:"turn,omitempty"`
	Agent      Agent     `json:"agent,omitempty"`
	AgentName  string    `json:"agentName,omitempty"`
	Content    string    `json:"content,omitempty"`
	Message    string    `json:"message,omitempty"`
}
