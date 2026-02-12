package assessment

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type UserStatus struct {
	DiagnosticCompleted bool `json:"diagnostic_completed"`
	DashboardUnlocked   bool `json:"dashboard_unlocked"`
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

func (r *Repository) GetTestByID(ctx context.Context, testID string) error {
	_ = ctx
	_ = testID
	return nil
}

func (r *Repository) StartAttempt(ctx context.Context, userID, testID string) error {
	_ = ctx
	_ = userID
	_ = testID
	return nil
}

func (r *Repository) SubmitAnswer(ctx context.Context, attemptID, questionID string, selectedOption int) error {
	_ = ctx
	_ = attemptID
	_ = questionID
	_ = selectedOption
	return nil
}

func (r *Repository) SubmitAttempt(ctx context.Context, attemptID string) error {
	_ = ctx
	_ = attemptID
	return nil
}
