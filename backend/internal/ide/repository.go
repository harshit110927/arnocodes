package ide

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

type MasteryUpdater interface {
	SaveLearningQuestionCompletionTx(ctx context.Context, tx pgx.Tx, userID, questionID string) error
}

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) QuestionExists(ctx context.Context, questionID string) (bool, error) {
	if r.pool == nil {
		return false, fmt.Errorf("ide repository is not initialized")
	}
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM questions WHERE id=$1::uuid)`, questionID).Scan(&exists)
	return exists, err
}

func (r *Repository) CreateSubmission(ctx context.Context, s Submission) (string, error) {
	if r.pool == nil {
		return "", fmt.Errorf("ide repository is not initialized")
	}
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO coding_submissions (id, attempt_id, question_id, user_id, code, language, evaluation_status, score, created_at, evaluated_at)
		VALUES (gen_random_uuid(), $1::uuid, $2::uuid, $3::uuid, $4, $5, 'pending', NULL, NOW(), NULL)
		RETURNING id::text
	`, s.AttemptID, s.QuestionID, s.UserID, s.Code, s.Language).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *Repository) GetSubmissionStatus(ctx context.Context, submissionID, userID string) (SubmissionStatus, error) {
	if r.pool == nil {
		return SubmissionStatus{}, fmt.Errorf("ide repository is not initialized")
	}
	var out SubmissionStatus
	err := r.pool.QueryRow(ctx, `
		SELECT evaluation_status, score, evaluated_at
		FROM coding_submissions
		WHERE id=$1::uuid AND user_id=$2::uuid
	`, submissionID, userID).Scan(&out.EvaluationStatus, &out.Score, &out.EvaluatedAt)
	return out, err
}

func (r *Repository) ListTestCases(ctx context.Context, questionID string, sampleOnly bool) ([]CodingQuestionTestCase, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("ide repository is not initialized")
	}
	q := `
		SELECT id::text, question_id::text, input, expected_output, is_sample, weight, order_index, created_at
		FROM coding_question_test_cases
		WHERE question_id=$1::uuid`
	if sampleOnly {
		q += ` AND is_sample=TRUE`
	}
	q += ` ORDER BY order_index ASC`
	rows, err := r.pool.Query(ctx, q, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CodingQuestionTestCase, 0)
	for rows.Next() {
		var tc CodingQuestionTestCase
		if err := rows.Scan(&tc.ID, &tc.QuestionID, &tc.Input, &tc.ExpectedOutput, &tc.IsSample, &tc.Weight, &tc.OrderIndex, &tc.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, tc)
	}
	return out, rows.Err()
}

