package dashboard

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type ExternalSolveInput struct {
	UserID                 string
	Platform               string
	PlatformQuestionID     string
	Title                  string
	TopicID                *string
	Difficulty             string
	Link                   string
	SolvedAt               time.Time
	TotalExternalQuestions int
}

type Summary struct {
	StreakCount     int        `json:"streak_count"`
	QuestionsSolved int        `json:"questions_solved"`
	MasteryScore    float64    `json:"mastery_score"`
	TopicsCompleted int        `json:"topics_completed"`
	LastActivityAt  *time.Time `json:"last_activity_at,omitempty"`
}

type HeatmapPoint struct {
	Date            time.Time `json:"date"`
	QuestionsSolved int       `json:"questions_solved"`
	HasDiagnostic   bool      `json:"has_diagnostic"`
}

type ActivityItem struct {
	Source     string    `json:"source"`
	Title      string    `json:"title"`
	TopicID    *string   `json:"topic_id,omitempty"`
	Difficulty string    `json:"difficulty,omitempty"`
	Link       string    `json:"link,omitempty"`
	SolvedAt   time.Time `json:"solved_at"`
}

type TopicMastery struct {
	TopicID      string  `json:"topic_id"`
	TopicName    string  `json:"topic_name"`
	Status       string  `json:"status"`
	MasteryScore float64 `json:"mastery_score"`
}

type UpcomingEvent struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Phase     string    `json:"phase"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type PlatformConnection struct {
	ID              string     `json:"id"`
	Platform        string     `json:"platform"`
	PlatformHandle  string     `json:"platform_handle"`
	Status          string     `json:"status"`
	ConnectedAt     *time.Time `json:"connected_at,omitempty"`
	LastValidatedAt *time.Time `json:"last_validated_at,omitempty"`
}

type PlatformSyncJob struct {
	ID            string     `json:"id"`
	ConnectionID  string     `json:"connection_id"`
	Status        string     `json:"status"`
	TriggerSource string     `json:"trigger_source"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
}

type PlatformSyncOverview struct {
	Queued       int               `json:"queued"`
	Running      int               `json:"running"`
	Succeeded    int               `json:"succeeded"`
	Failed       int               `json:"failed"`
	RateLimited  int               `json:"rate_limited"`
	LastError    string            `json:"last_error,omitempty"`
	RecentJobs   []PlatformSyncJob `json:"recent_jobs"`
	WindowHours  int               `json:"window_hours"`
	TotalInRange int               `json:"total_in_range"`
}

type DashboardData struct {
	Summary        Summary         `json:"summary"`
	Heatmap        []HeatmapPoint  `json:"heatmap"`
	RecentActivity []ActivityItem  `json:"recent_activity"`
	TopicMastery   []TopicMastery  `json:"topic_mastery"`
	WeakTopics     []TopicMastery  `json:"weak_topics"`
	UpcomingEvents []UpcomingEvent `json:"upcoming_events"`
}

