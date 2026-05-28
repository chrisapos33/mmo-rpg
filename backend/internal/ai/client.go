package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiURL  = "https://api.anthropic.com/v1/messages"
	version = "2023-06-01"
	model   = "claude-sonnet-4-6"
)

type Client struct {
	apiKey string
	http   *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, http: &http.Client{}}
}

type msgRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system,omitempty"`
	Messages  []msgContent `json:"messages"`
}

type msgContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type msgResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) Complete(ctx context.Context, system, user string, maxTokens int) (string, error) {
	body, err := json.Marshal(msgRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []msgContent{{Role: "user", Content: user}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", version)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var result msgResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from anthropic")
	}

	return result.Content[0].Text, nil
}
