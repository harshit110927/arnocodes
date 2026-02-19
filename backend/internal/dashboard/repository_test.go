package dashboard

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/harshit110927/arnocodes/backend/internal/database"
)

func setupDB(t *testing.T) (*database.DB, *Repository, string) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := database.New(ctx, dsn)
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	if err := database.RunMigrations(ctx, db); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	if err := database.RunSeed(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	_, _ = db.Pool().Exec(ctx, `INSERT INTO auth.users(id) VALUES ($1::uuid) ON CONFLICT DO NOTHING`, userID)
	_, _ = db.Pool().Exec(ctx, `INSERT INTO profiles(id, full_name) VALUES ($1::uuid,'Dash User') ON CONFLICT DO NOTHING`, userID)
	return db, NewRepository(db.Pool()), userID
}

func createTestUser(t *testing.T, db *database.DB, label string) string {
	t.Helper()
	userID := fmt.Sprintf("%08x-1111-2222-3333-444444444444", time.Now().UnixNano()&0xffffffff)
	ctx := context.Background()
	if _, err := db.Pool().Exec(ctx, `INSERT INTO auth.users(id) VALUES ($1::uuid) ON CONFLICT DO NOTHING`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool().Exec(ctx, `INSERT INTO profiles(id, full_name) VALUES ($1::uuid,$2) ON CONFLICT DO NOTHING`, userID, label); err != nil {
		t.Fatal(err)
	}
	return userID
}

func TestExternalIdempotentInsertion(t *testing.T) {
	db, repo, userID := setupDB(t)
	defer db.Close()
	ctx := context.Background()
	topic := "22222222-2222-2222-2222-222222222221"
	in := ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "two-sum", Title: "Two Sum", TopicID: &topic, SolvedAt: time.Now(), TotalExternalQuestions: 10}
	if err := repo.RecordExternalSolve(ctx, in); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordExternalSolve(ctx, in); err != nil {
		t.Fatal(err)
	}
	var c int
	if err := db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM external_question_activity WHERE user_id=$1::uuid AND platform='leetcode' AND platform_question_id='two-sum'`, userID).Scan(&c); err != nil {
		t.Fatal(err)
	}
	if c != 1 {
		t.Fatalf("expected 1 activity row got %d", c)
	}
}

func TestMasteryAndUnlockAt80(t *testing.T) {
	db, repo, userID := setupDB(t)
	defer db.Close()
	ctx := context.Background()
	arrays := "22222222-2222-2222-2222-222222222221"
	strings := "22222222-2222-2222-2222-222222222222"
	_, _ = db.Pool().Exec(ctx, `INSERT INTO topic_prerequisites(topic_id, prerequisite_id) VALUES ($1::uuid,$2::uuid) ON CONFLICT DO NOTHING`, strings, arrays)
	for i := 0; i < 8; i++ {
		qid := "arr-q-" + time.Now().Add(time.Duration(i)*time.Nanosecond).Format("150405.000000000")
		err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: qid, Title: "A", TopicID: &arrays, SolvedAt: time.Now(), TotalExternalQuestions: 10})
		if err != nil {
			t.Fatal(err)
		}
	}
	var mastery float64
	var status string
	if err := db.Pool().QueryRow(ctx, `SELECT mastery_score, status::text FROM user_topic_progress WHERE user_id=$1::uuid AND topic_id=$2::uuid`, userID, arrays).Scan(&mastery, &status); err != nil {
		t.Fatal(err)
	}
	if mastery < 80 || status != "completed" {
		t.Fatalf("expected completed at >=80, got %.2f %s", mastery, status)
	}
	var unlocked int
	if err := db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM user_topic_progress WHERE user_id=$1::uuid AND topic_id=$2::uuid`, userID, strings).Scan(&unlocked); err != nil {
		t.Fatal(err)
	}
	if unlocked == 0 {
		t.Fatalf("expected dependent topic unlocked")
	}
}

func TestSnapshotIncrementOnceOnTransition(t *testing.T) {
	db, repo, userID := setupDB(t)
	defer db.Close()
	ctx := context.Background()
	topic := "22222222-2222-2222-2222-222222222223"
	for i := 0; i < 9; i++ {
		qid := "tree-q-" + time.Now().Add(time.Duration(i)*time.Nanosecond).Format("150405.000000000")
		if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "cf", PlatformQuestionID: qid, Title: "T", TopicID: &topic, SolvedAt: time.Now(), TotalExternalQuestions: 10}); err != nil {
			t.Fatal(err)
		}
	}
	var topicsCompleted int
	if err := db.Pool().QueryRow(ctx, `SELECT topics_completed FROM dashboard_daily_snapshots WHERE user_id=$1::uuid ORDER BY snapshot_date DESC LIMIT 1`, userID).Scan(&topicsCompleted); err != nil {
		t.Fatal(err)
	}
	if topicsCompleted < 1 {
		t.Fatalf("expected topics_completed increment")
	}
}

