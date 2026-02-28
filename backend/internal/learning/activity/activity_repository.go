package activity

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityRepository struct {
	pool *pgxpool.Pool
}

func NewActivityRepository(pool *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{pool: pool}
}

func (r *ActivityRepository) SaveLearningQuestionCompletion(ctx context.Context, userID, questionID string) error {
	if r.pool == nil {
		return fmt.Errorf("learning activity repository is not initialized")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := r.SaveLearningQuestionCompletionTx(ctx, tx, userID, questionID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ActivityRepository) SaveLearningQuestionCompletionTx(ctx context.Context, tx pgx.Tx, userID, questionID string) error {
	var topicID string
	if err := tx.QueryRow(ctx, `SELECT topic_id::text FROM learning_questions WHERE id=$1::uuid`, questionID).Scan(&topicID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_learning_question_activity (user_id, question_id, solved, solved_at, time_taken_minutes)
		VALUES ($1::uuid, $2::uuid, true, CURRENT_DATE, NULL)
		ON CONFLICT (user_id, question_id)
		DO UPDATE SET solved=true, solved_at=CURRENT_DATE
	`, userID, questionID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_topic_progress (user_id, topic_id, status, mastery_score, completed_at)
		VALUES ($1::uuid, $2::uuid, 'in_progress'::learning_progress_status, 0, NULL)
		ON CONFLICT (user_id, topic_id) DO NOTHING
	`, userID, topicID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE user_topic_progress
		SET status='in_progress'::learning_progress_status
		WHERE user_id=$1::uuid
		  AND topic_id=$2::uuid
		  AND status='not_started'::learning_progress_status
	`, userID, topicID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_activity_feed (id, user_id, source, title, topic_id, solved_at)
		VALUES (gen_random_uuid(), $1::uuid, 'ide', 'IDE question completed', $2::uuid, NOW())
	`, userID, topicID); err != nil {
		return err
	}

	return nil
}
