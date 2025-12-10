package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// Chat sends a prompt to Ollama and returns the response text.
func Chat(model string, systemPrompt string, userPrompt string) (string, error) {
	reqBody := ollamaChatRequest{
		Model: model,
		Messages: []ollamaMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("POST", "http://localhost:11434/api/chat", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call ollama: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", res.StatusCode, string(bodyBytes))
	}

	var resp ollamaChatResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	out := resp.Message.Content
	return out, nil
}
