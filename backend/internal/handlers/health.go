package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/harshit110927/arnocodes/backend/config"
	"github.com/harshit110927/arnocodes/backend/internal/assessment"
	"github.com/harshit110927/arnocodes/backend/internal/course"
	"github.com/harshit110927/arnocodes/backend/internal/dashboard"
	"github.com/harshit110927/arnocodes/backend/internal/ide"
)

type assessmentCourseStatusAdapter struct {
	repo *assessment.Repository
}

func NewAssessmentCourseStatusAdapter(repo *assessment.Repository) course.DiagnosticStatusProvider {
	return assessmentCourseStatusAdapter{repo: repo}
}

func (a assessmentCourseStatusAdapter) GetUserStatus(ctx context.Context, userID string) (course.DiagnosticUserStatus, error) {
	status, err := a.repo.GetUserStatus(ctx, userID)
	if err != nil {
		return course.DiagnosticUserStatus{}, err
	}
	return course.DiagnosticUserStatus{DiagnosticCompleted: status.DiagnosticCompleted}, nil
}

type AuthProtector interface {
	Middleware(next http.Handler) http.Handler
}

type Handler struct {
	config            *config.Config
	router            http.Handler
	assessmentRepo    *assessment.Repository
	assessmentService *assessment.Service
	courseRepo        *course.CourseRepository
	courseService     *course.CourseService
	dashboardRepo     *dashboard.Repository
	ideService        *ide.Service
	authMiddleware    AuthProtector
}

func NewHandler(cfg *config.Config, assessmentRepo *assessment.Repository, courseRepo *course.CourseRepository, courseStatusProvider course.DiagnosticStatusProvider, dashboardRepo *dashboard.Repository, ideService *ide.Service, authMiddleware AuthProtector) *Handler {
	return &Handler{
		config:            cfg,
		assessmentRepo:    assessmentRepo,
		assessmentService: assessment.NewService(assessmentRepo),
		courseRepo:        courseRepo,
		courseService:     course.NewCourseService(courseRepo, courseStatusProvider),
		dashboardRepo:     dashboardRepo,
		ideService:        ideService,
		authMiddleware:    authMiddleware,
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
