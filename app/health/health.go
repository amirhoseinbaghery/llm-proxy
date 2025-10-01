package health

import (
	"io"
	"net/http"
	"os"
	"time"
)

func PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

var httpClient = &http.Client{Timeout: 2 * time.Second}

func HealthzHandler(w http.ResponseWriter, _ *http.Request) {
	base := os.Getenv("LLM_BASE_URL")
	if base == "" {
		http.Error(w, "LLM_BASE_URL not set", http.StatusServiceUnavailable)
		return
	}

	req, err := http.NewRequest(http.MethodGet, base+"/models", nil)
	if err != nil {
		http.Error(w, "request build error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if bearer := os.Getenv("LLM_HEALTH_BEARER"); bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

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
