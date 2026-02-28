BEGIN;

ALTER TABLE question_attempts
ADD COLUMN IF NOT EXISTS user_id UUID;

UPDATE question_attempts qa
SET user_id = ta.user_id
FROM test_attempts ta
WHERE qa.attempt_id = ta.id
  AND qa.user_id IS NULL;

ALTER TABLE question_attempts
ALTER COLUMN user_id SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.table_constraints
    WHERE constraint_name = 'fk_question_attempts_user'
  ) THEN
    ALTER TABLE question_attempts
    ADD CONSTRAINT fk_question_attempts_user
    FOREIGN KEY (user_id)
    REFERENCES profiles(id)
    ON DELETE CASCADE;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_question_attempts_user_id
ON question_attempts(user_id);

ALTER TABLE IF EXISTS user_topic_progress ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS user_subtopic_progress ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS test_attempts ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS question_attempts ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS external_question_activity ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS dashboard_daily_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS daily_activity ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS platform_activity ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS ai_usage ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS ai_query_gists ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS user_topic_progress_own_data ON user_topic_progress;
CREATE POLICY user_topic_progress_own_data ON user_topic_progress
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS user_subtopic_progress_own_data ON user_subtopic_progress;
CREATE POLICY user_subtopic_progress_own_data ON user_subtopic_progress
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS test_attempts_own_data ON test_attempts;
CREATE POLICY test_attempts_own_data ON test_attempts
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS question_attempts_own_data ON question_attempts;
CREATE POLICY question_attempts_own_data ON question_attempts
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS external_question_activity_own_data ON external_question_activity;
CREATE POLICY external_question_activity_own_data ON external_question_activity
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS dashboard_daily_snapshots_own_data ON dashboard_daily_snapshots;
CREATE POLICY dashboard_daily_snapshots_own_data ON dashboard_daily_snapshots
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS daily_activity_own_data ON daily_activity;
CREATE POLICY daily_activity_own_data ON daily_activity
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS platform_activity_own_data ON platform_activity;
CREATE POLICY platform_activity_own_data ON platform_activity
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS ai_usage_own_data ON ai_usage;
CREATE POLICY ai_usage_own_data ON ai_usage
FOR ALL USING (user_id = auth.uid());

DROP POLICY IF EXISTS ai_query_gists_own_data ON ai_query_gists;
CREATE POLICY ai_query_gists_own_data ON ai_query_gists
FOR ALL USING (user_id = auth.uid());

COMMIT;