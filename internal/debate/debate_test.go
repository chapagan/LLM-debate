package debate

import (
	"context"
	"testing"
	"time"

	"llmdebate/internal/ai"
	"llmdebate/internal/protocol"
)

type recordingStreamer struct {
	calls []ai.StreamRequest
}

func (r *recordingStreamer) Stream(ctx context.Context, req ai.StreamRequest, emit func(string) error) (string, error) {
	r.calls = append(r.calls, req)
	text := req.AgentName + " response"
	if err := emit(text); err != nil {
		return "", err
	}
	return text, nil
}

func TestRunEmitsFourAlternatingTurns(t *testing.T) {
	streamer := &recordingStreamer{}
	var events []protocol.OutboundEvent
	runner := Runner{Streamer: streamer, NewSessionID: func() string { return "session-1" }}
	err := runner.Run(context.Background(), "Rust rewrite", func(event protocol.OutboundEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	wantAgents := []protocol.Agent{protocol.AgentVisionary, protocol.AgentCFO, protocol.AgentVisionary, protocol.AgentCFO}
	var gotAgents []protocol.Agent
	for _, event := range events {
		if event.Type == protocol.EventTurnStarted {
			gotAgents = append(gotAgents, event.Agent)
		}
	}
	if len(gotAgents) != len(wantAgents) {
		t.Fatalf("turns = %v", gotAgents)
	}
	for i := range wantAgents {
		if gotAgents[i] != wantAgents[i] {
			t.Fatalf("turn %d agent = %s, want %s", i+1, gotAgents[i], wantAgents[i])
		}
	}
	if events[0].Type != protocol.EventSessionStarted {
		t.Fatalf("first event = %s", events[0].Type)
	}
	if events[len(events)-1].Type != protocol.EventDebateEnded {
		t.Fatalf("last event = %s", events[len(events)-1].Type)
	}
	if len(streamer.calls) != 4 {
		t.Fatalf("stream calls = %d", len(streamer.calls))
	}
	if got := len(streamer.calls[3].Messages); got != 4 {
		t.Fatalf("final call messages = %d, want 4", got)
	}
}

func TestRunRejectsEmptyTopic(t *testing.T) {
	runner := Runner{Streamer: &recordingStreamer{}, NewSessionID: func() string { return "session-1" }}
	err := runner.Run(context.Background(), "   ", func(protocol.OutboundEvent) error { return nil })
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunStopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runner := Runner{Streamer: ai.NewMockStreamer(), NewSessionID: func() string { return "session-1" }}
	start := time.Now()
	err := runner.Run(ctx, "Rust rewrite", func(protocol.OutboundEvent) error { return nil })
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatal("cancellation was not prompt")
	}
}
