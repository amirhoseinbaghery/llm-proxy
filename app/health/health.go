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
	query := r.URL.Query()
	model := query.Get("model")
	apiKey := query.Get("apikey")

	if model == "" || apiKey == "" {
		http.Error(w, "model and apikey required", http.StatusBadRequest)
		return
	}

	base := os.Getenv("LLM_BASE_URL")
	if base == "" {
		http.Error(w, "LLM_BASE_URL not set", http.StatusServiceUnavailable)
		return
	}

	endpoint := base + "/chat/completions"

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "if you are ok say 'I'm alive, damn it'"},
		},
		"max_tokens": 1,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		http.Error(w, "request build error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "llm unreachable: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}
