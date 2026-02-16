package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/harshit110927/arnocodes/backend/internal/assessment"
	"github.com/harshit110927/arnocodes/backend/internal/skeleton"
)

const apiV1BasePath = "/api/v1"
const masteryCompletionThreshold = 0.70

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type subtopicCompletionRequest struct {
	MasteryScore float64 `json:"mastery_score"`
}

type startDiagnosticRequest struct {
	SelectedTopics []string `json:"selected_topics"`
}

type answerDiagnosticRequest struct {
	QuestionID     string `json:"question_id"`
	QuestionType   string `json:"question_type"`
	SelectedOption *int   `json:"selected_option,omitempty"`
	Code           string `json:"code,omitempty"`
	Language       string `json:"language,omitempty"`
}

func writeJSON(w http.ResponseWriter, statusCode int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeErrorCode(w http.ResponseWriter, statusCode int, errCode string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": errCode})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Status: "error", Message: "method not allowed"})
}

func pathParts(path string) []string {
	trimmed := strings.Trim(path, "/")
	prefix := strings.Trim(apiV1BasePath, "/") + "/"
	if strings.HasPrefix(trimmed, prefix) {
		trimmed = strings.TrimPrefix(trimmed, prefix)
	} else if trimmed == strings.Trim(apiV1BasePath, "/") {
		trimmed = ""
	}
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

func currentUserID(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("X-User-ID"))
}

func requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	uid := currentUserID(r)
	if uid == "" {
		writeErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED")
		return "", false
	}
	return uid, true
}

func (h *Handler) ensureDiagnosticCompleted(w http.ResponseWriter, r *http.Request) bool {
	userID, ok := requireUserID(w, r)
	if !ok {
		return false
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	status, err := h.assessmentRepo.GetUserStatus(ctx, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Status: "error", Message: "failed to check diagnostic status"})
		return false
	}
	if status.DiagnosticCompleted {
		return true
	}
	writeErrorCode(w, http.StatusForbidden, "DIAGNOSTIC_REQUIRED")
	return false
}

func (h *Handler) ProfileStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	status, err := h.assessmentRepo.GetUserStatus(ctx, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Status: "error", Message: "failed to read profile status"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "profile status", Data: status})
}

func (h *Handler) ProfileMeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "profile fetch endpoint placeholder"})
	case http.MethodPatch:
		writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "profile update endpoint placeholder"})
	default:
		methodNotAllowed(w)
	}
}

func (h *Handler) PlatformConnectionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "platform list endpoint placeholder"})
	case http.MethodPost:
		writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "platform connect endpoint placeholder"})
	default:
		methodNotAllowed(w)
	}
}

func (h *Handler) PlatformConnectionByNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 4 || parts[3] == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "platform is required"})
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "platform disconnect endpoint placeholder", Data: map[string]string{"platform": parts[3]}})
}

func (h *Handler) DashboardSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "dashboard summary endpoint placeholder"})
}
func (h *Handler) DashboardHeatmapHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "dashboard heatmap endpoint placeholder"})
}
func (h *Handler) DashboardLeaderboardsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "dashboard leaderboard endpoint placeholder"})
}

func (h *Handler) CourseStructureHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "course structure endpoint placeholder"})
}

func (h *Handler) TopicsRouter(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path)
	if len(parts) == 1 && parts[0] == "topics" {
		h.TopicsHandler(w, r)
		return
	}
	if len(parts) == 3 && parts[0] == "topics" && parts[2] == "unlock-status" {
		h.TopicUnlockStatusHandler(w, r)
		return
	}
	if len(parts) == 2 && parts[0] == "topics" {
		h.TopicByIDHandler(w, r)
		return
	}
	writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "not found"})
}

func (h *Handler) TopicUnlockStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "topic unlock status endpoint placeholder"})
}
func (h *Handler) TopicsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "topics list endpoint placeholder"})
}
func (h *Handler) TopicByIDHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "topic details endpoint placeholder"})
}
func (h *Handler) SubtopicByIDHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "subtopic details endpoint placeholder"})
}
func (h *Handler) CompleteLearningQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "complete learning question endpoint placeholder"})
}
func (h *Handler) CompleteSubtopicHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureDiagnosticCompleted(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req subtopicCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "invalid request body"})
		return
	}
	if req.MasteryScore < masteryCompletionThreshold {
		writeJSON(w, http.StatusUnprocessableEntity, APIResponse{Status: "error", Message: "mastery threshold not met; completion denied"})
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "subtopic completion accepted; pending server validation"})
}

func (h *Handler) SubtopicsRouter(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path)
	if len(parts) == 2 && parts[0] == "subtopics" {
		h.SubtopicByIDHandler(w, r)
		return
	}
	if len(parts) == 3 && parts[0] == "subtopics" && parts[2] == "complete" {
		h.CompleteSubtopicHandler(w, r)
		return
	}
	writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "not found"})
}
func (h *Handler) LearningQuestionsRouter(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path)
	if len(parts) == 4 && parts[0] == "learning" && parts[1] == "questions" && parts[3] == "complete" {
		h.CompleteLearningQuestionHandler(w, r)
		return
	}
	writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "not found"})
}

