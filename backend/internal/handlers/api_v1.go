package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
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

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[3] == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "platform is required"})
		return
	}

	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "platform disconnect endpoint placeholder", Data: map[string]string{"platform": parts[3]}})
}

func (h *Handler) DashboardSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "dashboard summary endpoint placeholder"})
}

func (h *Handler) DashboardHeatmapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "dashboard heatmap endpoint placeholder"})
}

func (h *Handler) DashboardLeaderboardsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "dashboard leaderboard endpoint placeholder"})
}

func (h *Handler) TopicsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

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
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "topics list endpoint placeholder"})
}

func (h *Handler) TopicByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "topic details endpoint placeholder"})
}

func (h *Handler) SubtopicByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "subtopic details endpoint placeholder"})
}

func (h *Handler) CompleteLearningQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "complete learning question endpoint placeholder"})
}

func (h *Handler) CompleteSubtopicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "complete subtopic endpoint placeholder"})
}

func (h *Handler) SubtopicsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

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
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
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
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "test details endpoint placeholder"})
}

func (h *Handler) TestStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "start test endpoint placeholder"})
}

func (h *Handler) TestSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "active test session endpoint placeholder"})
}

func (h *Handler) TestsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

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
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "submit answer endpoint placeholder"})
}

func (h *Handler) AttemptSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "submit attempt endpoint placeholder"})
}

func (h *Handler) AttemptResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "attempt result endpoint placeholder"})
}

func (h *Handler) AttemptNextQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "next question endpoint placeholder"})
}

func (h *Handler) AttemptExpireHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "attempt expire endpoint placeholder"})
}

func (h *Handler) AttemptResumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "attempt resume endpoint placeholder"})
}

func (h *Handler) TestAttemptsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
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
