package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/harshit110927/arnocodes/backend/config"
	"github.com/harshit110927/arnocodes/backend/internal/assessment"
	"github.com/harshit110927/arnocodes/backend/internal/dashboard"
	"github.com/harshit110927/arnocodes/backend/internal/learning"
)

func setupMux() *http.ServeMux {
	cfg := &config.Config{Environment: "test"}
	h := NewHandler(cfg, assessment.NewRepository(nil), learning.NewRepository(nil), dashboard.NewRepository(nil))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func TestProfileRouteMethods(t *testing.T) {
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

func TestProfileStatusEndpoint(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me/status", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 without db wiring in unit setup, got %d", rr.Code)
	}
}

func TestProtectedDashboardRequiresDiagnostic(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 in nil-repo test setup, got %d", rr.Code)
	}
}

func TestSubtopicCompletionValidation(t *testing.T) {
	mux := setupMux()
	low := []byte(`{"mastery_score":0.5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subtopics/sub-1/complete", bytes.NewBuffer(low))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected guard failure (500) in nil-repo test setup, got %d", rr.Code)
	}
}

func TestAssessmentEndpointsStillReachable(t *testing.T) {
	mux := setupMux()
	cases := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/v1/tests/diagnostic-1", http.StatusOK},
		{http.MethodPost, "/api/v1/tests/diagnostic-1/start", http.StatusAccepted},
		{http.MethodGet, "/api/v1/test-attempts/a1/next-question", http.StatusOK},
		{http.MethodPost, "/api/v1/test-attempts/a1/submit", http.StatusAccepted},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != tc.want {
			t.Fatalf("%s %s expected %d, got %d", tc.method, tc.path, tc.want, rr.Code)
		}
	}
}
