package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/harshit110927/arnocodes/backend/config"
)

type Handler struct {
	config *config.Config
	router http.Handler
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
	}
}

type HealthResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:      "healthy",
		Environment: h.config.Environment,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
