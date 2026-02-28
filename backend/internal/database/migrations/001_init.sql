CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS auth;
CREATE TABLE IF NOT EXISTS auth.users (
  id UUID PRIMARY KEY
);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'test_attempt_status') THEN
    CREATE TYPE test_attempt_status AS ENUM ('started','in_progress','submitted','auto_submitted','expired','abandoned');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'question_attempt_state') THEN
    CREATE TYPE question_attempt_state AS ENUM ('not_visited','visited','answered','skipped','marked_for_review');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'learning_progress_status') THEN
    CREATE TYPE learning_progress_status AS ENUM ('not_started','in_progress','completed','mastered');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'event_phase') THEN
    CREATE TYPE event_phase AS ENUM ('registration','active','evaluation','completed','archived');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'event_participant_status') THEN
    CREATE TYPE event_participant_status AS ENUM ('registered','active','disqualified','completed');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'sync_job_status') THEN
    CREATE TYPE sync_job_status AS ENUM ('queued','running','succeeded','failed','rate_limited');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id),
  full_name TEXT,
  college TEXT,
  graduation_year INT
);

CREATE TABLE IF NOT EXISTS topics (
  id UUID PRIMARY KEY,
  name TEXT
);

CREATE TABLE IF NOT EXISTS topic_prerequisites (
  topic_id UUID REFERENCES topics(id),
  prerequisite_id UUID REFERENCES topics(id),
  PRIMARY KEY (topic_id, prerequisite_id)
);

CREATE TABLE IF NOT EXISTS subtopics (
  id UUID PRIMARY KEY,
  topic_id UUID REFERENCES topics(id),
  title TEXT,
  order_index INT
);

CREATE TABLE IF NOT EXISTS user_subtopic_progress (
  user_id UUID REFERENCES profiles(id),
  subtopic_id UUID REFERENCES subtopics(id),
  status learning_progress_status,
  mastery_score FLOAT,
  completed_at TIMESTAMP,
  PRIMARY KEY (user_id, subtopic_id)
);

CREATE TABLE IF NOT EXISTS user_topic_progress (
  user_id UUID REFERENCES profiles(id),
  topic_id UUID REFERENCES topics(id),
  status learning_progress_status,
  mastery_score FLOAT,
  completed_at TIMESTAMP,
  PRIMARY KEY (user_id, topic_id)
);

CREATE TABLE IF NOT EXISTS learning_questions (
  id UUID PRIMARY KEY,
  topic_id UUID REFERENCES topics(id),
  source TEXT,
  difficulty TEXT,
  link TEXT
);

CREATE TABLE IF NOT EXISTS user_learning_question_activity (
  user_id UUID REFERENCES profiles(id),
  question_id UUID REFERENCES learning_questions(id),
  solved BOOLEAN,
  solved_at DATE,
  time_taken_minutes INT,
  PRIMARY KEY (user_id, question_id)
);

CREATE TABLE IF NOT EXISTS tests (
  id UUID PRIMARY KEY,
  type TEXT,
  duration_minutes INT,
  total_marks INT
);

CREATE TABLE IF NOT EXISTS questions (
  id UUID PRIMARY KEY,
  test_id UUID REFERENCES tests(id),
  question_type TEXT,
  content TEXT,
  options JSONB,
  correct_option INT,
  marks INT,
  order_index INT
);

CREATE TABLE IF NOT EXISTS test_attempts (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES profiles(id),
  test_id UUID REFERENCES tests(id),
  score INT,
  status test_attempt_status,
  started_at TIMESTAMP,
  expires_at TIMESTAMP,
  submitted_at TIMESTAMP,
  evaluation_version TEXT
);

CREATE TABLE IF NOT EXISTS question_attempts (
  attempt_id UUID REFERENCES test_attempts(id),
  question_id UUID REFERENCES questions(id),
  selected_option INT,
  time_taken_seconds INT,
  is_correct BOOLEAN,
  state question_attempt_state,
  answered_at TIMESTAMP,
  is_marked_for_review BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (attempt_id, question_id)
);

CREATE TABLE IF NOT EXISTS daily_activity (
  user_id UUID REFERENCES profiles(id),
  activity_date DATE,
  questions_solved INT,
  PRIMARY KEY (user_id, activity_date)
);

