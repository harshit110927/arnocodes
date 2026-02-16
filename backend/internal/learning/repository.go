package learning

import (
	"context"

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
	_ = ctx
	_ = userID
	_ = questionID
	return nil
}