func TestConcurrencyParallelExternalSolves(t *testing.T) {
	db, repo, userID := setupDB(t)
	defer db.Close()
	ctx := context.Background()
	topic := "22222222-2222-2222-2222-222222222221"
	in := ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "parallel-q", Title: "P", TopicID: &topic, SolvedAt: time.Now(), TotalExternalQuestions: 10}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = repo.RecordExternalSolve(ctx, in)
		}()
	}
	wg.Wait()
	var c int
	if err := db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM external_question_activity WHERE user_id=$1::uuid AND platform='leetcode' AND platform_question_id='parallel-q'`, userID).Scan(&c); err != nil {
		t.Fatal(err)
	}
	if c != 1 {
		t.Fatalf("expected 1 row after parallel insert, got %d", c)
	}
}

func TestDivisionByZeroGuard(t *testing.T) {
	db, repo, userID := setupDB(t)
	defer db.Close()
	ctx := context.Background()
	topic := "22222222-2222-2222-2222-222222222224"

	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO user_topic_progress (user_id, topic_id, status, mastery_score, diagnostic_mastery, external_solved_count, total_external_questions)
		VALUES ($1::uuid, $2::uuid, 'in_progress', 0, 0, 0, 0)
		ON CONFLICT (user_id, topic_id) DO UPDATE SET status='in_progress'::learning_progress_status, mastery_score=0, diagnostic_mastery=0, external_solved_count=0, total_external_questions=0
	`, userID, topic); err != nil {
		t.Fatal(err)
	}

	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "div-zero-q", Title: "DZ", TopicID: &topic, SolvedAt: time.Now(), TotalExternalQuestions: 0}); err != nil {
		t.Fatal(err)
	}

	var mastery float64
	if err := db.Pool().QueryRow(ctx, `SELECT mastery_score FROM user_topic_progress WHERE user_id=$1::uuid AND topic_id=$2::uuid`, userID, topic).Scan(&mastery); err != nil {
		t.Fatal(err)
	}
	if mastery != 0 {
		t.Fatalf("expected 0 mastery when total_external_questions=0, got %.2f", mastery)
	}
}

