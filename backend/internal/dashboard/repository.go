package dashboard

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

func (r *Repository) GetSummary(ctx context.Context, userID string) error {
	_ = ctx
	_ = userID
	return nil
}

func (r *Repository) GetHeatmap(ctx context.Context, userID string) error {
	_ = ctx
	_ = userID
	return nil
}
