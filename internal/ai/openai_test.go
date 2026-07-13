package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultHTTPClientSetsResponseHeaderTimeout(t *testing.T) {
	client := defaultHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.Transport)
	}
	if transport.ResponseHeaderTimeout != defaultResponseHeaderTimeout {
		t.Fatalf("ResponseHeaderTimeout = %v, want %v", transport.ResponseHeaderTimeout, defaultResponseHeaderTimeout)
	}
	if client.Timeout != 0 {
		t.Fatalf("Timeout = %v, want 0 for streaming", client.Timeout)
	}
}

func TestMockStreamerEmitsDeterministicChunks(t *testing.T) {
	streamer := NewMockStreamer()
	var chunks []string
	result, err := streamer.Stream(context.Background(), StreamRequest{
		AgentName: "The Skeptical CFO",
		Topic:     "Rust rewrite",
		Messages:  []ChatMessage{{Role: "user", Content: "Opening context"}},
	}, func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected streamed chunks")
	}
	if !strings.Contains(result, "financial case") {
		t.Fatalf("result = %q", result)
	}
}

func TestOpenAIStreamerReadsResponsesDeltas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: response.output_text.delta\n"))
		_, _ = w.Write([]byte(`data: {"type":"response.output_text.delta","delta":"Hello"}` + "\n\n"))
		_, _ = w.Write([]byte("event: response.output_text.delta\n"))
		_, _ = w.Write([]byte(`data: {"type":"response.output_text.delta","delta":" world"}` + "\n\n"))
		_, _ = w.Write([]byte("event: response.completed\n"))
		_, _ = w.Write([]byte(`data: {"type":"response.completed"}` + "\n\n"))
	}))
	defer server.Close()

	streamer := NewOpenAIStreamer(OpenAIConfig{
		APIKey:  "test-key",
		Model:   "gpt-5.5",
		BaseURL: server.URL + "/v1",
		Client:  server.Client(),
	})
	var chunks []string
	result, err := streamer.Stream(context.Background(), StreamRequest{
		AgentName: "The Disruptive Visionary",
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

func TestOpenAIStreamerReturnsResponseFailedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: response.failed\n"))
		_, _ = w.Write([]byte(`data: {"type":"response.failed","response":{"error":{"message":"quota exhausted"}}}` + "\n\n"))
	}))
	defer server.Close()

	streamer := NewOpenAIStreamer(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL + "/v1",
		Client:  server.Client(),
	})
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
	if !strings.Contains(err.Error(), "quota exhausted") {
		t.Fatalf("error = %v", err)
	}
}

func TestOpenAIStreamerReturnsTopLevelErrorEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: error\n"))
		_, _ = w.Write([]byte(`data: {"type":"error","message":"bad request"}` + "\n\n"))
	}))
	defer server.Close()

	streamer := NewOpenAIStreamer(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL + "/v1",
		Client:  server.Client(),
	})
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
	if !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("error = %v", err)
	}
}
