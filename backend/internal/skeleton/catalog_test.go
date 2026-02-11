package skeleton_test

import (
	"net/http"
	"testing"

	"github.com/harshit110927/arnocodes/backend/config"
	"github.com/harshit110927/arnocodes/backend/internal/handlers"
	"github.com/harshit110927/arnocodes/backend/internal/skeleton"
)

func TestRunSmokeCheckAllPass(t *testing.T) {
	h := handlers.NewHandler(&config.Config{Environment: "test"})
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	report := skeleton.RunSmokeCheck(mux)
	if report.Failed != 0 {
		for _, r := range report.Results {
			if !r.Passed {
				t.Logf("failed: %s %s expected=%d actual=%d", r.Method, r.Path, r.ExpectedStatus, r.ActualStatus)
			}
		}
		t.Fatalf("expected 0 failed checks, got %d", report.Failed)
	}
	if report.Total == 0 {
		t.Fatalf("expected non-zero catalog")
	}
}
