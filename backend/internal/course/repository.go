package course

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

type TopicRow struct {
	TopicID          string
	Name             string
	ProgressStatus   *string
	MasteryScore     *float64
	SubtopicCount    int
	CompletionStatus string
}

type TopicPrerequisite struct {
	TopicID        string
	PrerequisiteID string
}

type TopicBase struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SubtopicProgressRow struct {
	SubtopicID    string   `json:"subtopic_id"`
	TopicID       string   `json:"topic_id"`
	Title         string   `json:"title"`
	OrderIndex    int      `json:"order_index"`
	Status        *string  `json:"status,omitempty"`
	MasteryScore  *float64 `json:"mastery_score,omitempty"`
	CompletedAt   *string  `json:"completed_at,omitempty"`
	IsCompleted   bool     `json:"is_completed"`
	DisplayStatus string   `json:"display_status"`
}

type LearningQuestionLink struct {
	ID         string  `json:"id"`
	Source     string  `json:"source"`
	Difficulty *string `json:"difficulty,omitempty"`
	Link       *string `json:"link,omitempty"`
}

func (r *CourseRepository) ListCourseTopicRows(ctx context.Context, userID string) ([]TopicRow, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("course repository is not initialized")
	}
	rows, err := r.pool.Query(ctx, `
		SELECT t.id::text,
		       t.name,
		       utp.status::text,
		       utp.mastery_score,
		       COALESCE(sc.subtopic_count, 0)
		FROM topics t
		LEFT JOIN user_topic_progress utp
		  ON utp.topic_id=t.id
		 AND utp.user_id=$1::uuid
		LEFT JOIN (
			SELECT topic_id, COUNT(*)::int AS subtopic_count
			FROM subtopics
			GROUP BY topic_id
		) sc ON sc.topic_id=t.id
		ORDER BY t.name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]TopicRow, 0)
	for rows.Next() {
		var tr TopicRow
		if err := rows.Scan(&tr.TopicID, &tr.Name, &tr.ProgressStatus, &tr.MasteryScore, &tr.SubtopicCount); err != nil {
			return nil, err
		}
		tr.CompletionStatus = "not_started"
		if tr.ProgressStatus != nil && *tr.ProgressStatus != "" {
			tr.CompletionStatus = *tr.ProgressStatus
		}
		out = append(out, tr)
	}
	return out, rows.Err()
}

func (r *CourseRepository) ListTopicPrerequisites(ctx context.Context) ([]TopicPrerequisite, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("course repository is not initialized")
	}
	rows, err := r.pool.Query(ctx, `SELECT topic_id::text, prerequisite_id::text FROM topic_prerequisites`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]TopicPrerequisite, 0)
	for rows.Next() {
		var p TopicPrerequisite
		if err := rows.Scan(&p.TopicID, &p.PrerequisiteID); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *CourseRepository) GetTopicBase(ctx context.Context, topicID string) (TopicBase, error) {
	if r.pool == nil {
		return TopicBase{}, fmt.Errorf("course repository is not initialized")
	}
	var t TopicBase
	err := r.pool.QueryRow(ctx, `SELECT id::text, name FROM topics WHERE id=$1::uuid`, topicID).Scan(&t.ID, &t.Name)
	return t, err
}

func (r *CourseRepository) ListSubtopicsWithProgressByTopic(ctx context.Context, userID, topicID string) ([]SubtopicProgressRow, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("course repository is not initialized")
	}
	rows, err := r.pool.Query(ctx, `
		SELECT s.id::text,
		       s.topic_id::text,
		       s.title,
		       s.order_index,
		       usp.status::text,
		       usp.mastery_score,
		       usp.completed_at::text
		FROM subtopics s
		LEFT JOIN user_subtopic_progress usp
		  ON usp.subtopic_id=s.id
		 AND usp.user_id=$1::uuid
		WHERE s.topic_id=$2::uuid
		ORDER BY s.order_index ASC
	`, userID, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SubtopicProgressRow, 0)
	for rows.Next() {
		var sp SubtopicProgressRow
		if err := rows.Scan(&sp.SubtopicID, &sp.TopicID, &sp.Title, &sp.OrderIndex, &sp.Status, &sp.MasteryScore, &sp.CompletedAt); err != nil {
			return nil, err
		}
		sp.DisplayStatus = "not_started"
		if sp.Status != nil && *sp.Status != "" {
			sp.DisplayStatus = *sp.Status
		}
		sp.IsCompleted = sp.DisplayStatus == "completed" || sp.DisplayStatus == "mastered"
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (r *CourseRepository) ListLearningQuestionsByTopic(ctx context.Context, topicID string) ([]LearningQuestionLink, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("course repository is not initialized")
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, source, difficulty, link
		FROM learning_questions
		WHERE topic_id=$1::uuid
		ORDER BY id
	`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LearningQuestionLink, 0)
	for rows.Next() {
		var q LearningQuestionLink
		if err := rows.Scan(&q.ID, &q.Source, &q.Difficulty, &q.Link); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (r *CourseRepository) GetSubtopicWithProgress(ctx context.Context, userID, subtopicID string) (SubtopicProgressRow, error) {
	if r.pool == nil {
		return SubtopicProgressRow{}, fmt.Errorf("course repository is not initialized")
	}
	var sp SubtopicProgressRow
	err := r.pool.QueryRow(ctx, `
		SELECT s.id::text,
		       s.topic_id::text,
		       s.title,
		       s.order_index,
		       usp.status::text,
		       usp.mastery_score,
		       usp.completed_at::text
		FROM subtopics s
		LEFT JOIN user_subtopic_progress usp
		  ON usp.subtopic_id=s.id
		 AND usp.user_id=$1::uuid
		WHERE s.id=$2::uuid
	`, userID, subtopicID).Scan(&sp.SubtopicID, &sp.TopicID, &sp.Title, &sp.OrderIndex, &sp.Status, &sp.MasteryScore, &sp.CompletedAt)
	if err != nil {
		return SubtopicProgressRow{}, err
	}
	sp.DisplayStatus = "not_started"
	if sp.Status != nil && *sp.Status != "" {
		sp.DisplayStatus = *sp.Status
	}
	sp.IsCompleted = sp.DisplayStatus == "completed" || sp.DisplayStatus == "mastered"
	return sp, nil
}
