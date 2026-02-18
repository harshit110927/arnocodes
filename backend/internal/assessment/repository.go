package assessment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type UserStatus struct {
	DiagnosticCompleted bool `json:"diagnostic_completed"`
	DashboardUnlocked   bool `json:"dashboard_unlocked"`
}

type DiagnosticQuestion struct {
	QuestionID   string          `json:"question_id"`
	QuestionType string          `json:"question_type"`
	Content      string          `json:"content"`
	Options      json.RawMessage `json:"options,omitempty"`
	OrderIndex   int             `json:"order_index"`
	TopicID      string          `json:"topic_id"`
	AllottedSecs int             `json:"allotted_seconds"`
}

type CodingSubmission struct {
	ID               string
	AttemptID        string
	QuestionID       string
	UserID           string
	Code             string
	Language         string
	EvaluationStatus string
	Score            *float64
}

type DiagnosticAttemptStatus struct {
	AttemptID            string     `json:"attempt_id"`
	UserID               string     `json:"user_id"`
	Status               string     `json:"status"`
	StartedAt            time.Time  `json:"started_at"`
	ExpiresAt            time.Time  `json:"expires_at"`
	SubmittedAt          *time.Time `json:"submitted_at,omitempty"`
	LastAnsweredOrderIdx int        `json:"last_answered_order_index"`
	TotalQuestions       int        `json:"total_questions"`
	AnsweredQuestions    int        `json:"answered_questions"`
	TotalAllowedSeconds  int        `json:"total_allowed_seconds"`
}

