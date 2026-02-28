package learning

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListTopics(ctx context.Context) error {
	_ = ctx
	return nil
}

func (r *Repository) GetTopic(ctx context.Context, topicID string) error {
	_ = ctx
	_ = topicID
	return nil
}

func (r *Repository) SaveLearningQuestionCompletion(ctx context.Context, userID, questionID string) error {
	if r.pool == nil {
		return fmt.Errorf("learning repository is not initialized")
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

func (r *Repository) SaveLearningQuestionCompletionTx(ctx context.Context, tx pgx.Tx, userID, questionID string) error {
	_ = ctx
	_ = userID
	_ = questionID
	_ = tx
	return nil
}
