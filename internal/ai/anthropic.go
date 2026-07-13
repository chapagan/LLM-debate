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

type AnthropicConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Client  *http.Client
}

type AnthropicStreamer struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

const anthropicAPIVersion = "2023-06-01"
const anthropicMaxTokens = 1024

func NewAnthropicStreamer(cfg AnthropicConfig) (*AnthropicStreamer, error) {
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		return nil, errors.New("anthropic model is required")
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	client := cfg.Client
	if client == nil {
		client = defaultHTTPClient()
	}
	return &AnthropicStreamer{
		apiKey:  cfg.APIKey,
		model:   model,
		baseURL: baseURL,
		client:  client,
	}, nil
}

func (s *AnthropicStreamer) Stream(ctx context.Context, req StreamRequest, emit func(string) error) (string, error) {
	if s.apiKey == "" {
		return "", errors.New("anthropic api key is empty")
	}
	body, err := json.Marshal(map[string]any{
		"model":      s.model,
		"max_tokens": anthropicMaxTokens,
		"stream":     true,
		"system":     "You are participating in a two-agent executive debate. Output only the body of the current agent response. Do not include speaker labels, JSON, markdown fences, or role headers.",
		"messages":   buildAnthropicMessages(req),
	})
	if err != nil {
		return "", fmt.Errorf("marshal anthropic request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create anthropic request: %w", err)
	}
	httpReq.Header.Set("x-api-key", s.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("anthropic status %d: %s", resp.StatusCode, strings.TrimSpace(string(limited)))
	}
	return readAnthropicStream(ctx, resp.Body, emit)
}

func buildAnthropicMessages(req StreamRequest) []map[string]string {
	messages := []map[string]string{{
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

func readAnthropicStream(ctx context.Context, r io.Reader, emit func(string) error) (string, error) {
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
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return b.String(), fmt.Errorf("decode anthropic stream event: %w", err)
		}
		switch event.Type {
		case "content_block_delta":
			if event.Delta.Type != "text_delta" || event.Delta.Text == "" {
				continue
			}
			if err := emit(event.Delta.Text); err != nil {
				return b.String(), err
			}
			b.WriteString(event.Delta.Text)
		case "error":
			msg := event.Error.Message
			if msg == "" {
				msg = "anthropic stream error"
			}
			return b.String(), errors.New(msg)
		case "message_stop":
			return b.String(), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return b.String(), fmt.Errorf("read anthropic stream: %w", err)
	}
	return b.String(), nil
}
