package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/harshit110927/arnocodes/backend/config"
)

type testResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func setupMux() *http.ServeMux {
	cfg := &config.Config{Environment: "test"}
	h := NewHandler(cfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func TestProfileMeSupportsGetAndPatch(t *testing.T) {
	mux := setupMux()

	for _, method := range []string{http.MethodGet, http.MethodPatch} {
		req := httptest.NewRequest(method, "/api/v1/profiles/me", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code == http.StatusMethodNotAllowed {
			t.Fatalf("expected %s to be allowed", method)
		}
	}
}

func TestUnversionedProfileRouteNotFound(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodGet, "/profiles/me", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unversioned route, got %d", rr.Code)
	}
}

func TestAIUsageGetOnly(t *testing.T) {
	mux := setupMux()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/usage", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestTestsSessionEndpoint(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tests/test-1/session", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var payload testResponse
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if payload.Status != "ok" {
		t.Fatalf("expected status ok, got %s", payload.Status)
	}
}

func TestTestAttemptsNextQuestionGetOnly(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test-attempts/attempt-1/next-question", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/test-attempts/attempt-1/next-question", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestCompleteSubtopicRequiresMasteryThreshold(t *testing.T) {
	mux := setupMux()

	low := []byte(`{"mastery_score":0.5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subtopics/sub-1/complete", bytes.NewBuffer(low))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for low mastery, got %d", rr.Code)
	}

	high := []byte(`{"mastery_score":0.9}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/subtopics/sub-1/complete", bytes.NewBuffer(high))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for valid mastery, got %d", rr.Code)
	}
}

func TestCourseStructureEndpoint(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/course/structure", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