func TestExternalMasteryCannotLowerDiagnosticMastery(t *testing.T) {
	db, repo, userID := setupDB(t)
	defer db.Close()
	ctx := context.Background()
	topic := "22222222-2222-2222-2222-222222222225"

	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO user_topic_progress (user_id, topic_id, status, mastery_score, diagnostic_mastery, external_solved_count, total_external_questions)
		VALUES ($1::uuid, $2::uuid, 'completed', 90, 90, 0, 20)
		ON CONFLICT (user_id, topic_id) DO UPDATE SET status='completed'::learning_progress_status, mastery_score=90, diagnostic_mastery=90, external_solved_count=0, total_external_questions=20
	`, userID, topic); err != nil {
		t.Fatal(err)
	}

	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "diag-floor-q", Title: "DF", TopicID: &topic, SolvedAt: time.Now(), TotalExternalQuestions: 20}); err != nil {
		t.Fatal(err)
	}

	var mastery, diagnostic float64
	if err := db.Pool().QueryRow(ctx, `SELECT mastery_score, diagnostic_mastery FROM user_topic_progress WHERE user_id=$1::uuid AND topic_id=$2::uuid`, userID, topic).Scan(&mastery, &diagnostic); err != nil {
		t.Fatal(err)
	}
	if diagnostic != 90 || mastery < diagnostic {
		t.Fatalf("expected mastery to keep diagnostic floor; mastery=%.2f diagnostic=%.2f", mastery, diagnostic)
	}
}

func TestStreakIncrementsConsecutiveDaysAndResetsAfterGap(t *testing.T) {
	db, repo, _ := setupDB(t)
	defer db.Close()
	userID := createTestUser(t, db, "Streak User")
	ctx := context.Background()
	topic := "22222222-2222-2222-2222-222222222221"

	day1 := time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC)
	day2 := day1.Add(24 * time.Hour)
	day4 := day1.Add(72 * time.Hour)

	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "streak-d1", Title: "S1", TopicID: &topic, SolvedAt: day1, TotalExternalQuestions: 10}); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "streak-d2", Title: "S2", TopicID: &topic, SolvedAt: day2, TotalExternalQuestions: 10}); err != nil {
		t.Fatal(err)
	}

	var streakDay2 int
	if err := db.Pool().QueryRow(ctx, `SELECT streak_count FROM dashboard_daily_snapshots WHERE user_id=$1::uuid AND snapshot_date=$2::date`, userID, day2).Scan(&streakDay2); err != nil {
		t.Fatal(err)
	}
	if streakDay2 != 2 {
		t.Fatalf("expected streak 2 on consecutive day, got %d", streakDay2)
	}

	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "streak-d4", Title: "S4", TopicID: &topic, SolvedAt: day4, TotalExternalQuestions: 10}); err != nil {
		t.Fatal(err)
	}

	var streakDay4 int
	if err := db.Pool().QueryRow(ctx, `SELECT streak_count FROM dashboard_daily_snapshots WHERE user_id=$1::uuid AND snapshot_date=$2::date`, userID, day4).Scan(&streakDay4); err != nil {
		t.Fatal(err)
	}
	if streakDay4 != 1 {
		t.Fatalf("expected streak reset to 1 after gap, got %d", streakDay4)
	}
}

func TestMidnightBoundaryDateNormalization(t *testing.T) {
	db, repo, _ := setupDB(t)
	defer db.Close()
	userID := createTestUser(t, db, "Boundary User")
	ctx := context.Background()
	topic := "22222222-2222-2222-2222-222222222221"

	first := time.Date(2026, 1, 10, 23, 59, 59, 0, time.UTC)
	second := time.Date(2026, 1, 11, 0, 0, 1, 0, time.UTC)

	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "midnight-1", Title: "M1", TopicID: &topic, SolvedAt: first, TotalExternalQuestions: 10}); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "midnight-2", Title: "M2", TopicID: &topic, SolvedAt: second, TotalExternalQuestions: 10}); err != nil {
		t.Fatal(err)
	}

	var rowCount int
	if err := db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM dashboard_daily_snapshots WHERE user_id=$1::uuid AND snapshot_date IN ($2::date,$3::date)`, userID, first, second).Scan(&rowCount); err != nil {
		t.Fatal(err)
	}
	if rowCount != 2 {
		t.Fatalf("expected two snapshot dates across midnight boundary, got %d", rowCount)
	}
}

func TestCompletionDoesNotOverwriteStreak(t *testing.T) {
	db, repo, _ := setupDB(t)
	defer db.Close()
	userID := createTestUser(t, db, "Completion Streak User")
	ctx := context.Background()
	topic := "22222222-2222-2222-2222-222222222221"

	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO user_topic_progress (user_id, topic_id, status, mastery_score, diagnostic_mastery, external_solved_count, total_external_questions)
		VALUES ($1::uuid, $2::uuid, 'in_progress', 0, 0, 0, 1)
		ON CONFLICT (user_id, topic_id) DO UPDATE SET status='in_progress'::learning_progress_status, mastery_score=0, diagnostic_mastery=0, external_solved_count=0, total_external_questions=1
	`, userID, topic); err != nil {
		t.Fatal(err)
	}

	yesterday := time.Date(2026, 1, 20, 9, 0, 0, 0, time.UTC)
	today := yesterday.Add(24 * time.Hour)

	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "keep-streak-y", Title: "Y", TopicID: &topic, SolvedAt: yesterday, TotalExternalQuestions: 10}); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Pool().Exec(ctx, `
		UPDATE user_topic_progress
		SET status='in_progress'::learning_progress_status, mastery_score=0, completed_at=NULL, external_solved_count=0, total_external_questions=1
		WHERE user_id=$1::uuid AND topic_id=$2::uuid
	`, userID, topic); err != nil {
		t.Fatal(err)
	}

	if err := repo.RecordExternalSolve(ctx, ExternalSolveInput{UserID: userID, Platform: "leetcode", PlatformQuestionID: "keep-streak-t", Title: "T", TopicID: &topic, SolvedAt: today, TotalExternalQuestions: 1}); err != nil {
		t.Fatal(err)
	}

	var streak int
	if err := db.Pool().QueryRow(ctx, `SELECT streak_count FROM dashboard_daily_snapshots WHERE user_id=$1::uuid AND snapshot_date=$2::date`, userID, today).Scan(&streak); err != nil {
		t.Fatal(err)
	}
	if streak != 2 {
		t.Fatalf("expected streak to remain 2 after completion update, got %d", streak)
	}
}
