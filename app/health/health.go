package health

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

var httpClient = &http.Client{}

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"status":"unhealthy","error":"Only POST method allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"status":"unhealthy","error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	apiKey, ok := AllowedModels[input.Model]
	if !ok {
		http.Error(w, `{"status":"unhealthy","error":"model not allowed"}`, http.StatusBadRequest)
		return
	}

	base := os.Getenv("LLM_BASE_URL")
	if base == "" {
		http.Error(w, `{"status":"unhealthy","error":"LLM_BASE_URL not set"}`, http.StatusServiceUnavailable)
		return
	}

	endpoint := base + "/chat/completions"

	payload := map[string]interface{}{
		"model": input.Model,
		"messages": []map[string]string{
			{"role": "user", "content": "healthcheck"},
		},
		"max_tokens": 1,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		http.Error(w, `{"status":"unhealthy","error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, `{"status":"unhealthy","error":"`+err.Error()+`"}`, http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unhealthy",
			"error":  string(body),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"model":   input.Model,
		"message": "I'm alive, damn it",
	})
}
