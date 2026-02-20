package ide

import "time"

const PassThreshold = 80.0

type Submission struct {
	ID               string
	AttemptID        *string
	QuestionID       string
	UserID           string
	Code             string
	Language         string
	EvaluationStatus string
	Score            *float64
	CreatedAt        time.Time
	EvaluatedAt      *time.Time
}

type CodingQuestionTestCase struct {
	ID             string
	QuestionID     string
	Input          string
	ExpectedOutput string
	IsSample       bool
	Weight         float64
	OrderIndex     int
	CreatedAt      time.Time
}

type SubmissionStatus struct {
	EvaluationStatus string     `json:"evaluation_status"`
	Score            *float64   `json:"score,omitempty"`
	EvaluatedAt      *time.Time `json:"evaluated_at,omitempty"`
}

type EvaluationResult struct {
	Status string
	Score  *float64
	Detail string
}
