package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnthropicStreamerReadsTextDeltas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "anthropic-key" {
			t.Fatalf("x-api-key = %q", got)
		}
		if got := r.Header.Get("anthropic-version"); got == "" {
			t.Fatal("missing anthropic-version header")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: content_block_delta\n"))
		_, _ = w.Write([]byte(`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}` + "\n\n"))
		_, _ = w.Write([]byte("event: content_block_delta\n"))
		_, _ = w.Write([]byte(`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":" world"}}` + "\n\n"))
		_, _ = w.Write([]byte("event: message_stop\n"))
		_, _ = w.Write([]byte(`data: {"type":"message_stop"}` + "\n\n"))
	}))
	defer server.Close()

	streamer, err := NewAnthropicStreamer(AnthropicConfig{
		APIKey:  "anthropic-key",
		Model:   "claude-sonnet-4-5",
		BaseURL: server.URL + "/v1",
		Client:  server.Client(),
	})
	if err != nil {
		t.Fatalf("new streamer: %v", err)
	}
	var chunks []string
	result, err := streamer.Stream(context.Background(), StreamRequest{
		AgentName: "The Disruptive Visionary",
		Persona:   "Argue for bold change.",
		Topic:     "Rust rewrite",
		Messages:  []ChatMessage{{Role: "user", Content: "Debate context"}},
	}, func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if result != "Hello world" {
		t.Fatalf("result = %q", result)
	}
	if strings.Join(chunks, "") != "Hello world" {
		t.Fatalf("chunks = %q", chunks)
	}
}

func TestAnthropicStreamerReturnsHTTPStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	streamer, err := NewAnthropicStreamer(AnthropicConfig{
		APIKey:  "anthropic-key",
		Model:   "claude-sonnet-4-5",
		BaseURL: server.URL + "/v1",
		Client:  server.Client(),
	})
	if err != nil {
		t.Fatalf("new streamer: %v", err)
	}
	result, err := streamer.Stream(context.Background(), StreamRequest{
		AgentName: "The Skeptical CFO",
		Topic:     "Rust rewrite",
	}, func(chunk string) error {
		t.Fatalf("unexpected chunk: %q", chunk)
		return nil
	})
	if err == nil {
		t.Fatalf("expected error, got result %q", result)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("error = %v", err)
	}
}

func TestAnthropicStreamerReturnsStreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: error\n"))
		_, _ = w.Write([]byte(`data: {"type":"error","error":{"type":"api_error","message":"overloaded"}}` + "\n\n"))
	}))
	defer server.Close()

	streamer, err := NewAnthropicStreamer(AnthropicConfig{
		APIKey:  "anthropic-key",
		Model:   "claude-sonnet-4-5",
		BaseURL: server.URL + "/v1",
		Client:  server.Client(),
	})
	if err != nil {
		t.Fatalf("new streamer: %v", err)
	}
	result, err := streamer.Stream(context.Background(), StreamRequest{
		AgentName: "The Skeptical CFO",
		Topic:     "Rust rewrite",
	}, func(chunk string) error {
		t.Fatalf("unexpected chunk: %q", chunk)
		return nil
	})
	if err == nil {
		t.Fatalf("expected error, got result %q", result)
	}
	if !strings.Contains(err.Error(), "overloaded") {
		t.Fatalf("error = %v", err)
	}
}

func TestNewAnthropicStreamerRequiresModel(t *testing.T) {
	_, err := NewAnthropicStreamer(AnthropicConfig{APIKey: "anthropic-key"})
	if err == nil {
		t.Fatal("expected error for empty model")
	}
	if !strings.Contains(err.Error(), "model") {
		t.Fatalf("error = %v", err)
	}
}
