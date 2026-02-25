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

	var diagnosticEvents int
	if err := db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM user_activity_feed WHERE user_id=$1::uuid AND source='diagnostic'`, userID).Scan(&diagnosticEvents); err != nil {
		t.Fatalf("diagnostic feed count: %v", err)
	}
	if diagnosticEvents != 1 {
		t.Fatalf("expected exactly one diagnostic activity feed entry, got %d", diagnosticEvents)
	}

	var snapshotRows int
	if err := db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM dashboard_daily_snapshots WHERE user_id=$1::uuid AND snapshot_date=CURRENT_DATE`, userID).Scan(&snapshotRows); err != nil {
		t.Fatalf("snapshot existence check: %v", err)
	}
	if snapshotRows != 1 {
		t.Fatalf("expected snapshot row for current date, got %d", snapshotRows)
	}
}

func TestDiagnosticFinalizationCreatesSnapshotIfMissing(t *testing.T) {
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

	userID := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	_, _ = db.Pool().Exec(ctx, `INSERT INTO auth.users(id) VALUES ($1::uuid) ON CONFLICT DO NOTHING`, userID)
	_, _ = db.Pool().Exec(ctx, `INSERT INTO profiles(id, full_name) VALUES ($1::uuid, 'Snapshot User') ON CONFLICT DO NOTHING`, userID)

	repo := NewRepository(db.Pool())
	svc := NewService(repo)
	attemptID, err := svc.StartDiagnostic(ctx, userID, []string{"22222222-2222-2222-2222-222222222221"})
	if err != nil {
		t.Fatalf("start diagnostic: %v", err)
	}

	for i := 0; i < 8; i++ {
		q, err := svc.FetchNextQuestion(ctx, userID, attemptID)
		if err != nil {
			break
		}
		if q.QuestionType == "mcq" {
			opt := 0
			if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "mcq", SelectedOption: &opt}); err != nil {
				t.Fatalf("submit mcq: %v", err)
			}
			continue
		}
		if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "coding", Code: "package main\nfunc solve(){}", Language: "go"}); err != nil {
			t.Fatalf("submit coding: %v", err)
		}
	}

	if _, err := db.Pool().Exec(ctx, `DELETE FROM dashboard_daily_snapshots WHERE user_id=$1::uuid AND snapshot_date=CURRENT_DATE`, userID); err != nil {
		t.Fatalf("delete snapshot row: %v", err)
	}

	if err := svc.SubmitTest(ctx, userID, attemptID); err != nil {
		t.Fatalf("submit test: %v", err)
	}

	var snapshotRows int
	if err := db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM dashboard_daily_snapshots WHERE user_id=$1::uuid AND snapshot_date=CURRENT_DATE`, userID).Scan(&snapshotRows); err != nil {
		t.Fatalf("snapshot existence check: %v", err)
	}
	if snapshotRows != 1 {
		t.Fatalf("expected diagnostic finalization to recreate snapshot row, got %d", snapshotRows)
	}
}

func TestDiagnosticOnlyDayIncrementsStreak(t *testing.T) {
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

	userID := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	_, _ = db.Pool().Exec(ctx, `INSERT INTO auth.users(id) VALUES ($1::uuid) ON CONFLICT DO NOTHING`, userID)
	_, _ = db.Pool().Exec(ctx, `INSERT INTO profiles(id, full_name) VALUES ($1::uuid, 'Diagnostic Streak User') ON CONFLICT DO NOTHING`, userID)

	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO dashboard_daily_snapshots(user_id, snapshot_date, streak_count, questions_solved, mastery_score, topics_completed, last_activity_at, computed_at)
		VALUES ($1::uuid, CURRENT_DATE - INTERVAL '1 day', 3, 0, 0, 0, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day')
		ON CONFLICT (user_id, snapshot_date) DO UPDATE SET streak_count=3, questions_solved=0
	`, userID); err != nil {
		t.Fatalf("insert yesterday snapshot: %v", err)
	}
	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO user_activity_feed(id, user_id, source, title, topic_id, solved_at)
		VALUES (gen_random_uuid(), $1::uuid, 'diagnostic', 'Yesterday Diagnostic', NULL, (CURRENT_DATE - INTERVAL '1 day')::timestamp + INTERVAL '12 hour')
	`, userID); err != nil {
		t.Fatalf("insert yesterday diagnostic activity: %v", err)
	}

	repo := NewRepository(db.Pool())
	svc := NewService(repo)
	attemptID, err := svc.StartDiagnostic(ctx, userID, []string{"22222222-2222-2222-2222-222222222221"})
	if err != nil {
		t.Fatalf("start diagnostic: %v", err)
	}

	for i := 0; i < 8; i++ {
		q, err := svc.FetchNextQuestion(ctx, userID, attemptID)
		if err != nil {
			break
		}
		if q.QuestionType == "mcq" {
			opt := 0
			if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "mcq", SelectedOption: &opt}); err != nil {
				t.Fatalf("submit mcq: %v", err)
			}
			continue
		}
		if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "coding", Code: "package main\nfunc solve(){}", Language: "go"}); err != nil {
			t.Fatalf("submit coding: %v", err)
		}
	}

	if err := svc.SubmitTest(ctx, userID, attemptID); err != nil {
		t.Fatalf("submit test: %v", err)
	}

	var streak int
	if err := db.Pool().QueryRow(ctx, `SELECT streak_count FROM dashboard_daily_snapshots WHERE user_id=$1::uuid AND snapshot_date=CURRENT_DATE`, userID).Scan(&streak); err != nil {
		t.Fatalf("read streak: %v", err)
	}
	if streak != 4 {
		t.Fatalf("expected streak to increment from diagnostic-only activity day, got %d", streak)
	}
}

func TestDiagnosticOverrideReplacesExistingMastery(t *testing.T) {
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

	userID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	topicID := "22222222-2222-2222-2222-222222222221"
	_, _ = db.Pool().Exec(ctx, `INSERT INTO auth.users(id) VALUES ($1::uuid) ON CONFLICT DO NOTHING`, userID)
	_, _ = db.Pool().Exec(ctx, `INSERT INTO profiles(id, full_name) VALUES ($1::uuid, 'Override User') ON CONFLICT DO NOTHING`, userID)
	_, _ = db.Pool().Exec(ctx, `
		INSERT INTO user_topic_progress(user_id, topic_id, status, mastery_score, completed_at, external_solved_count, total_external_questions)
		VALUES ($1::uuid,$2::uuid,'completed',95,NOW(),9,10)
		ON CONFLICT (user_id, topic_id)
		DO UPDATE SET mastery_score=95,status='completed',completed_at=NOW(),external_solved_count=9,total_external_questions=10
	`, userID, topicID)

	repo := NewRepository(db.Pool())
	svc := NewService(repo)
	attemptID, err := svc.StartDiagnostic(ctx, userID, []string{topicID})
	if err != nil {
		t.Fatalf("start diagnostic: %v", err)
	}

	for i := 0; i < 10; i++ {
		q, err := svc.FetchNextQuestion(ctx, userID, attemptID)
		if err != nil {
			break
		}
		switch q.QuestionType {
		case "mcq":
			opt := 1 // intentionally wrong for seeded questions
			if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "mcq", SelectedOption: &opt}); err != nil {
				t.Fatalf("submit answer: %v", err)
			}
		case "coding":
			if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "coding", Code: "", Language: "go"}); err == nil {
				t.Fatalf("expected validation error for empty code")
			}
			if _, err := svc.SubmitAnswer(ctx, userID, attemptID, AnswerData{QuestionID: q.QuestionID, QuestionType: "coding", Code: "package main\nfunc solve(){}", Language: "go"}); err != nil {
				t.Fatalf("submit coding: %v", err)
			}
		}
	}

	if err := svc.SubmitTest(ctx, userID, attemptID); err != nil {
		t.Fatalf("submit test: %v", err)
	}

	var mastery float64
	if err := db.Pool().QueryRow(ctx, `SELECT mastery_score FROM user_topic_progress WHERE user_id=$1::uuid AND topic_id=$2::uuid`, userID, topicID).Scan(&mastery); err != nil {
		t.Fatalf("read mastery: %v", err)
	}
	if mastery >= 95 {
		t.Fatalf("expected diagnostic override to replace previous mastery, got %.2f", mastery)
	}
}
