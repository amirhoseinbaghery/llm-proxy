package api

import (
	health "github.com/amirhoseinbaghery/llm-proxy/app/health"
	"net/http"
)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", health.PingHandler)
	mux.HandleFunc("/healthz", health.HealthzHandler)
	return mux
}
