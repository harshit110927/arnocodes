package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

func writeJSON(w http.ResponseWriter, statusCode int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
		Status:  "error",
		Message: "method not allowed",
	})
}

func pathParts(path string) []string {
	trimmed := strings.Trim(path, "/")
	if strings.HasPrefix(trimmed, strings.Trim(apiV1BasePath, "/")+"/") {
		trimmed = strings.TrimPrefix(trimmed, strings.Trim(apiV1BasePath, "/")+"/")
	} else if trimmed == strings.Trim(apiV1BasePath, "/") {
		trimmed = ""
	}

	if trimmed == "" {
		return []string{}
	}

	return strings.Split(trimmed, "/")
}

func currentUserID(r *http.Request) string {
	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userID == "" {
		return "demo-user"
	}
	return userID
}

func writeAssessmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, assessment.ErrTestNotFound), errors.Is(err, assessment.ErrAttemptNotFound), errors.Is(err, assessment.ErrInvalidQuestionID):
		writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: err.Error()})
	case errors.Is(err, assessment.ErrAttemptAlreadyDone):
		writeJSON(w, http.StatusForbidden, APIResponse{Status: "error", Message: err.Error()})
	case errors.Is(err, assessment.ErrNoQuestionsForTopics), errors.Is(err, assessment.ErrQuestionOutOfOrder):
		writeJSON(w, http.StatusUnprocessableEntity, APIResponse{Status: "error", Message: err.Error()})
	case errors.Is(err, assessment.ErrAttemptExpired):
		writeJSON(w, http.StatusGone, APIResponse{Status: "error", Message: err.Error()})
	case errors.Is(err, assessment.ErrAttemptNotActive):
		writeJSON(w, http.StatusConflict, APIResponse{Status: "error", Message: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, APIResponse{Status: "error", Message: "internal error"})
	}
}

func (h *Handler) ensureDiagnosticCompleted(w http.ResponseWriter, r *http.Request) bool {
	status := h.assessmentEngine.DiagnosticStatusForUser(currentUserID(r))
	if status.DiagnosticCompleted {
		return true
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "DIAGNOSTIC_REQUIRED"})
	return false
}

func (h *Handler) ProfileStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	status := h.assessmentEngine.DiagnosticStatusForUser(currentUserID(r))
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

func (h *Handler) TestByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 2 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "test id is required"})
		return
	}
	testID := parts[1]
	topics := []string{}
	if raw := r.URL.Query().Get("topics"); raw != "" {
		topics = strings.Split(raw, ",")
	}
	view, err := h.assessmentEngine.GetTestView(testID, topics)
	if err != nil {
		writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "test details", Data: view})
}

func (h *Handler) TestStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 2 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "test id is required"})
		return
	}
	testID := parts[1]
	var req struct {
		TopicsKnown []string `json:"topics_known"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	attempt, created, err := h.assessmentEngine.StartAttempt(currentUserID(r), testID, req.TopicsKnown)
	if err != nil {
		writeAssessmentError(w, err)
		return
	}
	msg := "attempt resumed"
	if created {
		msg = "attempt started"
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: msg, Data: map[string]interface{}{"attempt_id": attempt.ID, "status": attempt.Status, "topics_known": attempt.TopicsKnown}})
}

func (h *Handler) TestSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 2 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "test id is required"})
		return
	}
	attemptID := r.URL.Query().Get("attempt_id")
	if attemptID == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "attempt_id query param is required"})
		return
	}
	session, err := h.assessmentEngine.GetSession(currentUserID(r), attemptID)
	if err != nil {
		writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "active test session", Data: session})
}

func (h *Handler) TestsRouter(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path)

	if len(parts) == 2 && parts[0] == "tests" {
		h.TestByIDHandler(w, r)
		return
	}

	if len(parts) == 3 && parts[0] == "tests" && parts[2] == "start" {
		h.TestStartHandler(w, r)
		return
	}

	if len(parts) == 3 && parts[0] == "tests" && parts[2] == "session" {
		h.TestSessionHandler(w, r)
		return
	}

	writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "not found"})
}

func (h *Handler) AttemptAnswerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 3 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "attempt id is required"})
		return
	}
	attemptID := parts[1]
	var req struct {
		QuestionID     string `json:"question_id"`
		SelectedOption int    `json:"selected_option"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "invalid request body"})
		return
	}
	res, err := h.assessmentEngine.SubmitAnswer(currentUserID(r), attemptID, req.QuestionID, req.SelectedOption)
	if err != nil {
		writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "answer accepted", Data: res})
}

func (h *Handler) AttemptSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 3 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "attempt id is required"})
		return
	}
	attemptID := parts[1]
	result, err := h.assessmentEngine.SubmitAttempt(currentUserID(r), attemptID)
	if err != nil {
		writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "attempt submitted", Data: result})
}

func (h *Handler) AttemptResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 3 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "attempt id is required"})
		return
	}
	attemptID := parts[1]
	result, err := h.assessmentEngine.GetResult(currentUserID(r), attemptID)
	if err != nil {
		writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "attempt result", Data: result})
}

func (h *Handler) AttemptNextQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 3 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "attempt id is required"})
		return
	}
	attemptID := parts[1]
	question, err := h.assessmentEngine.NextQuestion(currentUserID(r), attemptID)
	if err != nil {
		writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "next question", Data: question})
}

func (h *Handler) AttemptExpireHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 3 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "attempt id is required"})
		return
	}
	attemptID := parts[1]
	if err := h.assessmentEngine.ExpireAttempt(currentUserID(r), attemptID); err != nil {
		writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "attempt expired"})
}

func (h *Handler) AttemptResumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 3 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "attempt id is required"})
		return
	}
	attemptID := parts[1]
	attempt, err := h.assessmentEngine.ResumeAttempt(currentUserID(r), attemptID)
	if err != nil {
		writeAssessmentError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "attempt resumed", Data: map[string]interface{}{"attempt_id": attempt.ID, "status": attempt.Status}})
}

func (h *Handler) TestAttemptsRouter(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path)
	if len(parts) != 3 || parts[0] != "test-attempts" {
		writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "not found"})
		return
	}

	switch parts[2] {
	case "answers":
		h.AttemptAnswerHandler(w, r)
	case "submit":
		h.AttemptSubmitHandler(w, r)
	case "result":
		h.AttemptResultHandler(w, r)
	case "next-question":
		h.AttemptNextQuestionHandler(w, r)
	case "expire":
		h.AttemptExpireHandler(w, r)
	case "resume":
		h.AttemptResumeHandler(w, r)
	default:
		writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "not found"})
	}
}

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
	report := skeleton.RunSmokeCheck(h.router)
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "api smoke check completed", Data: report})
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
