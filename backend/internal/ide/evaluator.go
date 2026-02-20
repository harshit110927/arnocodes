package ide

import "context"

type Evaluator interface {
	Evaluate(ctx context.Context, submission Submission, testCases []CodingQuestionTestCase) (EvaluationResult, error)
}
