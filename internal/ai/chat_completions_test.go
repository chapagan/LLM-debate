package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatCompletionsStreamerReadsDeltaContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer cursor-key" {
			t.Fatalf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":"Hello"}}]}` + "\n\n"))
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":" world"}}]}` + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	streamer, err := NewChatCompletionsStreamer(ChatCompletionsConfig{
		APIKey:  "cursor-key",
		Model:   "composer-2.5",
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

func TestChatCompletionsStreamerReturnsHTTPStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	streamer, err := NewChatCompletionsStreamer(ChatCompletionsConfig{
		APIKey:  "cursor-key",
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

func TestChatCompletionsStreamerReturnsStreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`data: {"error":{"message":"model unavailable"}}` + "\n\n"))
	}))
	defer server.Close()

	streamer, err := NewChatCompletionsStreamer(ChatCompletionsConfig{
		APIKey:  "cursor-key",
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
	if !strings.Contains(err.Error(), "model unavailable") {
		t.Fatalf("error = %v", err)
	}
}

func TestNewChatCompletionsStreamerRequiresBaseURL(t *testing.T) {
	_, err := NewChatCompletionsStreamer(ChatCompletionsConfig{APIKey: "cursor-key"})
	if err == nil {
		t.Fatal("expected error for empty base URL")
	}
	if !strings.Contains(err.Error(), "base URL") {
		t.Fatalf("error = %v", err)
	}
}
