package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/harshit110927/arnocodes/backend/config"
	"github.com/harshit110927/arnocodes/backend/internal/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize handlers
	h := handlers.NewHandler(cfg)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.HealthHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
