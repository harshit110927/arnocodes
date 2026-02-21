package ide

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("NOT_FOUND")

type MasteryUpdater interface {
	SaveLearningQuestionCompletion(ctx context.Context, userID, questionID string) error
}

type Service struct {
	repo      *Repository
	evaluator Evaluator
}

func NewService(repo *Repository, evaluator Evaluator, masteryUpdate MasteryUpdater) *Service {
	_ = masteryUpdate
	return &Service{repo: repo, evaluator: evaluator}
}

func (s *Service) Submit(ctx context.Context, userID string, in Submission) (string, error) {
	exists, err := s.repo.QuestionExists(ctx, in.QuestionID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrNotFound
	}
	in.UserID = userID
	return s.repo.CreateSubmission(ctx, in)
}

func (s *Service) Status(ctx context.Context, userID, submissionID string) (SubmissionStatus, error) {
	return s.repo.GetSubmissionStatus(ctx, submissionID, userID)
}

func (s *Service) RunSample(ctx context.Context, userID string, in Submission) (EvaluationResult, error) {
	exists, err := s.repo.QuestionExists(ctx, in.QuestionID)
	if err != nil {
		return EvaluationResult{}, err
	}
	if !exists {
		return EvaluationResult{}, ErrNotFound
	}
	tests, err := s.repo.ListTestCases(ctx, in.QuestionID, true)
	if err != nil {
		return EvaluationResult{}, err
	}
	in.UserID = userID
	return s.evaluator.Evaluate(ctx, in, tests)
}

func (s *Service) ResetStuckProcessingSubmissions(ctx context.Context) error {
	return s.repo.ResetStuckProcessingSubmissions(ctx)
}
