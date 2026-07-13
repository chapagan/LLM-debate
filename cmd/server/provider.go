package main

import (
	"fmt"
	"strings"

	"llmdebate/internal/ai"
)

type streamerOptions struct {
	Provider      string
	OpenAIAPIKey  string
	OpenAIModel   string
	CursorAPIKey  string
	CursorBaseURL string
	CursorModel   string
}

func selectStreamer(opts streamerOptions) (ai.Streamer, string, error) {
	provider := strings.ToLower(strings.TrimSpace(opts.Provider))
	if provider == "" {
		provider = "mock"
	}

	switch provider {
	case "mock":
		return ai.NewMockStreamer(), "mock", nil
	case "openai":
		if opts.OpenAIAPIKey == "" {
			return nil, "", fmt.Errorf("AI_PROVIDER=openai requires OPENAI_API_KEY")
		}
		return ai.NewOpenAIStreamer(ai.OpenAIConfig{
			APIKey: opts.OpenAIAPIKey,
			Model:  opts.OpenAIModel,
		}), "openai", nil
	case "cursor":
		if opts.CursorAPIKey == "" {
			return nil, "", fmt.Errorf("AI_PROVIDER=cursor requires CURSOR_API_KEY")
		}
		if opts.CursorBaseURL == "" {
			return nil, "", fmt.Errorf("AI_PROVIDER=cursor requires CURSOR_BASE_URL")
		}
		streamer, err := ai.NewChatCompletionsStreamer(ai.ChatCompletionsConfig{
			APIKey:  opts.CursorAPIKey,
			Model:   opts.CursorModel,
			BaseURL: opts.CursorBaseURL,
		})
		if err != nil {
			return nil, "", err
		}
		return streamer, "cursor", nil
	default:
		return nil, "", fmt.Errorf("unsupported AI_PROVIDER %q (want mock, openai, or cursor)", opts.Provider)
	}
}
