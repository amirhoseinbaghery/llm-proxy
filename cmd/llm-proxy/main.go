package main

import (
	"errors"
	"github.com/amirhoseinbaghery/llm-proxy/app/api"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := api.NewMux()

	addr := getEnv("LISTEN_ADDR", ":8080")
	srv := &http.Server{Addr: addr, Handler: mux}

	log.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
