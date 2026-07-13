package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ChatCompletionsConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Client  *http.Client
}

type ChatCompletionsStreamer struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewChatCompletionsStreamer(cfg ChatCompletionsConfig) (*ChatCompletionsStreamer, error) {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		return nil, errors.New("chat completions base URL is required")
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		return nil, errors.New("chat completions model is required")
	}
	client := cfg.Client
	if client == nil {
		client = defaultHTTPClient()
	}
	return &ChatCompletionsStreamer{
		apiKey:  cfg.APIKey,
		model:   model,
		baseURL: baseURL,
		client:  client,
	}, nil
}

func (s *ChatCompletionsStreamer) Stream(ctx context.Context, req StreamRequest, emit func(string) error) (string, error) {
	if s.apiKey == "" {
		return "", errors.New("chat completions api key is empty")
	}
	body, err := json.Marshal(map[string]any{
		"model":    s.model,
		"stream":   true,
		"messages": buildChatCompletionsMessages(req),
	})
	if err != nil {
		return "", fmt.Errorf("marshal chat completions request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create chat completions request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("chat completions request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("chat completions status %d: %s", resp.StatusCode, strings.TrimSpace(string(limited)))
	}
	return readChatCompletionsStream(ctx, resp.Body, emit)
}

func buildChatCompletionsMessages(req StreamRequest) []map[string]string {
	messages := []map[string]string{{
		"role":    "system",
		"content": "You are participating in a two-agent executive debate. Output only the body of the current agent response. Do not include speaker labels, JSON, markdown fences, or role headers.",
	}, {
		"role":    "user",
		"content": fmt.Sprintf("Topic: %s\nCurrent agent: %s\nPersona: %s", req.Topic, req.AgentName, req.Persona),
	}}
	for _, msg := range req.Messages {
		role := msg.Role
		if role != "assistant" {
			role = "user"
		}
		messages = append(messages, map[string]string{
			"role":    role,
			"content": msg.Content,
		})
	}
	return messages
}

func readChatCompletionsStream(ctx context.Context, r io.Reader, emit func(string) error) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	var b strings.Builder
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return b.String(), ctx.Err()
		default:
		}
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}
		var event struct {
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return b.String(), fmt.Errorf("decode chat completions stream event: %w", err)
		}
		if event.Error != nil && event.Error.Message != "" {
			return b.String(), errors.New(event.Error.Message)
		}
		for _, choice := range event.Choices {
			if choice.Delta.Content == "" {
				continue
			}
			if err := emit(choice.Delta.Content); err != nil {
				return b.String(), err
			}
			b.WriteString(choice.Delta.Content)
		}
	}
	if err := scanner.Err(); err != nil {
		return b.String(), fmt.Errorf("read chat completions stream: %w", err)
	}
	return b.String(), nil
}
