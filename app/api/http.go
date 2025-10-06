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
	mux.Handle("/register", auth.JWTMiddleware(auth.SuperuserOnly(http.HandlerFunc(auth.RegisterHandler))))
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.Handle("/healthz", auth.JWTMiddleware(http.HandlerFunc(health.HealthzHandler)))
	mux.Handle("/chat", auth.JWTMiddleware(http.HandlerFunc(chat.ChatHandler)))
	mux.Handle("/user", auth.JWTMiddleware(http.HandlerFunc(auth.GetUserHandler)))
	mux.Handle("/users", auth.JWTMiddleware(auth.SuperuserOnly(http.HandlerFunc(auth.ListUsersHandler))))
	mux.Handle("/delete", auth.JWTMiddleware(auth.SuperuserOnly(http.HandlerFunc(auth.DeleteUserHandler))))
	mux.Handle("/user/update/", auth.JWTMiddleware(auth.SuperuserOnly(http.HandlerFunc(auth.UpdateUserHandler))))

	return mux
}
