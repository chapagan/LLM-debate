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

type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Client  *http.Client
}

type OpenAIStreamer struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewOpenAIStreamer(cfg OpenAIConfig) *OpenAIStreamer {
	model := cfg.Model
	if model == "" {
		model = "gpt-5.5"
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	client := cfg.Client
	if client == nil {
		client = defaultHTTPClient()
	}
	return &OpenAIStreamer{apiKey: cfg.APIKey, model: model, baseURL: baseURL, client: client}
}

func (s *OpenAIStreamer) Stream(ctx context.Context, req StreamRequest, emit func(string) error) (string, error) {
	if s.apiKey == "" {
		return "", errors.New("openai api key is empty")
	}
	body, err := json.Marshal(map[string]any{
		"model":        s.model,
		"stream":       true,
		"input":        buildResponsesInput(req),
		"instructions": "You are participating in a two-agent executive debate. Output only the body of the current agent response. Do not include speaker labels, JSON, markdown fences, or role headers.",
	})
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create openai request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("openai status %d: %s", resp.StatusCode, strings.TrimSpace(string(limited)))
	}
	return readResponsesStream(ctx, resp.Body, emit)
}

func buildResponsesInput(req StreamRequest) []map[string]string {
	input := []map[string]string{{
		"role":    "user",
		"content": fmt.Sprintf("Topic: %s\nCurrent agent: %s\nPersona: %s", req.Topic, req.AgentName, req.Persona),
	}}
	for _, msg := range req.Messages {
		role := msg.Role
		if role != "assistant" {
			role = "user"
		}
		input = append(input, map[string]string{
			"role":    role,
			"content": msg.Content,
		})
	}
	return input
}

func readResponsesStream(ctx context.Context, r io.Reader, emit func(string) error) (string, error) {
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
			Type     string `json:"type"`
			Delta    string `json:"delta"`
			Message  string `json:"message"`
			Response struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			} `json:"response"`
		}
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return b.String(), fmt.Errorf("decode openai stream event: %w", err)
		}
		switch event.Type {
		case "response.output_text.delta":
			if event.Delta == "" {
				continue
			}
			if err := emit(event.Delta); err != nil {
				return b.String(), err
			}
			b.WriteString(event.Delta)
		case "error":
			return b.String(), errors.New(event.Message)
		case "response.failed":
			return b.String(), errors.New(event.Response.Error.Message)
		}
	}
	if err := scanner.Err(); err != nil {
		return b.String(), fmt.Errorf("read openai stream: %w", err)
	}
	return b.String(), nil
}
