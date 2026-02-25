ALTER TABLE user_topic_progress
  ADD COLUMN IF NOT EXISTS external_solved_count INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS total_external_questions INT NOT NULL DEFAULT 0;

ALTER TABLE user_topic_progress
  ALTER COLUMN mastery_score SET DEFAULT 0;

CREATE TABLE IF NOT EXISTS external_question_activity (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES profiles(id),
  platform TEXT NOT NULL,
  platform_question_id TEXT NOT NULL,
  title TEXT,
  topic_id UUID NULL REFERENCES topics(id),
  difficulty TEXT,
  solved_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, platform, platform_question_id)
);

CREATE INDEX IF NOT EXISTS idx_external_question_activity_user_solved_at
ON external_question_activity (user_id, solved_at DESC);

CREATE TABLE IF NOT EXISTS user_activity_feed (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES profiles(id),
  source TEXT NOT NULL,
  title TEXT NOT NULL,
  topic_id UUID NULL REFERENCES topics(id),
  difficulty TEXT,
  link TEXT,
  solved_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_activity_feed_user_solved_at
ON user_activity_feed (user_id, solved_at DESC);