func (r *Repository) ClaimNextPendingSubmission(ctx context.Context) (Submission, bool, error) {
	if r.pool == nil {
		return Submission{}, false, fmt.Errorf("ide repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Submission{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var sub Submission
	err = tx.QueryRow(ctx, `
		SELECT id::text, attempt_id::text, question_id::text, user_id::text, code, language, evaluation_status, score, created_at, evaluated_at
		FROM coding_submissions
		WHERE evaluation_status='pending'
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`).Scan(&sub.ID, &sub.AttemptID, &sub.QuestionID, &sub.UserID, &sub.Code, &sub.Language, &sub.EvaluationStatus, &sub.Score, &sub.CreatedAt, &sub.EvaluatedAt)
	if err == pgx.ErrNoRows {
		if err := tx.Commit(ctx); err != nil {
			return Submission{}, false, err
		}
		return Submission{}, false, nil
	}
	if err != nil {
		return Submission{}, false, err
	}

	res, err := tx.Exec(ctx, `
		UPDATE coding_submissions
		SET evaluation_status='processing'
		WHERE id=$1::uuid AND evaluation_status='pending'
	`, sub.ID)
	if err != nil {
		return Submission{}, false, err
	}
	if res.RowsAffected() == 0 {
		return Submission{}, false, fmt.Errorf("failed to claim submission")
	}

	if err := tx.Commit(ctx); err != nil {
		return Submission{}, false, err
	}
	sub.EvaluationStatus = "processing"
	return sub, true, nil
}

func (r *Repository) FinalizeSubmission(ctx context.Context, submissionID, status string, score *float64) error {
	if r.pool == nil {
		return fmt.Errorf("ide repository is not initialized")
	}
	res, err := r.pool.Exec(ctx, `
		UPDATE coding_submissions
		SET evaluation_status=$2,
		    score=$3,
		    evaluated_at=NOW()
		WHERE id=$1::uuid
		  AND evaluation_status='processing'
	`, submissionID, status, score)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return nil
	}
	return nil
}

func (r *Repository) FinalizeSubmissionWithMastery(ctx context.Context, submissionID string, status string, score *float64, passThreshold float64, masteryUpdater MasteryUpdater) error {
	if r.pool == nil {
		return fmt.Errorf("ide repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	res, err := tx.Exec(ctx, `
		UPDATE coding_submissions
		SET evaluation_status=$2,
		    score=$3,
		    evaluated_at=NOW()
		WHERE id=$1::uuid
		  AND evaluation_status='processing'
	`, submissionID, status, score)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}

	var sub Submission
	if err := tx.QueryRow(ctx, `
		SELECT id::text, attempt_id::text, question_id::text, user_id::text, evaluation_status, score
		FROM coding_submissions
		WHERE id=$1::uuid
		FOR UPDATE
	`, submissionID).Scan(&sub.ID, &sub.AttemptID, &sub.QuestionID, &sub.UserID, &sub.EvaluationStatus, &sub.Score); err != nil {
		return err
	}
	if sub.EvaluationStatus != status {
		return tx.Commit(ctx)
	}
	if status != "completed" {
		return tx.Commit(ctx)
	}

	if masteryUpdater != nil && score != nil && *score >= passThreshold && sub.AttemptID == nil {
		var latestID string
		if err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM coding_submissions
			WHERE user_id=$1::uuid
			  AND question_id=$2::uuid
			  AND evaluation_status IN ('processing','completed')
			ORDER BY created_at DESC, id DESC
			LIMIT 1
			FOR UPDATE
		`, sub.UserID, sub.QuestionID).Scan(&latestID); err != nil {
			return err
		}
		if latestID == sub.ID {
			if err := masteryUpdater.SaveLearningQuestionCompletionTx(ctx, tx, sub.UserID, sub.QuestionID); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) IsLatestSubmission(ctx context.Context, submissionID, userID, questionID string) (bool, error) {
	if r.pool == nil {
		return false, fmt.Errorf("ide repository is not initialized")
	}
	var latestID string
	err := r.pool.QueryRow(ctx, `
		SELECT id::text
		FROM coding_submissions
		WHERE user_id=$1::uuid
		  AND question_id=$2::uuid
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, userID, questionID).Scan(&latestID)
	if err != nil {
		return false, err
	}
	return latestID == submissionID, nil
}

func (r *Repository) ResetStuckProcessingSubmissions(ctx context.Context) error {
	if r.pool == nil {
		return fmt.Errorf("ide repository is not initialized")
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE coding_submissions
		SET evaluation_status='pending'
		WHERE evaluation_status='processing'
		  AND evaluated_at IS NULL
		  AND score IS NULL
		  AND NOW() - created_at > INTERVAL '5 minutes'
	`)
	return err
}

func (r *Repository) GetSubmissionByID(ctx context.Context, submissionID string) (Submission, error) {
	if r.pool == nil {
		return Submission{}, fmt.Errorf("ide repository is not initialized")
	}
	var sub Submission
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, attempt_id::text, question_id::text, user_id::text, code, language, evaluation_status, score, created_at, evaluated_at
		FROM coding_submissions
		WHERE id=$1::uuid
	`, submissionID).Scan(&sub.ID, &sub.AttemptID, &sub.QuestionID, &sub.UserID, &sub.Code, &sub.Language, &sub.EvaluationStatus, &sub.Score, &sub.CreatedAt, &sub.EvaluatedAt)
	return sub, err
}

func (r *Repository) InsertFailedSubmission(ctx context.Context, s Submission, failureScore *float64) (string, error) {
	if r.pool == nil {
		return "", fmt.Errorf("ide repository is not initialized")
	}
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO coding_submissions (id, attempt_id, question_id, user_id, code, language, evaluation_status, score, created_at, evaluated_at)
		VALUES (gen_random_uuid(), $1::uuid, $2::uuid, $3::uuid, $4, $5, 'failed', $6, NOW(), NOW())
		RETURNING id::text
	`, s.AttemptID, s.QuestionID, s.UserID, s.Code, s.Language, failureScore).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *Repository) NowUTCDate(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
