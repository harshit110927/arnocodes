ALTER TABLE user_topic_progress
  ADD COLUMN IF NOT EXISTS diagnostic_mastery FLOAT NOT NULL DEFAULT 0;

ALTER TABLE user_topic_progress
  ALTER COLUMN mastery_score SET DEFAULT 0;

DELETE FROM external_question_activity
WHERE topic_id IS NULL;

ALTER TABLE external_question_activity
  ALTER COLUMN topic_id SET NOT NULL;
