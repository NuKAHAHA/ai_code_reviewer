package llama

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type LlamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type LlamaChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func Generate(prompt string) (string, error) {
	reqBody := LlamaRequest{
		Model:  "llama3.1",
		Prompt: prompt,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "http://localhost:11434/api/generate", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// NDJSON chunked stream, читаем построчно
	decoder := json.NewDecoder(resp.Body)

	full := ""

	for {
		var chunk LlamaChunk
		if err := decoder.Decode(&chunk); err != nil {
			break
		}
		full += chunk.Response
		if chunk.Done {
			break
		}
	}

	return full, nil
}
