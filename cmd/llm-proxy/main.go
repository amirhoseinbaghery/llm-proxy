package main

import (
	"database/sql"
	"errors"
	"github.com/amirhoseinbaghery/llm-proxy/app/api"
	"github.com/amirhoseinbaghery/llm-proxy/app/auth"
	"log"
	"net/http"
	"os"
)

func main() {
	auth.InitDB("/data/llm-proxy.db")
	defer func(DB *sql.DB) {
		if err := DB.Close(); err != nil {
			log.Println("failed to close DB:", err)
		}
	}(auth.DB)

	if _, err := auth.GetUserByUsername("admin"); err != nil {
		hashed, _ := auth.HashPassword("2wsxdr5#E$")
		if err := auth.CreateUser("admin", hashed, true); err != nil {
			log.Println("failed to create default user:", err)
		} else {
			log.Println("default admin user created (username: admin, password: *****)")
		}
	}

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