func (r *Repository) GetDashboard(ctx context.Context, userID string) (DashboardData, error) {
	if r.pool == nil {
		return DashboardData{}, fmt.Errorf("dashboard repository is not initialized")
	}

	out := DashboardData{
		Heatmap:        make([]HeatmapPoint, 0),
		RecentActivity: make([]ActivityItem, 0),
		TopicMastery:   make([]TopicMastery, 0),
		WeakTopics:     make([]TopicMastery, 0),
		UpcomingEvents: make([]UpcomingEvent, 0),
	}

	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(streak_count,0), COALESCE(questions_solved,0), COALESCE(mastery_score,0), COALESCE(topics_completed,0), last_activity_at
		FROM dashboard_daily_snapshots
		WHERE user_id=$1::uuid
		ORDER BY snapshot_date DESC
		LIMIT 1
	`, userID).Scan(&out.Summary.StreakCount, &out.Summary.QuestionsSolved, &out.Summary.MasteryScore, &out.Summary.TopicsCompleted, &out.Summary.LastActivityAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			out.Summary = Summary{
				StreakCount:     0,
				QuestionsSolved: 0,
				MasteryScore:    0.0,
				TopicsCompleted: 0,
				LastActivityAt:  nil,
			}
		} else {
			return DashboardData{}, err
		}
	}

	hRows, err := r.pool.Query(ctx, `
		WITH diagnostic_days AS (
			SELECT DISTINCT solved_at::date AS day
			FROM user_activity_feed
			WHERE user_id=$1::uuid AND source='diagnostic'
		)
		SELECT da.activity_date::timestamp,
			   da.questions_solved,
			   (dd.day IS NOT NULL) AS has_diagnostic
		FROM daily_activity da
		LEFT JOIN diagnostic_days dd ON dd.day = da.activity_date
		WHERE da.user_id=$1::uuid
		  AND da.activity_date >= CURRENT_DATE - INTERVAL '365 days'
		ORDER BY da.activity_date DESC
	`, userID)
	if err != nil {
		return DashboardData{}, err
	}
	defer hRows.Close()

	for hRows.Next() {
		var p HeatmapPoint
		if err := hRows.Scan(&p.Date, &p.QuestionsSolved, &p.HasDiagnostic); err != nil {
			return DashboardData{}, err
		}
		out.Heatmap = append(out.Heatmap, p)
	}

	aRows, err := r.pool.Query(ctx, `
		SELECT source, title, topic_id::text, COALESCE(difficulty, ''), COALESCE(link, ''), solved_at
		FROM user_activity_feed
		WHERE user_id=$1::uuid
		ORDER BY solved_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return DashboardData{}, fmt.Errorf("failed to fetch recent activity: %w", err)
	}
	defer aRows.Close()

	for aRows.Next() {
		var a ActivityItem
		if err := aRows.Scan(&a.Source, &a.Title, &a.TopicID, &a.Difficulty, &a.Link, &a.SolvedAt); err != nil {
			return DashboardData{}, fmt.Errorf("failed to scan activity item: %w", err)
		}
		out.RecentActivity = append(out.RecentActivity, a)
	}

	mRows, err := r.pool.Query(ctx, `
		SELECT utp.topic_id::text, t.name, utp.status::text, COALESCE(utp.mastery_score,0)
		FROM user_topic_progress utp
		JOIN topics t ON t.id=utp.topic_id
		WHERE utp.user_id=$1::uuid
		  AND utp.status <> 'not_started'::learning_progress_status
		ORDER BY t.name
		LIMIT 21
	`, userID)
	if err != nil {
		return DashboardData{}, err
	}
	defer mRows.Close()

	for mRows.Next() {
		var m TopicMastery
		if err := mRows.Scan(&m.TopicID, &m.TopicName, &m.Status, &m.MasteryScore); err != nil {
			return DashboardData{}, err
		}
		out.TopicMastery = append(out.TopicMastery, m)
		if m.MasteryScore < 40 {
			out.WeakTopics = append(out.WeakTopics, m)
		}
	}

	eRows, err := r.pool.Query(ctx, `
		SELECT id::text, name, phase::text, start_date::timestamp, end_date::timestamp
		FROM events
		WHERE end_date >= CURRENT_DATE
		ORDER BY start_date ASC
		LIMIT 5
	`)
	if err != nil {
		return DashboardData{}, err
	}
	defer eRows.Close()

	for eRows.Next() {
		var e UpcomingEvent
		if err := eRows.Scan(&e.ID, &e.Name, &e.Phase, &e.StartDate, &e.EndDate); err != nil {
			return DashboardData{}, err
		}
		out.UpcomingEvents = append(out.UpcomingEvents, e)
	}

	return out, nil
}

