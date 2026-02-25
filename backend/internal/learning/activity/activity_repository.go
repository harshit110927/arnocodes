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
	_ = ctx
	_ = userID
	_ = questionID
	_ = tx
	return nil
}
