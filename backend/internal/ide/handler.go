package ide

import (
	"encoding/json"
	"errors"
	"net/http"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

type SubmitRequest struct {
	AttemptID  *string `json:"attempt_id,omitempty"`
	QuestionID string  `json:"question_id"`
	Code       string  `json:"code"`
	Language   string  `json:"language"`
}

type RunRequest struct {
	QuestionID string `json:"question_id"`
	Code       string `json:"code"`
	Language   string `json:"language"`
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request, userID string) {
	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "invalid request body"})
		return
	}
	if req.QuestionID == "" || req.Code == "" || req.Language == "" {
		writeJSON(w, http.StatusUnprocessableEntity, APIResponse{Status: "error", Message: "question_id, code, language are required"})
		return
	}
	id, err := h.service.Submit(r.Context(), userID, Submission{AttemptID: req.AttemptID, QuestionID: req.QuestionID, Code: req.Code, Language: req.Language})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "question not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, APIResponse{Status: "error", Message: "failed to submit code"})
		return
	}
	writeJSON(w, http.StatusAccepted, APIResponse{Status: "ok", Message: "submission queued", Data: map[string]string{"submission_id": id}})
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request, userID string) {
	submissionID := r.URL.Query().Get("id")
	if submissionID == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "id is required"})
		return
	}
	status, err := h.service.Status(r.Context(), userID, submissionID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "submission not found"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "submission status", Data: status})
}

func (h *Handler) RunSample(w http.ResponseWriter, r *http.Request, userID string) {
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Status: "error", Message: "invalid request body"})
		return
	}
	if req.QuestionID == "" || req.Code == "" || req.Language == "" {
		writeJSON(w, http.StatusUnprocessableEntity, APIResponse{Status: "error", Message: "question_id, code, language are required"})
		return
	}
	res, err := h.service.RunSample(r.Context(), userID, Submission{QuestionID: req.QuestionID, Code: req.Code, Language: req.Language})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{Status: "error", Message: "question not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, APIResponse{Status: "error", Message: "failed to run sample tests"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Status: "ok", Message: "sample run result", Data: res})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
