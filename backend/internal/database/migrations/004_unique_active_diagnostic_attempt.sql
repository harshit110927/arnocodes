CREATE UNIQUE INDEX IF NOT EXISTS unique_active_diagnostic_attempt
ON test_attempts (user_id, test_id)
WHERE status IN ('started', 'in_progress');
