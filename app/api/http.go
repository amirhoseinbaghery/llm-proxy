package api

import (
	"github.com/amirhoseinbaghery/llm-proxy/app/auth"
	"github.com/amirhoseinbaghery/llm-proxy/app/chat"
	health "github.com/amirhoseinbaghery/llm-proxy/app/health"
	"net/http"
)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", health.PingHandler)
	mux.HandleFunc("/register", auth.RegisterHandler)
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.Handle("/healthz", auth.JWTMiddleware(http.HandlerFunc(health.HealthzHandler)))
	mux.Handle("/chat", auth.JWTMiddleware(http.HandlerFunc(chat.ChatHandler)))

	return mux
}