CREATE TABLE IF NOT EXISTS platform_activity (
  user_id UUID REFERENCES profiles(id),
  activity_date DATE,
  topic_id UUID REFERENCES topics(id),
  difficulty TEXT,
  time_spent_minutes INT,
  questions_solved INT,
  PRIMARY KEY (user_id, activity_date, topic_id)
);

CREATE TABLE IF NOT EXISTS ai_usage (
  user_id UUID REFERENCES profiles(id),
  date DATE,
  slm_requests INT,
  llm_requests INT,
  PRIMARY KEY (user_id, date)
);

CREATE TABLE IF NOT EXISTS ai_query_gists (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES profiles(id),
  topic_id UUID REFERENCES topics(id),
  gist TEXT,
  created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS events (
  id UUID PRIMARY KEY,
  name TEXT,
  phase event_phase,
  start_date DATE,
  end_date DATE
);

CREATE TABLE IF NOT EXISTS event_participants (
  event_id UUID REFERENCES events(id),
  user_id UUID REFERENCES profiles(id),
  status event_participant_status,
  final_score INT,
  rank INT,
  PRIMARY KEY (event_id, user_id)
);

CREATE TABLE IF NOT EXISTS leaderboards (
  id UUID PRIMARY KEY,
  name TEXT,
  leaderboard_type TEXT,
  scope_id UUID,
  window TEXT,
  metric TEXT,
  starts_at TIMESTAMP,
  ends_at TIMESTAMP,
  created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS leaderboard_entries (
  leaderboard_id UUID REFERENCES leaderboards(id),
  user_id UUID REFERENCES profiles(id),
  rank INT,
  score FLOAT,
  computed_at TIMESTAMP,
  PRIMARY KEY (leaderboard_id, user_id)
);

CREATE TABLE IF NOT EXISTS dashboard_daily_snapshots (
  user_id UUID REFERENCES profiles(id),
  snapshot_date DATE,
  streak_count INT,
  questions_solved INT,
  mastery_score FLOAT,
  topics_completed INT,
  last_activity_at TIMESTAMP,
  computed_at TIMESTAMP,
  PRIMARY KEY (user_id, snapshot_date)
);

CREATE TABLE IF NOT EXISTS platform_connections (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES profiles(id),
  platform TEXT,
  platform_handle TEXT,
  access_token_ref TEXT,
  status TEXT,
  connected_at TIMESTAMP,
  last_validated_at TIMESTAMP,
  UNIQUE (user_id, platform)
);

CREATE TABLE IF NOT EXISTS platform_sync_jobs (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES profiles(id),
  connection_id UUID REFERENCES platform_connections(id),
  status sync_job_status,
  trigger_source TEXT,
  started_at TIMESTAMP,
  finished_at TIMESTAMP,
  error_message TEXT
);

CREATE TABLE IF NOT EXISTS ai_code_helper_sessions (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES profiles(id),
  topic_id UUID REFERENCES topics(id),
  question_id UUID REFERENCES learning_questions(id),
  step_index INT,
  max_steps INT,
  status TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subtopics_topic_id ON subtopics(topic_id);
CREATE INDEX IF NOT EXISTS idx_questions_test_id ON questions(test_id);
CREATE INDEX IF NOT EXISTS idx_questions_test_order ON questions(test_id, order_index);
CREATE INDEX IF NOT EXISTS idx_test_attempts_user_test ON test_attempts(user_id, test_id);
CREATE INDEX IF NOT EXISTS idx_test_attempts_user_submitted ON test_attempts(user_id, submitted_at);
CREATE INDEX IF NOT EXISTS idx_question_attempts_question ON question_attempts(question_id);
CREATE INDEX IF NOT EXISTS idx_platform_activity_user_date ON platform_activity(user_id, activity_date);
CREATE INDEX IF NOT EXISTS idx_daily_activity_user_date ON daily_activity(user_id, activity_date);
CREATE INDEX IF NOT EXISTS idx_leaderboard_entries_rank ON leaderboard_entries(leaderboard_id, rank);
CREATE INDEX IF NOT EXISTS idx_leaderboard_entries_score ON leaderboard_entries(leaderboard_id, score);
CREATE INDEX IF NOT EXISTS idx_platform_sync_jobs_user_status ON platform_sync_jobs(user_id, status);


fmt.Println("DB URL:", os.Getenv("DATABASE_URL"))
fmt.Println("DB HOST:", os.Getenv("DB_HOST"))
fmt.Println("DB USER:", os.Getenv("DB_USER"))