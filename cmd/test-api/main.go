package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ChatRequest struct {
	Messages  []Message `json:"messages"`
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Stream    bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	req := ChatRequest{
		Messages:  []Message{{Role: "user", Content: "say hello"}},
		Model:     "claude-opus-4",
		MaxTokens: 100,
		Stream:    true,
	}

	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", "https://api.z.ai/api/anthropic/v1/messages", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", "2e36da42d9a047cfa8f56831161f808d.K3KAwj5kDAGrW0AS")
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Fprintf(os.Stderr, "Status: %d\n", resp.StatusCode)

	buf := make([]byte, 1)
	lineCount := 0
	for {
		_, err := resp.Body.Read(buf)
		if err != nil {
			break
		}
		fmt.Printf("%s", string(buf))
		lineCount++
		if lineCount > 5000 {
			break
		}
	}
	fmt.Fprintf(os.Stderr, "\nDone\n")
}