func (h *Handler) DiagnosticStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req startDiagnosticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY")
		return
	}
	attemptID, err := h.assessmentService.StartDiagnostic(r.Context(), userID, req.SelectedTopics)
	if err != nil {
		h.writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, APIResponse{Status: "ok", Message: "diagnostic attempt started", Data: map[string]string{"attempt_id": attemptID}})
}

func (h *Handler) DiagnosticRouter(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path)
	if len(parts) == 2 && parts[0] == "diagnostic" && parts[1] == "start" {
		h.DiagnosticStartHandler(w, r)
		return
	}
	if len(parts) != 3 || parts[0] != "diagnostic" {
		writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "not found"})
		return
	}
	switch parts[2] {
	case "next":
		h.DiagnosticNextHandler(w, r)
	case "answer":
		h.DiagnosticAnswerHandler(w, r)
	case "coding":
		h.DiagnosticCodingHandler(w, r)
	case "status":
		h.DiagnosticStatusHandler(w, r)
	case "submit":
		h.DiagnosticSubmitHandler(w, r)
	default:
		writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "not found"})
	}
}

func diagnosticAttemptIDFromPath(path string) string {
	parts := pathParts(path)
	if len(parts) >= 2 && parts[0] == "diagnostic" {
		return parts[1]
	}
	return ""
}

func (h *Handler) DiagnosticNextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	question, err := h.assessmentService.FetchNextQuestion(r.Context(), userID, diagnosticAttemptIDFromPath(r.URL.Path))
	if err != nil {
		h.writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "next diagnostic question", Data: question})
}

func (h *Handler) DiagnosticAnswerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req answerDiagnosticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY")
		return
	}
	submissionID, err := h.assessmentService.SubmitAnswer(r.Context(), userID, diagnosticAttemptIDFromPath(r.URL.Path), assessment.AnswerData{
		QuestionID:     req.QuestionID,
		QuestionType:   req.QuestionType,
		SelectedOption: req.SelectedOption,
		Code:           req.Code,
		Language:       req.Language,
	})
	if err != nil {
		h.writeAssessmentError(w, err)
		return
	}
	data := map[string]string{"accepted": "true"}
	if submissionID != "" {
		data["submission_id"] = submissionID
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "diagnostic answer accepted", Data: data})
}

func (h *Handler) DiagnosticCodingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	h.DiagnosticAnswerHandler(w, r)
}

func (h *Handler) DiagnosticStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	status, err := h.assessmentRepo.GetDiagnosticAttemptStatus(r.Context(), diagnosticAttemptIDFromPath(r.URL.Path))
	if err != nil {
		h.writeAssessmentError(w, err)
		return
	}
	if status.AttemptID == "" {
		writeErrorCode(w, http.StatusNotFound, "NOT_FOUND")
		return
	}
	if status.UserID != userID {
		writeErrorCode(w, http.StatusForbidden, "UNAUTHORIZED")
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "diagnostic status", Data: status})
}

func (h *Handler) DiagnosticSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	if err := h.assessmentService.SubmitTest(r.Context(), userID, diagnosticAttemptIDFromPath(r.URL.Path)); err != nil {
		h.writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "diagnostic submitted"})
}

func (h *Handler) writeAssessmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, assessment.ErrDiagnosticBlocked):
		writeErrorCode(w, http.StatusForbidden, "DIAGNOSTIC_BLOCKED")
	case errors.Is(err, assessment.ErrNotFound):
		writeErrorCode(w, http.StatusNotFound, "NOT_FOUND")
	case errors.Is(err, assessment.ErrInvalidInput):
		writeErrorCode(w, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY")
	case errors.Is(err, assessment.ErrUnauthorized):
		writeErrorCode(w, http.StatusForbidden, "UNAUTHORIZED")
	case errors.Is(err, assessment.ErrTimeExpired):
		writeErrorCode(w, http.StatusForbidden, "TIME_EXPIRED")
	default:
		writeJSON(w, http.StatusInternalServerError, APIResponse{Status: "error", Message: err.Error()})
	}
}

// Existing endpoints kept as placeholders for non-diagnostic flows.
func (h *Handler) PlatformSyncTriggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "platform sync trigger endpoint placeholder"})
}
func (h *Handler) PlatformSyncJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "platform sync job endpoint placeholder"})
}
func (h *Handler) AIQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "ai query endpoint placeholder"})
}
func (h *Handler) AICodeHelperStepHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "ai code helper step endpoint placeholder"})
}
func (h *Handler) AIUsageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "ai usage endpoint placeholder"})
}

func (h *Handler) APICatalogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "api catalog", Data: skeleton.APICatalog()})
}
func (h *Handler) APISmokeCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "api smoke check completed", Data: skeleton.RunSmokeCheck(h.router)})
}
func (h *Handler) InternalRecomputeDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "internal recompute dashboard endpoint placeholder"})
}
func (h *Handler) InternalRefreshLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "internal refresh leaderboard endpoint placeholder"})
}
