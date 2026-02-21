package ide

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubmissionRepository interface {
	ClaimNextPendingSubmission(ctx context.Context) (Submission, bool, error)
	FinalizeSubmissionWithMastery(ctx context.Context, submissionID string, status string, score *float64, passThreshold float64, masteryUpdater MasteryUpdater) error
	ResetStuckProcessingSubmissions(ctx context.Context) error
}

type TestCaseRepository interface {
	ListTestCases(ctx context.Context, questionID string, sampleOnly bool) ([]CodingQuestionTestCase, error)
}

func StartIDEWorker(ctx context.Context, _ *pgxpool.Pool, submissionRepo SubmissionRepository, testCaseRepo TestCaseRepository, evaluator Evaluator, masteryUpdater MasteryUpdater) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := submissionRepo.ResetStuckProcessingSubmissions(ctx); err != nil {
				log.Printf("ide worker reset stuck submissions failed: %v", err)
			}
			sub, ok, err := submissionRepo.ClaimNextPendingSubmission(ctx)
			if err != nil {
				log.Printf("ide worker claim failed: %v", err)
				continue
			}
			if !ok {
				continue
			}

			tests, err := testCaseRepo.ListTestCases(ctx, sub.QuestionID, false)
			if err != nil {
				_ = submissionRepo.FinalizeSubmissionWithMastery(ctx, sub.ID, "failed", nil, PassThreshold, masteryUpdater)
				log.Printf("ide worker test-case load failed: %v", err)
				continue
			}

			result, err := evaluator.Evaluate(ctx, sub, tests)
			if err != nil {
				_ = submissionRepo.FinalizeSubmissionWithMastery(ctx, sub.ID, "failed", nil, PassThreshold, masteryUpdater)
				log.Printf("ide worker evaluate failed: %v", err)
				continue
			}

			if err := submissionRepo.FinalizeSubmissionWithMastery(ctx, sub.ID, result.Status, result.Score, PassThreshold, masteryUpdater); err != nil {
				log.Printf("ide worker finalize failed: %v", err)
				continue
			}

		}
	}
}
