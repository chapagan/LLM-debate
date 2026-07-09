package debate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"llmdebate/internal/ai"
	"llmdebate/internal/protocol"
)

const TotalTurns = 4

type EmitFunc func(protocol.OutboundEvent) error

type Runner struct {
	Streamer     ai.Streamer
	NewSessionID func() string
}

type agentTurn struct {
	Agent   protocol.Agent
	Name    string
	Persona string
}

var turns = []agentTurn{
	{Agent: protocol.AgentVisionary, Name: "The Disruptive Visionary", Persona: "Argue for bold technical change, speed, strategic upside, and long-term platform advantage."},
	{Agent: protocol.AgentCFO, Name: "The Skeptical CFO", Persona: "Argue from cost, risk, migration complexity, operational continuity, and measurable return on investment."},
	{Agent: protocol.AgentVisionary, Name: "The Disruptive Visionary", Persona: "Argue for bold technical change, speed, strategic upside, and long-term platform advantage."},
	{Agent: protocol.AgentCFO, Name: "The Skeptical CFO", Persona: "Argue from cost, risk, migration complexity, operational continuity, and measurable return on investment."},
}

func (r Runner) Run(ctx context.Context, topic string, emit EmitFunc) error {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return errors.New("topic is required")
	}
	if r.Streamer == nil {
		return errors.New("streamer is required")
	}
	sessionID := r.sessionID()
	if err := emit(protocol.OutboundEvent{Type: protocol.EventSessionStarted, SessionID: sessionID, Topic: topic, TotalTurns: TotalTurns}); err != nil {
		return err
	}
	conversation := []ai.ChatMessage{{Role: "user", Content: "Debate topic: " + topic}}
	for i, turn := range turns {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		turnNumber := i + 1
		if err := emit(protocol.OutboundEvent{Type: protocol.EventTurnStarted, SessionID: sessionID, Turn: turnNumber, Agent: turn.Agent, AgentName: turn.Name}); err != nil {
			return err
		}
		var full strings.Builder
		result, err := r.Streamer.Stream(ctx, ai.StreamRequest{
			AgentName: turn.Name,
			Persona:   turn.Persona,
			Topic:     topic,
			Messages:  append([]ai.ChatMessage(nil), conversation...),
		}, func(chunk string) error {
			full.WriteString(chunk)
			return emit(protocol.OutboundEvent{Type: protocol.EventToken, SessionID: sessionID, Turn: turnNumber, Agent: turn.Agent, Content: chunk})
		})
		if err != nil {
			return err
		}
		if result == "" {
			result = full.String()
		}
		conversation = append(conversation, ai.ChatMessage{Role: "assistant", Content: fmt.Sprintf("%s: %s", turn.Name, result)})
		if err := emit(protocol.OutboundEvent{Type: protocol.EventTurnEnded, SessionID: sessionID, Turn: turnNumber, Agent: turn.Agent}); err != nil {
			return err
		}
	}
	return emit(protocol.OutboundEvent{Type: protocol.EventDebateEnded, SessionID: sessionID})
}

func (r Runner) sessionID() string {
	if r.NewSessionID != nil {
		return r.NewSessionID()
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "session-fallback"
	}
	return hex.EncodeToString(b[:])
}
