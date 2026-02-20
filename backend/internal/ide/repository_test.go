package ide

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/harshit110927/arnocodes/backend/internal/database"
	"github.com/jackc/pgx/v5"
)

type failingMasteryUpdater struct{}

func (f failingMasteryUpdater) SaveLearningQuestionCompletionTx(ctx context.Context, tx pgx.Tx, userID, questionID string) error {
	_ = ctx
	_ = tx
	_ = userID
	_ = questionID
	return errors.New("mastery failure")
}

type countingMasteryUpdater struct{ count int }

func (c *countingMasteryUpdater) SaveLearningQuestionCompletionTx(ctx context.Context, tx pgx.Tx, userID, questionID string) error {
	_ = ctx
	_ = tx
	_ = userID
	_ = questionID
	c.count++
	return nil
}

func setupIDERepo(t *testing.T) (*database.DB, *Repository, string) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := database.New(ctx, dsn)
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	if err := database.RunMigrations(ctx, db); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	if err := database.RunSeed(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	_, _ = db.Pool().Exec(ctx, `INSERT INTO auth.users(id) VALUES ($1::uuid) ON CONFLICT DO NOTHING`, userID)
	_, _ = db.Pool().Exec(ctx, `INSERT INTO profiles(id, full_name) VALUES ($1::uuid,'IDE User') ON CONFLICT DO NOTHING`, userID)
	return db, NewRepository(db.Pool()), userID
}

func TestFinalizeSubmissionWithMasteryRollbackOnMasteryFailure(t *testing.T) {
	db, repo, userID := setupIDERepo(t)
	defer db.Close()
	ctx := context.Background()

	var submissionID string
	err := db.Pool().QueryRow(ctx, `
		INSERT INTO coding_submissions(id, attempt_id, question_id, user_id, code, language, evaluation_status, score, created_at, evaluated_at)
		VALUES (gen_random_uuid(), NULL, '55555555-5555-5555-5555-555555555554'::uuid, $1::uuid, 'print(1)', 'python', 'processing', NULL, NOW(), NULL)
		RETURNING id::text
	`, userID).Scan(&submissionID)
	if err != nil {
		t.Fatal(err)
	}

	score := 95.0
	err = repo.FinalizeSubmissionWithMastery(ctx, submissionID, "completed", &score, PassThreshold, failingMasteryUpdater{})
	if err == nil {
		t.Fatalf("expected mastery failure")
	}

	var status string
	var evaluatedAt *time.Time
	if err := db.Pool().QueryRow(ctx, `SELECT evaluation_status, evaluated_at FROM coding_submissions WHERE id=$1::uuid`, submissionID).Scan(&status, &evaluatedAt); err != nil {
		t.Fatal(err)
	}
	if status != "processing" || evaluatedAt != nil {
		t.Fatalf("expected rollback to keep processing/unevaluated, got status=%s evaluated_at=%v", status, evaluatedAt)
	}
}

func TestFinalizeSubmissionWithMasteryOnlyLatestTriggers(t *testing.T) {
	db, repo, userID := setupIDERepo(t)
	defer db.Close()
	ctx := context.Background()

	var olderID, newerID string
	err := db.Pool().QueryRow(ctx, `
		INSERT INTO coding_submissions(id, attempt_id, question_id, user_id, code, language, evaluation_status, score, created_at, evaluated_at)
		VALUES (gen_random_uuid(), NULL, '55555555-5555-5555-5555-555555555554'::uuid, $1::uuid, 'print(1)', 'python', 'processing', NULL, NOW() - INTERVAL '1 minute', NULL)
		RETURNING id::text
	`, userID).Scan(&olderID)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Pool().QueryRow(ctx, `
		INSERT INTO coding_submissions(id, attempt_id, question_id, user_id, code, language, evaluation_status, score, created_at, evaluated_at)
		VALUES (gen_random_uuid(), NULL, '55555555-5555-5555-5555-555555555554'::uuid, $1::uuid, 'print(2)', 'python', 'processing', NULL, NOW(), NULL)
		RETURNING id::text
	`, userID).Scan(&newerID)
	if err != nil {
		t.Fatal(err)
	}

	updater := &countingMasteryUpdater{}
	score := 100.0
	if err := repo.FinalizeSubmissionWithMastery(ctx, olderID, "completed", &score, PassThreshold, updater); err != nil {
		t.Fatal(err)
	}
	if updater.count != 0 {
		t.Fatalf("expected no mastery call for non-latest submission")
	}

	if err := repo.FinalizeSubmissionWithMastery(ctx, newerID, "completed", &score, PassThreshold, updater); err != nil {
		t.Fatal(err)
	}
	if updater.count != 1 {
		t.Fatalf("expected mastery call for latest submission once, got %d", updater.count)
	}
}

func TestFinalizeSubmissionWithMasteryRowsAffectedGuard(t *testing.T) {
	db, repo, userID := setupIDERepo(t)
	defer db.Close()
	ctx := context.Background()

	var submissionID string
	err := db.Pool().QueryRow(ctx, `
		INSERT INTO coding_submissions(id, attempt_id, question_id, user_id, code, language, evaluation_status, score, created_at, evaluated_at)
		VALUES (gen_random_uuid(), NULL, '55555555-5555-5555-5555-555555555554'::uuid, $1::uuid, 'print(1)', 'python', 'processing', NULL, NOW(), NULL)
		RETURNING id::text
	`, userID).Scan(&submissionID)
	if err != nil {
		t.Fatal(err)
	}

	updater := &countingMasteryUpdater{}
	score := 90.0
	if err := repo.FinalizeSubmissionWithMastery(ctx, submissionID, "completed", &score, PassThreshold, updater); err != nil {
		t.Fatal(err)
	}
	if err := repo.FinalizeSubmissionWithMastery(ctx, submissionID, "completed", &score, PassThreshold, updater); err != nil {
		t.Fatal(err)
	}
	if updater.count != 1 {
		t.Fatalf("expected mastery called once on first finalize only, got %d", updater.count)
	}
}

func TestFinalizeSubmissionWithMasteryDoesNotRunOnFailedStatus(t *testing.T) {
	db, repo, userID := setupIDERepo(t)
	defer db.Close()
	ctx := context.Background()

	var submissionID string
	err := db.Pool().QueryRow(ctx, `
		INSERT INTO coding_submissions(id, attempt_id, question_id, user_id, code, language, evaluation_status, score, created_at, evaluated_at)
		VALUES (gen_random_uuid(), NULL, '55555555-5555-5555-5555-555555555554'::uuid, $1::uuid, 'print(1)', 'python', 'processing', NULL, NOW(), NULL)
		RETURNING id::text
	`, userID).Scan(&submissionID)
	if err != nil {
		t.Fatal(err)
	}

	updater := &countingMasteryUpdater{}
	score := 100.0
	if err := repo.FinalizeSubmissionWithMastery(ctx, submissionID, "failed", &score, PassThreshold, updater); err != nil {
		t.Fatal(err)
	}
	if updater.count != 0 {
		t.Fatalf("expected no mastery calls for failed submissions")
	}
}
