package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/harshit110927/arnocodes/backend/config"
	"github.com/harshit110927/arnocodes/backend/internal/assessment"
	"github.com/harshit110927/arnocodes/backend/internal/dashboard"
	"github.com/harshit110927/arnocodes/backend/internal/ide"
	"github.com/harshit110927/arnocodes/backend/internal/learning"
)

type Handler struct {
	config            *config.Config
	router            http.Handler
	assessmentRepo    *assessment.Repository
	assessmentService *assessment.Service
	learningRepo      *learning.Repository
	dashboardRepo     *dashboard.Repository
	ideService        *ide.Service
}

func NewHandler(cfg *config.Config, assessmentRepo *assessment.Repository, learningRepo *learning.Repository, dashboardRepo *dashboard.Repository, ideService *ide.Service) *Handler {
	return &Handler{
		config:            cfg,
		assessmentRepo:    assessmentRepo,
		assessmentService: assessment.NewService(assessmentRepo),
		learningRepo:      learningRepo,
		dashboardRepo:     dashboardRepo,
		ideService:        ideService,
	}
}

type HealthResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{Status: "healthy", Environment: h.config.Environment}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
