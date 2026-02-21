CREATE TABLE IF NOT EXISTS coding_question_test_cases (
  id UUID PRIMARY KEY,
  question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  input TEXT NOT NULL,
  expected_output TEXT NOT NULL,
  is_sample BOOLEAN NOT NULL DEFAULT FALSE,
  weight FLOAT NOT NULL DEFAULT 1,
  order_index INT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coding_question_test_cases_question_id
  ON coding_question_test_cases(question_id);
