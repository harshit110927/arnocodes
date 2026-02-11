package handlers

import (
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
		req := httptest.NewRequest(method, "/profiles/me", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code == http.StatusMethodNotAllowed {
			t.Fatalf("expected %s to be allowed", method)
		}
	}
}

func TestAIUsageGetOnly(t *testing.T) {
	mux := setupMux()

	req := httptest.NewRequest(http.MethodGet, "/ai/usage", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/ai/usage", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestTestsSessionEndpoint(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodGet, "/tests/test-1/session", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/test-attempts/attempt-1/next-question", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/test-attempts/attempt-1/next-question", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
