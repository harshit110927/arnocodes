CREATE TABLE IF NOT EXISTS diagnostic_attempt_questions (
  attempt_id UUID NOT NULL REFERENCES test_attempts(id) ON DELETE CASCADE,
  question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  topic_id UUID NOT NULL REFERENCES topics(id),
  order_index INT NOT NULL,
  question_type TEXT NOT NULL,
  allotted_seconds INT NOT NULL,
  answered_at TIMESTAMP NULL,
  PRIMARY KEY (attempt_id, question_id)
);

CREATE TABLE IF NOT EXISTS coding_submissions (
  id UUID PRIMARY KEY,
  attempt_id UUID NOT NULL REFERENCES test_attempts(id) ON DELETE CASCADE,
  question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  code TEXT NOT NULL,
  language TEXT NOT NULL,
  evaluation_status TEXT NOT NULL,
  score FLOAT NULL,
  created_at TIMESTAMP NOT NULL,
  evaluated_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS diagnostic_topic_results (
  attempt_id UUID NOT NULL REFERENCES test_attempts(id) ON DELETE CASCADE,
  topic_id UUID NOT NULL REFERENCES topics(id),
  score INT NOT NULL,
  max_score INT NOT NULL,
  percentage FLOAT NOT NULL,
  passed BOOLEAN NOT NULL,
  created_at TIMESTAMP NOT NULL,
  PRIMARY KEY (attempt_id, topic_id)
);

CREATE INDEX IF NOT EXISTS idx_diagnostic_attempt_questions_attempt_order
  ON diagnostic_attempt_questions(attempt_id, order_index);

CREATE INDEX IF NOT EXISTS idx_coding_submissions_eval_status
  ON coding_submissions(evaluation_status);

CREATE INDEX IF NOT EXISTS idx_diagnostic_topic_results_attempt_topic
  ON diagnostic_topic_results(attempt_id, topic_id);
