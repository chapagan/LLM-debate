package ai

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ChatMessage struct {
	Role    string
	Content string
}

type StreamRequest struct {
	AgentName string
	Persona   string
	Topic     string
	Messages  []ChatMessage
}

type Streamer interface {
	Stream(ctx context.Context, req StreamRequest, emit func(string) error) (string, error)
}

type MockStreamer struct {
	Delay time.Duration
}

func NewMockStreamer() *MockStreamer {
	return &MockStreamer{Delay: 20 * time.Millisecond}
}

func (m *MockStreamer) Stream(ctx context.Context, req StreamRequest, emit func(string) error) (string, error) {
	text := fmt.Sprintf("On %q, my position is shaped by %d prior messages. %s", req.Topic, len(req.Messages), mockSentence(req.AgentName))
	parts := strings.SplitAfter(text, " ")
	var b strings.Builder
	for _, part := range parts {
		select {
		case <-ctx.Done():
			return b.String(), ctx.Err()
		default:
		}
		if err := emit(part); err != nil {
			return b.String(), err
		}
		b.WriteString(part)
		if m.Delay > 0 {
			timer := time.NewTimer(m.Delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return b.String(), ctx.Err()
			case <-timer.C:
			}
		}
	}
	return b.String(), nil
}

func mockSentence(agentName string) string {
	if strings.Contains(strings.ToLower(agentName), "cfo") {
		return "The financial case must survive migration cost, staffing risk, and measurable ROI."
	}
	return "The strategic upside is speed, leverage, and escaping legacy constraints before they compound."
}
