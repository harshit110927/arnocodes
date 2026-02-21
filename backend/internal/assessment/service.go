package assessment

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrDiagnosticBlocked = errors.New("DIAGNOSTIC_BLOCKED")
	ErrNotFound          = errors.New("NOT_FOUND")
	ErrUnauthorized      = errors.New("UNAUTHORIZED")
	ErrTimeExpired       = errors.New("TIME_EXPIRED")
	ErrInvalidInput      = errors.New("UNPROCESSABLE_ENTITY")
)

type AnswerData struct {
	QuestionType   string
	SelectedOption *int
	Code           string
	Language       string
	QuestionID     string
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) StartDiagnostic(ctx context.Context, userID string, selectedTopics []string) (string, error) {
	if userID == "" {
		return "", ErrUnauthorized
	}
	if len(selectedTopics) == 0 {
		return "", fmt.Errorf("%w: selected_topics required", ErrInvalidInput)
	}
	if err := s.repo.ValidateTopicSelection(selectedTopics); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	canStart, err := s.repo.CanStartDiagnostic(userID)
	if err != nil {
		return "", err
	}
	if !canStart {
		return "", ErrDiagnosticBlocked
	}
	return s.repo.CreateDiagnosticAttempt(ctx, userID, selectedTopics)
}

func (s *Service) FetchNextQuestion(ctx context.Context, userID, attemptID string) (DiagnosticQuestion, error) {
	status, err := s.repo.GetDiagnosticAttemptStatus(ctx, attemptID)
	if err != nil {
		return DiagnosticQuestion{}, err
	}
	if status.AttemptID == "" {
		return DiagnosticQuestion{}, ErrNotFound
	}
	if status.UserID != userID {
		return DiagnosticQuestion{}, ErrUnauthorized
	}
	if status.Status == "submitted" || status.Status == "expired" {
		return DiagnosticQuestion{}, ErrTimeExpired
	}
	if time.Now().After(status.ExpiresAt) {
		_ = s.repo.MarkAttemptExpired(ctx, attemptID)
		return DiagnosticQuestion{}, ErrTimeExpired
	}
	return s.repo.GetNextDiagnosticQuestion(ctx, attemptID, status.LastAnsweredOrderIdx)
}

func (s *Service) SubmitAnswer(ctx context.Context, userID, attemptID string, data AnswerData) (string, error) {
	status, err := s.repo.GetDiagnosticAttemptStatus(ctx, attemptID)
	if err != nil {
		return "", err
	}
	if status.AttemptID == "" {
		return "", ErrNotFound
	}
	if status.UserID != userID {
		return "", ErrUnauthorized
	}
	if time.Now().After(status.ExpiresAt) {
		_ = s.repo.MarkAttemptExpired(ctx, attemptID)
		return "", ErrTimeExpired
	}

	switch data.QuestionType {
	case "mcq":
		if data.SelectedOption == nil {
			return "", fmt.Errorf("%w: selected_option required", ErrInvalidInput)
		}
		if err := s.repo.SubmitMCQAnswer(ctx, attemptID, data.QuestionID, *data.SelectedOption); err != nil {
			return "", err
		}
		return "", nil
	case "coding":
		if data.Code == "" || data.Language == "" {
			return "", fmt.Errorf("%w: code and language required", ErrInvalidInput)
		}
		submissionID, err := s.repo.SaveCodingSubmission(ctx, attemptID, data.QuestionID, userID, data.Code, data.Language)
		if err != nil {
			return "", err
		}
		return submissionID, nil
	default:
		return "", fmt.Errorf("%w: unsupported question type", ErrInvalidInput)
	}
}

func (s *Service) SubmitTest(ctx context.Context, userID, attemptID string) error {
	status, err := s.repo.GetDiagnosticAttemptStatus(ctx, attemptID)
	if err != nil {
		return err
	}
	if status.AttemptID == "" {
		return ErrNotFound
	}
	if status.UserID != userID {
		return ErrUnauthorized
	}
	return s.repo.CompleteDiagnosticAttempt(ctx, attemptID)
}

func (s *Service) HandleCodingEvaluationWorker(ctx context.Context, limit int) error {
	subs, err := s.repo.GetPendingCodingSubmissions(ctx, limit)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		score := evaluateSubmissionStub(sub)
		if err := s.repo.UpdateCodingSubmissionResult(ctx, sub.ID, "completed", score); err != nil {
			return err
		}
	}
	return nil
}

func evaluateSubmissionStub(sub CodingSubmission) float64 {
	if len(sub.Code) == 0 {
		return 0
	}
	if len(sub.Code) > 500 {
		return 100
	}
	return 70
}
