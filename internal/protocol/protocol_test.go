package protocol

import (
	"encoding/json"
	"testing"
)

func TestInboundEventDecodesStartDebate(t *testing.T) {
	var event InboundEvent
	raw := []byte(`{"action":"START_DEBATE","topic":"Should we rewrite?"}`)
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("decode inbound event: %v", err)
	}
	if event.Action != ActionStartDebate {
		t.Fatalf("action = %q, want %q", event.Action, ActionStartDebate)
	}
	if event.Topic != "Should we rewrite?" {
		t.Fatalf("topic = %q", event.Topic)
	}
}

func TestOutboundEventOmitsEmptyFields(t *testing.T) {
	event := OutboundEvent{
		Type:      EventToken,
		SessionID: "session-1",
		Turn:      2,
		Agent:     AgentCFO,
		Content:   "No budget.",
	}
	got, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal outbound event: %v", err)
	}
	want := `{"type":"TOKEN","sessionId":"session-1","turn":2,"agent":"cfo","content":"No budget."}`
	if string(got) != want {
		t.Fatalf("json = %s, want %s", got, want)
	}
}
