package main

//go:generate templ generate

import (
	"log"
	"os"

	httpx "github.com/AdventurousNerd/thepm/internal/http"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	engine, err := httpx.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
