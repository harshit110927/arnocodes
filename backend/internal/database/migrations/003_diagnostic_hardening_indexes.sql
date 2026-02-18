CREATE INDEX IF NOT EXISTS idx_daq_attempt_unanswered
ON diagnostic_attempt_questions (attempt_id, answered_at)
WHERE answered_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_question_attempts_attempt
ON question_attempts (attempt_id);

CREATE INDEX IF NOT EXISTS idx_coding_pending
ON coding_submissions (evaluation_status)
WHERE evaluation_status='pending';

CREATE INDEX IF NOT EXISTS idx_test_attempts_user
ON test_attempts (user_id);

CREATE INDEX IF NOT EXISTS idx_daq_attempt_order
ON diagnostic_attempt_questions (attempt_id, order_index);