func (r *Repository) TriggerPlatformSync(ctx context.Context, userID string) error {
	if r.pool == nil {
		return fmt.Errorf("dashboard repository is not initialized")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO platform_sync_jobs (id, user_id, connection_id, status, trigger_source, started_at)
		SELECT gen_random_uuid(), pc.user_id, pc.id, 'queued'::sync_job_status, 'user', NOW()
		FROM platform_connections pc
		WHERE pc.user_id=$1::uuid
		  AND pc.status='connected'
	`, userID); err != nil {
		return err
	}

	var jobID string
	err = tx.QueryRow(ctx, `
		SELECT id::text
		FROM platform_sync_jobs
		WHERE user_id=$1::uuid
		  AND status='queued'::sync_job_status
		ORDER BY started_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`, userID).Scan(&jobID)
	if err == pgx.ErrNoRows {
		return tx.Commit(ctx)
	}
	if err != nil {
		return err
	}

	res, err := tx.Exec(ctx, `
		UPDATE platform_sync_jobs
		SET status='running'::sync_job_status, started_at=NOW()
		WHERE id=$1::uuid AND status='queued'::sync_job_status
	`, jobID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to claim sync job")
	}

	res, err = tx.Exec(ctx, `
		UPDATE platform_sync_jobs
		SET status='succeeded'::sync_job_status, finished_at=NOW()
		WHERE id=$1::uuid AND status='running'::sync_job_status
	`, jobID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to finish sync job")
	}

	return tx.Commit(ctx)
}

func (r *Repository) RecordExternalSolve(ctx context.Context, in ExternalSolveInput) error {
	if r.pool == nil {
		return fmt.Errorf("dashboard repository is not initialized")
	}
	if in.TopicID == nil || *in.TopicID == "" {
		return fmt.Errorf("topic_id is required")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var activityID string
	err = tx.QueryRow(ctx, `
		INSERT INTO external_question_activity (id,user_id,platform,platform_question_id,title,topic_id,difficulty,solved_at)
		VALUES (gen_random_uuid(),$1::uuid,$2,$3,$4,$5::uuid,$6,$7)
		ON CONFLICT (user_id,platform,platform_question_id) DO NOTHING
		RETURNING id::text
	`, in.UserID, in.Platform, in.PlatformQuestionID, in.Title, *in.TopicID, in.Difficulty, in.SolvedAt).Scan(&activityID)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	if activityID == "" {
		return tx.Commit(ctx)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO daily_activity (user_id, activity_date, questions_solved)
		VALUES ($1::uuid, $2::date, 1)
		ON CONFLICT (user_id, activity_date)
		DO UPDATE SET questions_solved = daily_activity.questions_solved + 1
		`, in.UserID, in.SolvedAt); err != nil {
		return err
	}

	utc := in.SolvedAt.UTC()
	today := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	var todayStreak int
	err = tx.QueryRow(ctx, `
		SELECT streak_count
		FROM dashboard_daily_snapshots
		WHERE user_id=$1::uuid AND snapshot_date=$2::date
		FOR UPDATE
	`, in.UserID, today).Scan(&todayStreak)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	if err == pgx.ErrNoRows {
		newStreak := 1
		var yStreak, yQuestions int
		yesterday := today.Add(-24 * time.Hour)
		if err := tx.QueryRow(ctx, `
			SELECT streak_count, questions_solved
			FROM dashboard_daily_snapshots
			WHERE user_id=$1::uuid AND snapshot_date=$2::date
		`, in.UserID, yesterday).Scan(&yStreak, &yQuestions); err != nil {
			if err != pgx.ErrNoRows {
				return err
			}
		} else if yQuestions > 0 {
			newStreak = yStreak + 1
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO dashboard_daily_snapshots (user_id, snapshot_date, streak_count, questions_solved, mastery_score, topics_completed, last_activity_at, computed_at)
			VALUES ($1::uuid, $2::date, $3, 0, 0, 0, $2, NOW())
			ON CONFLICT (user_id, snapshot_date) DO NOTHING
		`, in.UserID, today, newStreak); err != nil {
			return err
		}
	}

	res, err := tx.Exec(ctx, `
		UPDATE dashboard_daily_snapshots
		SET questions_solved = questions_solved + 1,
			last_activity_at = GREATEST(last_activity_at, $3),
			computed_at = NOW()
		WHERE user_id=$1::uuid AND snapshot_date=$2::date
	`, in.UserID, today, in.SolvedAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to increment dashboard snapshot questions")
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_activity_feed (id,user_id,source,title,topic_id,difficulty,link,solved_at)
		VALUES (gen_random_uuid(),$1::uuid,'external',$2,$3::uuid,$4,$5,$6)
	`, in.UserID, in.Title, *in.TopicID, in.Difficulty, in.Link, in.SolvedAt); err != nil {
		return err
	}

	var status string
	var externalSolved, totalExternal int
	var diagnosticMastery float64
	err = tx.QueryRow(ctx, `
		SELECT status::text, external_solved_count, total_external_questions, COALESCE(diagnostic_mastery,0)
		FROM user_topic_progress
		WHERE user_id=$1::uuid AND topic_id=$2::uuid
		FOR UPDATE
	`, in.UserID, *in.TopicID).Scan(&status, &externalSolved, &totalExternal, &diagnosticMastery)
	if err == pgx.ErrNoRows || status == "not_started" {
		return tx.Commit(ctx)
	}
	if err != nil {
		return err
	}

	if in.TotalExternalQuestions <= 0 {
		in.TotalExternalQuestions = totalExternal
	}
	if in.TotalExternalQuestions < 0 {
		in.TotalExternalQuestions = 0
	}

	if totalExternal == 0 && in.TotalExternalQuestions > 0 {
		totalExternal = in.TotalExternalQuestions
	}
	externalSolved++
	externalMastery := 0.0
	if totalExternal > 0 {
		externalMastery = (float64(externalSolved) / float64(totalExternal)) * 100.0
	}
	mastery := diagnosticMastery
	if externalMastery > mastery {
		mastery = externalMastery
	}

	prevCompleted := status == "completed"
	res, err = tx.Exec(ctx, `
		UPDATE user_topic_progress
		SET external_solved_count=$3,
			total_external_questions=$4,
			mastery_score=$5,
			status = CASE WHEN $5 >= 80 THEN 'completed'::learning_progress_status ELSE status END,
			completed_at = CASE WHEN $5 >= 80 AND status != 'completed'::learning_progress_status THEN NOW() ELSE completed_at END
		WHERE user_id=$1::uuid AND topic_id=$2::uuid
	`, in.UserID, *in.TopicID, externalSolved, totalExternal, mastery)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to update topic progress")
	}

	if mastery >= 80 && !prevCompleted {
		res, err = tx.Exec(ctx, `
			UPDATE dashboard_daily_snapshots
			SET topics_completed = topics_completed + 1, computed_at=NOW()
			WHERE user_id=$1::uuid AND snapshot_date=$2::date
		`, in.UserID, today)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("failed to increment dashboard snapshot topics_completed")
		}
	}

	if mastery >= 80 {
		if err := r.evaluateUnlocksTx(ctx, tx, in.UserID, *in.TopicID); err != nil {
			return err
		}
	}

	res, err = tx.Exec(ctx, `
		UPDATE dashboard_daily_snapshots
		SET mastery_score = COALESCE((
			SELECT AVG(mastery_score) FROM user_topic_progress
			WHERE user_id=$1::uuid
			  AND status <> 'not_started'::learning_progress_status
		),0),
		computed_at=NOW()
		WHERE user_id=$1::uuid AND snapshot_date=$2::date
	`, in.UserID, today)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to update dashboard snapshot mastery")
	}

	return tx.Commit(ctx)
}

