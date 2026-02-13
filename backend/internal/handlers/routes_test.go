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

func TestDiagnosticStartRequiresUserID(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/diagnostic/start", bytes.NewBufferString(`{"selected_topics":["a"]}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", rr.Code)
	}
}

func TestDiagnosticStartInvalidBody(t *testing.T) {
	mux := setupMux()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/diagnostic/start", bytes.NewBufferString(`{`))
	req.Header.Set("X-User-ID", "00000000-0000-0000-0000-000000000001")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 got %d", rr.Code)
	}
}

func TestDiagnosticAttemptRoutesExist(t *testing.T) {
	mux := setupMux()
	cases := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/v1/diagnostic/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/next", http.StatusUnauthorized},
		{http.MethodPost, "/api/v1/diagnostic/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/answer", http.StatusUnauthorized},
		{http.MethodPost, "/api/v1/diagnostic/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/coding", http.StatusUnauthorized},
		{http.MethodGet, "/api/v1/diagnostic/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/status", http.StatusUnauthorized},
		{http.MethodPost, "/api/v1/diagnostic/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/submit", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != tc.want {
			t.Fatalf("%s %s expected %d got %d", tc.method, tc.path, tc.want, rr.Code)
		}
	}
}
