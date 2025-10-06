package chat

import (
	"bytes"
	"encoding/json"
	"github.com/amirhoseinbaghery/llm-proxy/app/health"
	"io"
	"net/http"
	"os"
)

var httpClient = &http.Client{}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Model       string              `json:"model"`
		Messages    []map[string]string `json:"messages"`
		MaxTokens   int                 `json:"max_tokens"`
		Temperature float64             `json:"temperature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	apiKey, ok := health.AllowedModels[input.Model]
	if !ok {
		http.Error(w, "model not allowed", http.StatusBadRequest)
		return
	}

	base := os.Getenv("LLM_BASE_URL")
	if base == "" {
		http.Error(w, "LLM_BASE_URL not set", http.StatusServiceUnavailable)
		return
	}

	endpoint := base + "/chat/completions"

	payload := map[string]interface{}{
		"model":       input.Model,
		"messages":    input.Messages,
		"max_tokens":  input.MaxTokens,
		"temperature": input.Temperature,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		http.Error(w, "request error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