func (r *Repository) evaluateUnlocksTx(ctx context.Context, tx pgx.Tx, userID, completedTopicID string) error {
	rows, err := tx.Query(ctx, `SELECT topic_id::text FROM topic_prerequisites WHERE prerequisite_id=$1::uuid`, completedTopicID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var topicID string
		if err := rows.Scan(&topicID); err != nil {
			return err
		}
		var blocked int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM topic_prerequisites tp
			LEFT JOIN user_topic_progress utp
			  ON utp.user_id=$1::uuid AND utp.topic_id=tp.prerequisite_id
			WHERE tp.topic_id=$2::uuid
			  AND COALESCE(utp.mastery_score,0) < 80
		`, userID, topicID).Scan(&blocked); err != nil {
			return err
		}
		if blocked == 0 {
			if _, err := tx.Exec(ctx, `
				INSERT INTO user_topic_progress (user_id, topic_id, status, mastery_score, completed_at, external_solved_count, total_external_questions, diagnostic_mastery)
				VALUES ($1::uuid,$2::uuid,'in_progress',0,NULL,0,0,0)
				ON CONFLICT (user_id, topic_id) DO NOTHING
			`, userID, topicID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Repository) ListPlatformConnections(ctx context.Context, userID string) ([]PlatformConnection, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("dashboard repository is not initialized")
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, platform, platform_handle, status, connected_at, last_validated_at
		FROM platform_connections
		WHERE user_id=$1::uuid
		ORDER BY platform ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PlatformConnection, 0)
	for rows.Next() {
		var c PlatformConnection
		if err := rows.Scan(&c.ID, &c.Platform, &c.PlatformHandle, &c.Status, &c.ConnectedAt, &c.LastValidatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertPlatformConnection(ctx context.Context, userID, platform, platformHandle string) (PlatformConnection, error) {
	if r.pool == nil {
		return PlatformConnection{}, fmt.Errorf("dashboard repository is not initialized")
	}
	var c PlatformConnection
	err := r.pool.QueryRow(ctx, `
		INSERT INTO platform_connections (id, user_id, platform, platform_handle, status, connected_at, last_validated_at)
		VALUES (gen_random_uuid(), $1::uuid, $2, $3, 'connected', NOW(), NOW())
		ON CONFLICT (user_id, platform)
		DO UPDATE SET platform_handle=EXCLUDED.platform_handle, status='connected', last_validated_at=NOW()
		RETURNING id::text, platform, platform_handle, status, connected_at, last_validated_at
	`, userID, platform, platformHandle).Scan(&c.ID, &c.Platform, &c.PlatformHandle, &c.Status, &c.ConnectedAt, &c.LastValidatedAt)
	return c, err
}

func (r *Repository) DeletePlatformConnection(ctx context.Context, userID, platform string) error {
	if r.pool == nil {
		return fmt.Errorf("dashboard repository is not initialized")
	}

	// Start a transaction to ensure both deletes happen safely
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Delete dependent sync jobs first to prevent foreign key violations
	_, err = tx.Exec(ctx, `
		DELETE FROM platform_sync_jobs 
		WHERE connection_id IN (
			SELECT id FROM platform_connections WHERE user_id=$1::uuid AND platform=$2
		)
	`, userID, platform)
	if err != nil {
		return fmt.Errorf("failed to delete dependent sync jobs: %w", err)
	}

	// 2. Now it is safe to delete the platform connection
	_, err = tx.Exec(ctx, `
		DELETE FROM platform_connections 
		WHERE user_id=$1::uuid AND platform=$2
	`, userID, platform)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	// Commit the transaction to save the changes
	return tx.Commit(ctx)
}

func (r *Repository) GetPlatformSyncJob(ctx context.Context, userID, jobID string) (PlatformSyncJob, error) {
	if r.pool == nil {
		return PlatformSyncJob{}, fmt.Errorf("dashboard repository is not initialized")
	}
	var j PlatformSyncJob
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, connection_id::text, status::text, trigger_source, started_at, finished_at, error_message
		FROM platform_sync_jobs
		WHERE id=$1::uuid AND user_id=$2::uuid
	`, jobID, userID).Scan(&j.ID, &j.ConnectionID, &j.Status, &j.TriggerSource, &j.StartedAt, &j.FinishedAt, &j.ErrorMessage)
	return j, err
}

