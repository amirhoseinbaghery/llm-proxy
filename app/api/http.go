package api

import (
	"github.com/amirhoseinbaghery/llm-proxy/app/auth"
	health "github.com/amirhoseinbaghery/llm-proxy/app/health"
	"net/http"
)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", health.PingHandler)
	mux.HandleFunc("/register", auth.RegisterHandler)
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.Handle("/healthz", auth.JWTMiddleware(http.HandlerFunc(health.HealthzHandler)))

	return mux
}
