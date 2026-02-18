package assessment

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/harshit110927/arnocodes/backend/internal/database"
)

func TestDiagnosticFlowIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := database.New(ctx, dsn)
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	defer db.Close()
	if err := database.RunMigrations(ctx, db); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	if err := database.RunSeed(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	userID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	_, _ = db.Pool().Exec(ctx, `INSERT INTO auth.users(id) VALUES ($1::uuid) ON CONFLICT DO NOTHING`, userID)
	_, _ = db.Pool().Exec(ctx, `INSERT INTO profiles(id, full_name) VALUES ($1::uuid, 'Integration User') ON CONFLICT DO NOTHING`, userID)

	repo := NewRepository(db.Pool())
	svc := NewService(repo)

	attemptID, err := svc.StartDiagnostic(ctx, userID, []string{"22222222-2222-2222-2222-222222222221", "22222222-2222-2222-2222-222222222222"})
	if err != nil {
		t.Fatalf("start diagnostic: %v", err)
	}

	seenCoding := false
	seenMCQ := false
	for i := 0; i < 8; i++ {
		q, err := svc.FetchNextQuestion(ctx, userID, attemptID)
		if err != nil {
			break
		}
		switch q.QuestionType {
		case "mcq":
			opt := 1
			if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "mcq", SelectedOption: &opt}); err != nil {
				t.Fatalf("submit mcq: %v", err)
			}
			seenMCQ = true
		case "coding":
			if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "coding", Code: "func solve(){}", Language: "go"}); err != nil {
				t.Fatalf("submit coding: %v", err)
			}
			seenCoding = true
		}
	}
	if !seenMCQ {
		t.Fatalf("expected at least one mcq question")
	}
	if !seenCoding {
		t.Fatalf("expected at least one coding question")
	}

	workerCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := svc.HandleCodingEvaluationWorker(workerCtx, 20); err != nil {
		t.Fatalf("worker cycle: %v", err)
	}

	if err := svc.SubmitTest(ctx, userID, attemptID); err != nil {
		t.Fatalf("submit test: %v", err)
	}

	status, err := repo.GetDiagnosticAttemptStatus(ctx, attemptID)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Status != "submitted" {
		t.Fatalf("expected submitted status, got %s", status.Status)
	}
}
