package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
	Model       string    `json:"model,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

func main() {
	client := &http.Client{}
	apiKey := "01488c4034ac47648997fe706ac3e367.g1rUs7OJOHefOU5x"

	// Try different models
	models := []string{"glm-4", "glm-4-plus", "glm-4.7", "glm-4-flash", "claude-opus-4"}

	for _, model := range models {
		fmt.Fprintf(os.Stderr, "\n--- Testing model: %s ---\n", model)

		req := ChatRequest{
			Model:     model,
			Messages:  []Message{{Role: "user", Content: "hello"}},
			MaxTokens: 20,
			Stream:    false,
		}
		body, _ := json.Marshal(req)

		url := "https://api.z.ai/api/coding/paas/v4/chat/completions"
		httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Fprintf(os.Stderr, "  Status: %d\n", resp.StatusCode)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("\n✓ SUCCESS with model: %s\n", model)
			fmt.Printf("Response: %s\n", string(bodyBytes))
			os.Exit(0)
		} else {
			if len(bodyBytes) > 300 {
				bodyBytes = bodyBytes[:300]
			}
			fmt.Fprintf(os.Stderr, "  Response: %s\n", string(bodyBytes))
		}
	}

	fmt.Fprintf(os.Stderr, "\n✗ No working model found\n")
}