func (r *Repository) ValidateTopicSelection(userSelectedTopics []string) error {
	if len(userSelectedTopics) == 0 {
		return fmt.Errorf("at least one topic must be selected")
	}
	if r.pool == nil {
		return fmt.Errorf("assessment repository is not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	set := map[string]bool{}
	for _, t := range userSelectedTopics {
		trimmed := strings.TrimSpace(t)
		if trimmed == "" {
			continue
		}
		set[trimmed] = true
	}
	if len(set) == 0 {
		return fmt.Errorf("at least one non-empty topic id must be selected")
	}

	for topicID := range set {
		rows, err := r.pool.Query(ctx, `SELECT prerequisite_id::text FROM topic_prerequisites WHERE topic_id=$1::uuid`, topicID)
		if err != nil {
			return fmt.Errorf("query prerequisites: %w", err)
		}
		for rows.Next() {
			var prerequisiteID string
			if err := rows.Scan(&prerequisiteID); err != nil {
				rows.Close()
				return fmt.Errorf("scan prerequisite: %w", err)
			}
			if !set[prerequisiteID] {
				rows.Close()
				return fmt.Errorf("topic %s requires prerequisite %s", topicID, prerequisiteID)
			}
		}
		rows.Close()
	}
	return nil
}

func (r *Repository) CanStartDiagnostic(userID string) (bool, error) {
	if r.pool == nil {
		return false, fmt.Errorf("assessment repository is not initialized")
	}
	const q = `
	SELECT COUNT(*)
	FROM test_attempts ta
	JOIN tests t ON t.id = ta.test_id
	WHERE ta.user_id = $1::uuid
	  AND t.type = 'diagnostic'
	  AND ta.started_at >= NOW() - INTERVAL '48 hours'
	`
	var count int
	if err := r.pool.QueryRow(context.Background(), q, userID).Scan(&count); err != nil {
		return false, fmt.Errorf("query recent diagnostic attempts: %w", err)
	}
	return count < 2, nil
}

func (r *Repository) CreateDiagnosticAttempt(ctx context.Context, userID string, selectedTopics []string) (string, error) {
	if r.pool == nil {
		return "", fmt.Errorf("assessment repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var diagnosticTestID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM tests WHERE type='diagnostic' ORDER BY id LIMIT 1`).Scan(&diagnosticTestID); err != nil {
		return "", fmt.Errorf("load diagnostic test: %w", err)
	}

	var attemptID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO test_attempts (id, user_id, test_id, score, status, started_at, expires_at, submitted_at, evaluation_version)
		VALUES (gen_random_uuid(), $1::uuid, $2::uuid, 0, 'in_progress', NOW(), NOW() + INTERVAL '2 hours', NULL, 'v1')
		RETURNING id::text
	`, userID, diagnosticTestID).Scan(&attemptID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if err := tx.QueryRow(ctx, `
				SELECT id::text
				FROM test_attempts
				WHERE user_id = $1::uuid
				  AND test_id = $2::uuid
				  AND status IN ('started','in_progress')
				LIMIT 1
			`, userID, diagnosticTestID).Scan(&attemptID); err != nil {
				return "", fmt.Errorf("load existing active attempt: %w", err)
			}
			return attemptID, nil
		}
		return "", fmt.Errorf("create test attempt: %w", err)
	}

	rows, err := tx.Query(ctx, `SELECT id::text, question_type FROM questions WHERE test_id=$1::uuid ORDER BY order_index`, diagnosticTestID)
	if err != nil {
		return "", fmt.Errorf("load diagnostic questions: %w", err)
	}
	questions := make([]struct{ ID, Type string }, 0)
	for rows.Next() {
		var id, qt string
		if err := rows.Scan(&id, &qt); err != nil {
			rows.Close()
			return "", err
		}
		if qt == "slide" {
			continue
		}
		questions = append(questions, struct{ ID, Type string }{ID: id, Type: qt})
	}
	rows.Close()
	if len(questions) == 0 {
		return "", fmt.Errorf("diagnostic test has no answerable questions")
	}
	if len(selectedTopics) == 0 {
		return "", fmt.Errorf("selected topics required")
	}

	for i, q := range questions {
		topicID := selectedTopics[i%len(selectedTopics)]
		allotted := 30
		if q.Type == "coding" {
			allotted = 1800
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO diagnostic_attempt_questions (attempt_id, question_id, topic_id, order_index, question_type, allotted_seconds)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6)
		`, attemptID, q.ID, topicID, i+1, q.Type, allotted); err != nil {
			return "", fmt.Errorf("insert attempt question map: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}
	return attemptID, nil
}

func validateAttemptLocked(ctx context.Context, tx pgx.Tx, attemptID string) error {
	var status string
	var expiresAt time.Time
	if err := tx.QueryRow(ctx, `SELECT status::text, expires_at FROM test_attempts WHERE id=$1::uuid FOR UPDATE`, attemptID).Scan(&status, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("lock attempt: %w", err)
	}
	if status != "in_progress" {
		return fmt.Errorf("attempt is not in progress")
	}
	if time.Now().After(expiresAt) {
		res, err := tx.Exec(ctx, `
			UPDATE test_attempts
			SET status='expired'
			WHERE id=$1::uuid
			  AND status='in_progress'
		`, attemptID)
		if err != nil {
			return fmt.Errorf("mark attempt expired: %w", err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("attempt expired")
		}
		return fmt.Errorf("attempt expired")
	}
	return nil
}

func (r *Repository) GetNextDiagnosticQuestion(ctx context.Context, attemptID string, lastOrderIndex int) (DiagnosticQuestion, error) {
	if r.pool == nil {
		return DiagnosticQuestion{}, fmt.Errorf("assessment repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DiagnosticQuestion{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := validateAttemptLocked(ctx, tx, attemptID); err != nil {
		return DiagnosticQuestion{}, err
	}

	const q = `
	SELECT daq.question_id::text, daq.question_type, qs.content, qs.options, daq.order_index, daq.topic_id::text, daq.allotted_seconds
	FROM diagnostic_attempt_questions daq
	JOIN questions qs ON qs.id = daq.question_id
	WHERE daq.attempt_id = $1::uuid
	  AND daq.order_index > $2
	  AND daq.answered_at IS NULL
	ORDER BY daq.order_index ASC
	LIMIT 1
	`
	var out DiagnosticQuestion
	if err := tx.QueryRow(ctx, q, attemptID, lastOrderIndex).Scan(&out.QuestionID, &out.QuestionType, &out.Content, &out.Options, &out.OrderIndex, &out.TopicID, &out.AllottedSecs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DiagnosticQuestion{}, ErrNotFound
		}
		return DiagnosticQuestion{}, fmt.Errorf("next question: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return DiagnosticQuestion{}, fmt.Errorf("commit tx: %w", err)
	}
	return out, nil
}

func (r *Repository) SubmitMCQAnswer(ctx context.Context, attemptID, questionID string, selectedOption int) error {
	if r.pool == nil {
		return fmt.Errorf("assessment repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := validateAttemptLocked(ctx, tx, attemptID); err != nil {
		return err
	}

	var correctOption int
	if err := tx.QueryRow(ctx, `SELECT correct_option FROM questions WHERE id=$1::uuid`, questionID).Scan(&correctOption); err != nil {
		return fmt.Errorf("load correct option: %w", err)
	}
	isCorrect := selectedOption == correctOption
	if _, err := tx.Exec(ctx, `
		INSERT INTO question_attempts (attempt_id, question_id, selected_option, time_taken_seconds, is_correct, state, answered_at, is_marked_for_review)
		VALUES ($1::uuid,$2::uuid,$3,0,$4,'answered',NOW(),FALSE)
		ON CONFLICT (attempt_id, question_id)
		DO UPDATE SET selected_option=EXCLUDED.selected_option,is_correct=EXCLUDED.is_correct,state='answered',answered_at=NOW()
	`, attemptID, questionID, selectedOption, isCorrect); err != nil {
		return fmt.Errorf("upsert question attempt: %w", err)
	}

	res, err := tx.Exec(ctx, `
		UPDATE diagnostic_attempt_questions
		SET answered_at=NOW()
		WHERE attempt_id=$1::uuid AND question_id=$2::uuid AND answered_at IS NULL
	`, attemptID, questionID)
	if err != nil {
		return fmt.Errorf("mark answered: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("question already answered or not part of attempt")
	}

	return tx.Commit(ctx)
}

func (r *Repository) SaveCodingSubmission(ctx context.Context, attemptID, questionID, userID, code, language string) (string, error) {
	if r.pool == nil {
		return "", fmt.Errorf("assessment repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := validateAttemptLocked(ctx, tx, attemptID); err != nil {
		return "", err
	}

	var submissionID string
	err = tx.QueryRow(ctx, `
		INSERT INTO coding_submissions (id, attempt_id, question_id, user_id, code, language, evaluation_status, score, created_at, evaluated_at)
		VALUES (gen_random_uuid(),$1::uuid,$2::uuid,$3::uuid,$4,$5,'pending',NULL,NOW(),NULL)
		RETURNING id::text
	`, attemptID, questionID, userID, code, language).Scan(&submissionID)
	if err != nil {
		return "", fmt.Errorf("insert coding submission: %w", err)
	}

	res, err := tx.Exec(ctx, `
		UPDATE diagnostic_attempt_questions
		SET answered_at=NOW()
		WHERE attempt_id=$1::uuid AND question_id=$2::uuid AND answered_at IS NULL
	`, attemptID, questionID)
	if err != nil {
		return "", fmt.Errorf("mark coding answered: %w", err)
	}
	if res.RowsAffected() == 0 {
		return "", fmt.Errorf("question already answered or not part of attempt")
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}
	return submissionID, nil
}

func (r *Repository) GetPendingCodingSubmissions(ctx context.Context, limit int) ([]CodingSubmission, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("assessment repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id::text
		FROM coding_submissions
		WHERE evaluation_status='pending'
		ORDER BY created_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("select pending coding submissions: %w", err)
	}

	ids := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan submission id: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit tx: %w", err)
		}
		return []CodingSubmission{}, nil
	}

	res, err := tx.Exec(ctx, `
		UPDATE coding_submissions
		SET evaluation_status='processing'
		WHERE id = ANY($1::uuid[])
		  AND evaluation_status='pending'
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("mark submissions processing: %w", err)
	}
	if res.RowsAffected() != int64(len(ids)) {
		return nil, fmt.Errorf("failed to claim all pending submissions")
	}

	dataRows, err := tx.Query(ctx, `
		SELECT id::text, attempt_id::text, question_id::text, user_id::text, code, language, evaluation_status, score
		FROM coding_submissions
		WHERE id = ANY($1::uuid[])
		ORDER BY created_at ASC
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("load claimed submissions: %w", err)
	}
	defer dataRows.Close()

	out := make([]CodingSubmission, 0, len(ids))
	for dataRows.Next() {
		var c CodingSubmission
		if err := dataRows.Scan(&c.ID, &c.AttemptID, &c.QuestionID, &c.UserID, &c.Code, &c.Language, &c.EvaluationStatus, &c.Score); err != nil {
			return nil, err
		}
		out = append(out, c)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return out, nil
}

func (r *Repository) UpdateCodingSubmissionResult(ctx context.Context, submissionID, status string, score float64) error {
	if r.pool == nil {
		return fmt.Errorf("assessment repository is not initialized")
	}
	res, err := r.pool.Exec(ctx, `
		UPDATE coding_submissions
		SET evaluation_status=$2, score=$3, evaluated_at=NOW()
		WHERE id=$1::uuid AND evaluation_status='processing'
	`, submissionID, status, score)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no processing submission updated")
	}
	return nil
}

func (r *Repository) CompleteDiagnosticAttempt(ctx context.Context, attemptID string) error {
	if r.pool == nil {
		return fmt.Errorf("assessment repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	var status string
	var expiresAt time.Time
	if err := tx.QueryRow(ctx, `SELECT user_id::text, status::text, expires_at FROM test_attempts WHERE id=$1::uuid FOR UPDATE`, attemptID).Scan(&userID, &status, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("lock attempt: %w", err)
	}
	if status == "submitted" {
		return tx.Commit(ctx)
	}
	if status != "in_progress" {
		return fmt.Errorf("attempt is not in progress")
	}
	if time.Now().After(expiresAt) {
		res, updErr := tx.Exec(ctx, `
			UPDATE test_attempts
			SET status='expired'
			WHERE id=$1::uuid
			  AND status='in_progress'
		`, attemptID)
		if updErr != nil {
			return fmt.Errorf("mark attempt expired: %w", updErr)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("attempt expired")
		}
		return fmt.Errorf("attempt expired")
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO diagnostic_topic_results (attempt_id, topic_id, score, max_score, percentage, passed, created_at)
		SELECT daq.attempt_id,
		       daq.topic_id,
		       COALESCE(SUM(CASE WHEN qa.is_correct THEN q.marks ELSE 0 END),0)
		         + COALESCE(SUM(CASE WHEN cs.evaluation_status='completed' THEN COALESCE(cs.score,0)::int ELSE 0 END),0) AS score,
		       GREATEST(COALESCE(SUM(q.marks),0),1) AS max_score,
		       ( (
		          COALESCE(SUM(CASE WHEN qa.is_correct THEN q.marks ELSE 0 END),0)
		          + COALESCE(SUM(CASE WHEN cs.evaluation_status='completed' THEN COALESCE(cs.score,0)::int ELSE 0 END),0)
		         )::float / GREATEST(COALESCE(SUM(q.marks),0),1)::float )*100.0 AS percentage,
		       (((
		          COALESCE(SUM(CASE WHEN qa.is_correct THEN q.marks ELSE 0 END),0)
		          + COALESCE(SUM(CASE WHEN cs.evaluation_status='completed' THEN COALESCE(cs.score,0)::int ELSE 0 END),0)
		         )::float / GREATEST(COALESCE(SUM(q.marks),0),1)::float )*100.0) >= 80.0 AS passed,
		       NOW()
		FROM diagnostic_attempt_questions daq
		JOIN questions q ON q.id = daq.question_id
		LEFT JOIN question_attempts qa ON qa.attempt_id = daq.attempt_id AND qa.question_id = daq.question_id
		LEFT JOIN LATERAL (
			SELECT score, evaluation_status
			FROM coding_submissions
			WHERE attempt_id = daq.attempt_id
			  AND question_id = daq.question_id
			ORDER BY created_at DESC
			LIMIT 1
		) cs ON TRUE
		WHERE daq.attempt_id = $1::uuid
		GROUP BY daq.attempt_id, daq.topic_id
		ON CONFLICT (attempt_id, topic_id) DO UPDATE SET
		  score = EXCLUDED.score,
		  max_score = EXCLUDED.max_score,
		  percentage = EXCLUDED.percentage,
		  passed = EXCLUDED.passed
	`, attemptID); err != nil {
		return fmt.Errorf("insert topic results: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_topic_progress (user_id, topic_id, status, mastery_score, completed_at)
		SELECT $1::uuid,
		       dtr.topic_id,
		       CASE WHEN dtr.percentage >= 80 THEN 'completed'::learning_progress_status ELSE 'in_progress'::learning_progress_status END,
		       dtr.percentage,
		       CASE WHEN dtr.percentage >= 80 THEN NOW() ELSE NULL END
		FROM diagnostic_topic_results dtr
		WHERE dtr.attempt_id = $2::uuid
		ON CONFLICT (user_id, topic_id)
		DO UPDATE SET
		  mastery_score = GREATEST(user_topic_progress.mastery_score, EXCLUDED.mastery_score),
		  status = CASE
		    WHEN GREATEST(user_topic_progress.mastery_score, EXCLUDED.mastery_score) >= 80 THEN 'completed'::learning_progress_status
		    ELSE user_topic_progress.status
		  END,
		  completed_at = CASE
		    WHEN GREATEST(user_topic_progress.mastery_score, EXCLUDED.mastery_score) >= 80 THEN NOW()
		    ELSE user_topic_progress.completed_at
		  END
	`, userID, attemptID); err != nil {
		return fmt.Errorf("upsert user_topic_progress: %w", err)
	}

	res, err := tx.Exec(ctx, `
		UPDATE test_attempts
		SET status='submitted',
		    submitted_at=NOW(),
		    score=(SELECT COALESCE(SUM(score),0) FROM diagnostic_topic_results WHERE attempt_id=$1::uuid)
		WHERE id=$1::uuid
		  AND status='in_progress'
	`, attemptID)
	if err != nil {
		return fmt.Errorf("finalize attempt: %w", err)
	}
	if res.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}
	return tx.Commit(ctx)
}

func (r *Repository) MarkAttemptExpired(ctx context.Context, attemptID string) error {
	if r.pool == nil {
		return fmt.Errorf("assessment repository is not initialized")
	}
	_, err := r.pool.Exec(ctx, `UPDATE test_attempts SET status='expired' WHERE id=$1::uuid AND status IN ('started','in_progress')`, attemptID)
	return err
}

func (r *Repository) GetDiagnosticAttemptStatus(ctx context.Context, attemptID string) (DiagnosticAttemptStatus, error) {
	if r.pool == nil {
		return DiagnosticAttemptStatus{}, fmt.Errorf("assessment repository is not initialized")
	}
	const q = `
	SELECT ta.id::text,
	       ta.user_id::text,
	       ta.status::text,
	       ta.started_at,
	       ta.expires_at,
	       ta.submitted_at,
	       COALESCE(MAX(daq.order_index) FILTER (WHERE daq.answered_at IS NOT NULL), 0) AS last_answered,
	       COUNT(daq.question_id) AS total_questions,
	       COUNT(daq.question_id) FILTER (WHERE daq.answered_at IS NOT NULL) AS answered_questions,
	       COALESCE(SUM(daq.allotted_seconds),0) AS total_allowed_seconds
	FROM test_attempts ta
	LEFT JOIN diagnostic_attempt_questions daq ON daq.attempt_id = ta.id
	WHERE ta.id=$1::uuid
	GROUP BY ta.id, ta.user_id, ta.status, ta.started_at, ta.expires_at, ta.submitted_at
	`
	var s DiagnosticAttemptStatus
	if err := r.pool.QueryRow(ctx, q, attemptID).Scan(
		&s.AttemptID, &s.UserID, &s.Status, &s.StartedAt, &s.ExpiresAt, &s.SubmittedAt,
		&s.LastAnsweredOrderIdx, &s.TotalQuestions, &s.AnsweredQuestions, &s.TotalAllowedSeconds,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DiagnosticAttemptStatus{}, nil
		}
		return DiagnosticAttemptStatus{}, fmt.Errorf("attempt status: %w", err)
	}
	return s, nil
}

func (r *Repository) GetUserStatus(ctx context.Context, userID string) (UserStatus, error) {
	if r.pool == nil {
		return UserStatus{}, fmt.Errorf("assessment repository is not initialized")
	}
	const q = `
	SELECT EXISTS (
		SELECT 1
		FROM test_attempts ta
		JOIN tests t ON t.id = ta.test_id
		WHERE ta.user_id = $1::uuid
		  AND t.type = 'diagnostic'
		  AND ta.status = 'submitted'
	)
	`
	var completed bool
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&completed); err != nil {
		return UserStatus{}, fmt.Errorf("query diagnostic status: %w", err)
	}
	return UserStatus{DiagnosticCompleted: completed, DashboardUnlocked: completed}, nil
}
