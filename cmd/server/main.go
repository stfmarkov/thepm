package main

//go:generate templ generate

import (
	"log"
	"os"
	"strings"

	httpx "github.com/AdventurousNerd/thepm/internal/http"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	addr := listenAddr()

	engine, err := httpx.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}

// Render sets PORT (a bare number). Locally we use ADDR (":8090").
func listenAddr() string {
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		if !strings.HasPrefix(p, ":") {
			p = ":" + p
		}
		return p
	}
	if a := strings.TrimSpace(os.Getenv("ADDR")); a != "" {
		return a
	}
	return ":8080"
}
