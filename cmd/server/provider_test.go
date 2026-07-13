package main

import (
	"strings"
	"testing"

	"llmdebate/internal/ai"
)

func TestDefaultListenAddrIsLoopback(t *testing.T) {
	if defaultListenAddr != "127.0.0.1:8080" {
		t.Fatalf("defaultListenAddr = %q, want 127.0.0.1:8080", defaultListenAddr)
	}
}

func TestSelectStreamerDefaultsToMock(t *testing.T) {
	streamer, name, err := selectStreamer(streamerOptions{Provider: ""})
	if err != nil {
		t.Fatalf("selectStreamer: %v", err)
	}
	if name != "mock" {
		t.Fatalf("name = %q, want mock", name)
	}
	if _, ok := streamer.(*ai.MockStreamer); !ok {
		t.Fatalf("streamer type = %T, want *ai.MockStreamer", streamer)
	}
}

func TestSelectStreamerOpenAI(t *testing.T) {
	streamer, name, err := selectStreamer(streamerOptions{
		Provider:     "openai",
		OpenAIAPIKey: "sk-test",
		OpenAIModel:  "gpt-5.5",
	})
	if err != nil {
		t.Fatalf("selectStreamer: %v", err)
	}
	if name != "openai" {
		t.Fatalf("name = %q, want openai", name)
	}
	if _, ok := streamer.(*ai.OpenAIStreamer); !ok {
		t.Fatalf("streamer type = %T, want *ai.OpenAIStreamer", streamer)
	}
}

func TestSelectStreamerCursor(t *testing.T) {
	streamer, name, err := selectStreamer(streamerOptions{
		Provider:       "cursor",
		CursorAPIKey:   "cursor-key",
		CursorBaseURL:  "http://127.0.0.1:9000/v1",
		CursorModel:    "composer-2.5",
	})
	if err != nil {
		t.Fatalf("selectStreamer: %v", err)
	}
	if name != "cursor" {
		t.Fatalf("name = %q, want cursor", name)
	}
	if _, ok := streamer.(*ai.ChatCompletionsStreamer); !ok {
		t.Fatalf("streamer type = %T, want *ai.ChatCompletionsStreamer", streamer)
	}
}

func TestSelectStreamerClaude(t *testing.T) {
	streamer, name, err := selectStreamer(streamerOptions{
		Provider:        "claude",
		AnthropicAPIKey: "anthropic-key",
		AnthropicModel:  "claude-sonnet-4-5",
	})
	if err != nil {
		t.Fatalf("selectStreamer: %v", err)
	}
	if name != "claude" {
		t.Fatalf("name = %q, want claude", name)
	}
	if _, ok := streamer.(*ai.AnthropicStreamer); !ok {
		t.Fatalf("streamer type = %T, want *ai.AnthropicStreamer", streamer)
	}
}

func TestSelectStreamerClaudeRequiresKeyAndModel(t *testing.T) {
	_, _, err := selectStreamer(streamerOptions{Provider: "claude", AnthropicModel: "claude-sonnet-4-5"})
	if err == nil {
		t.Fatal("expected error for missing Anthropic API key")
	}
	if !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Fatalf("error = %v", err)
	}

	_, _, err = selectStreamer(streamerOptions{Provider: "claude", AnthropicAPIKey: "anthropic-key"})
	if err == nil {
		t.Fatal("expected error for missing Anthropic model")
	}
	if !strings.Contains(err.Error(), "ANTHROPIC_MODEL") {
		t.Fatalf("error = %v", err)
	}
}

func TestSelectStreamerUnknownProvider(t *testing.T) {
	_, _, err := selectStreamer(streamerOptions{Provider: "anthropic"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if !strings.Contains(err.Error(), "AI_PROVIDER") {
		t.Fatalf("error = %v", err)
	}
}

func TestSelectStreamerOpenAIRequiresAPIKey(t *testing.T) {
	_, _, err := selectStreamer(streamerOptions{Provider: "openai"})
	if err == nil {
		t.Fatal("expected error for missing OpenAI API key")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("error = %v", err)
	}
}

func TestSelectStreamerCursorRequiresKeyBaseURLAndModel(t *testing.T) {
	_, _, err := selectStreamer(streamerOptions{Provider: "cursor", CursorAPIKey: "cursor-key", CursorModel: "composer-2.5"})
	if err == nil {
		t.Fatal("expected error for missing Cursor base URL")
	}
	if !strings.Contains(err.Error(), "CURSOR_BASE_URL") {
		t.Fatalf("error = %v", err)
	}

	_, _, err = selectStreamer(streamerOptions{Provider: "cursor", CursorBaseURL: "http://127.0.0.1:9000/v1", CursorModel: "composer-2.5"})
	if err == nil {
		t.Fatal("expected error for missing Cursor API key")
	}
	if !strings.Contains(err.Error(), "CURSOR_API_KEY") {
		t.Fatalf("error = %v", err)
	}

	_, _, err = selectStreamer(streamerOptions{
		Provider:      "cursor",
		CursorAPIKey:  "cursor-key",
		CursorBaseURL: "http://127.0.0.1:9000/v1",
	})
	if err == nil {
		t.Fatal("expected error for missing Cursor model")
	}
	if !strings.Contains(err.Error(), "CURSOR_MODEL") {
		t.Fatalf("error = %v", err)
	}
}

func TestSelectStreamerIgnoresOtherProviderKeys(t *testing.T) {
	streamer, name, err := selectStreamer(streamerOptions{
		Provider:       "mock",
		OpenAIAPIKey:   "sk-test",
		CursorAPIKey:   "cursor-key",
		CursorBaseURL:  "http://127.0.0.1:9000/v1",
	})
	if err != nil {
		t.Fatalf("selectStreamer: %v", err)
	}
	if name != "mock" {
		t.Fatalf("name = %q, want mock", name)
	}
	if _, ok := streamer.(*ai.MockStreamer); !ok {
		t.Fatalf("streamer type = %T, want *ai.MockStreamer", streamer)
	}
}