func (r *Repository) GetLatestPlatformSyncJob(ctx context.Context, userID string) (PlatformSyncJob, error) {
	if r.pool == nil {
		return PlatformSyncJob{}, fmt.Errorf("dashboard repository is not initialized")
	}
	var j PlatformSyncJob
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, connection_id::text, status::text, trigger_source, started_at, finished_at, error_message
		FROM platform_sync_jobs
		WHERE user_id=$1::uuid
		ORDER BY started_at DESC NULLS LAST, id DESC
		LIMIT 1
	`, userID).Scan(&j.ID, &j.ConnectionID, &j.Status, &j.TriggerSource, &j.StartedAt, &j.FinishedAt, &j.ErrorMessage)
	return j, err
}

func (r *Repository) GetPlatformSyncOverview(ctx context.Context, userID string, windowHours int, recentLimit int) (PlatformSyncOverview, error) {
	if r.pool == nil {
		return PlatformSyncOverview{}, fmt.Errorf("dashboard repository is not initialized")
	}
	if windowHours <= 0 {
		windowHours = 24
	}
	if recentLimit <= 0 {
		recentLimit = 10
	}

	overview := PlatformSyncOverview{WindowHours: windowHours, RecentJobs: make([]PlatformSyncJob, 0)}

	rows, err := r.pool.Query(ctx, `
		SELECT status::text, COUNT(*)::int
		FROM platform_sync_jobs
		WHERE user_id=$1::uuid
		  AND started_at >= NOW() - ($2::int * interval '1 hour')
		GROUP BY status
	`, userID, windowHours)
	if err != nil {
		return PlatformSyncOverview{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return PlatformSyncOverview{}, err
		}
		overview.TotalInRange += count
		switch status {
		case "queued":
			overview.Queued = count
		case "running":
			overview.Running = count
		case "succeeded":
			overview.Succeeded = count
		case "failed":
			overview.Failed = count
		case "rate_limited":
			overview.RateLimited = count
		}
	}
	if err := rows.Err(); err != nil {
		return PlatformSyncOverview{}, err
	}

	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(error_message,'')
		FROM platform_sync_jobs
		WHERE user_id=$1::uuid
		  AND error_message IS NOT NULL
		ORDER BY started_at DESC NULLS LAST, id DESC
		LIMIT 1
	`, userID).Scan(&overview.LastError)

	recentRows, err := r.pool.Query(ctx, `
		SELECT id::text, connection_id::text, status::text, trigger_source, started_at, finished_at, error_message
		FROM platform_sync_jobs
		WHERE user_id=$1::uuid
		ORDER BY started_at DESC NULLS LAST, id DESC
		LIMIT $2
	`, userID, recentLimit)
	if err != nil {
		return PlatformSyncOverview{}, err
	}
	defer recentRows.Close()

	for recentRows.Next() {
		var j PlatformSyncJob
		if err := recentRows.Scan(&j.ID, &j.ConnectionID, &j.Status, &j.TriggerSource, &j.StartedAt, &j.FinishedAt, &j.ErrorMessage); err != nil {
			return PlatformSyncOverview{}, err
		}
		overview.RecentJobs = append(overview.RecentJobs, j)
	}
	if err := recentRows.Err(); err != nil {
		return PlatformSyncOverview{}, err
	}

	return overview, nil
}
