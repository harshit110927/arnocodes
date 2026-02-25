package assessment

import (
	"context"
	"log"
	"time"
)

// StartCodingEvaluationWorker runs a polling loop that evaluates pending coding submissions.
// Real sandbox integration should replace evaluateSubmissionStub inside the service.
func StartCodingEvaluationWorker(ctx context.Context, svc *Service, interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := svc.HandleCodingEvaluationWorker(ctx, batchSize); err != nil {
				log.Printf("coding worker cycle failed: %v", err)
			}
		}
	}
}
