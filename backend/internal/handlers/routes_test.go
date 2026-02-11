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
	Data    any    `json:"data"`
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

func TestAPICatalogEndpoint(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal/api-catalog", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAPISmokeCheckEndpoint(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/api-smoke-check", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAssessmentLifecycle(t *testing.T) {
	mux := setupMux()
	userID := "user-a"

	startBody := []byte(`{"topics_known":["arrays","strings"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests/diagnostic-1/start", bytes.NewBuffer(startBody))
	req.Header.Set("X-User-ID", userID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("start expected 202, got %d", rr.Code)
	}

	var startResp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&startResp); err != nil {
		t.Fatalf("decode start: %v", err)
	}
	data := startResp["data"].(map[string]any)
	attemptID := data["attempt_id"].(string)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/test-attempts/"+attemptID+"/next-question", nil)
	req.Header.Set("X-User-ID", userID)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("next-question expected 200, got %d", rr.Code)
	}
	var nq map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&nq)
	q := nq["data"].(map[string]any)["question"].(map[string]any)
	questionID := q["id"].(string)

	ansBody := []byte(`{"question_id":"` + questionID + `","selected_option":2}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/test-attempts/"+attemptID+"/answers", bytes.NewBuffer(ansBody))
	req.Header.Set("X-User-ID", userID)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("answer expected 202, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/test-attempts/"+attemptID+"/submit", nil)
	req.Header.Set("X-User-ID", userID)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("submit expected 202, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/test-attempts/"+attemptID+"/result", nil)
	req.Header.Set("X-User-ID", userID)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("result expected 200, got %d", rr.Code)
	}
}
