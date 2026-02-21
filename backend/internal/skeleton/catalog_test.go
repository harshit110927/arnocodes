package skeleton_test

import (
	"net/http"
	"testing"

	"github.com/harshit110927/arnocodes/backend/config"
	"github.com/harshit110927/arnocodes/backend/internal/assessment"
	"github.com/harshit110927/arnocodes/backend/internal/course"
	"github.com/harshit110927/arnocodes/backend/internal/dashboard"
	"github.com/harshit110927/arnocodes/backend/internal/handlers"
	"github.com/harshit110927/arnocodes/backend/internal/ide"
	"github.com/harshit110927/arnocodes/backend/internal/learning/activity"
	"github.com/harshit110927/arnocodes/backend/internal/skeleton"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRunSmokeCheckAllPass(t *testing.T) {
	pool := &pgxpool.Pool{}
	assessmentRepo := assessment.NewRepository(pool)
	h := handlers.NewHandler(
		&config.Config{Environment: "test"},
		assessmentRepo,
		course.NewCourseRepository(pool),
		handlers.NewAssessmentCourseStatusAdapter(assessmentRepo),
		dashboard.NewRepository(pool),
		ide.NewService(ide.NewRepository(pool), nil, activity.NewActivityRepository(pool)),
	)
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
