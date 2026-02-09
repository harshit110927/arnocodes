# Database Schema

## Core Tables (Learning + Assessment)

### 1. `profiles`
```sql
profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id),
  full_name TEXT,
  college TEXT,
  graduation_year INT
)
```
Stores application-level user metadata separate from auth identity.

### 2. `topics`
```sql
topics (
  id UUID PRIMARY KEY,
  name TEXT
)
```
Defines DAG nodes (DSA topics).

### 3. `topic_prerequisites`
```sql
topic_prerequisites (
  topic_id UUID REFERENCES topics(id),
  prerequisite_id UUID REFERENCES topics(id),
  PRIMARY KEY (topic_id, prerequisite_id)
)
```
Directed edges between topics.

### 4. `subtopics`
```sql
subtopics (
  id UUID PRIMARY KEY,
  topic_id UUID REFERENCES topics(id),
  title TEXT,
  order_index INT
)
```
Ordered learning units inside a topic.

### 5. `user_subtopic_progress`
```sql
user_subtopic_progress (
  user_id UUID REFERENCES profiles(id),
  subtopic_id UUID REFERENCES subtopics(id),
  status TEXT,
  mastery_score FLOAT,
  completed_at TIMESTAMP,
  PRIMARY KEY (user_id, subtopic_id)
)
```
Tracks a user’s progress and mastery at the subtopic level.

### 6. `user_topic_progress`
```sql
user_topic_progress (
  user_id UUID REFERENCES profiles(id),
  topic_id UUID REFERENCES topics(id),
  status TEXT,
  mastery_score FLOAT,
  completed_at TIMESTAMP,
  PRIMARY KEY (user_id, topic_id)
)
```
Aggregated topic-level mastery derived from subtopic progress.

### 7. `learning_questions`
```sql
learning_questions (
  id UUID PRIMARY KEY,
  topic_id UUID REFERENCES topics(id),
  source TEXT,
  difficulty TEXT,
  link TEXT
)
```
Canonical representation of learning problems.

### 8. `user_learning_question_activity`
```sql
user_learning_question_activity (
  user_id UUID REFERENCES profiles(id),
  question_id UUID REFERENCES learning_questions(id),
  solved BOOLEAN,
  solved_at DATE,
  time_taken_minutes INT,
  PRIMARY KEY (user_id, question_id)
)
```
Tracks problem-solving activity for learning questions.

### 9. `tests`
```sql
tests (
  id UUID PRIMARY KEY,
  type TEXT,
  duration_minutes INT,
  total_marks INT
)
```
Defines test containers.

### 10. `questions`
```sql
questions (
  id UUID PRIMARY KEY,
  test_id UUID REFERENCES tests(id),
  question_type TEXT,
  content TEXT,
  options JSONB,
  correct_option INT,
  marks INT
)
```
Assessment questions for tests.

### 11. `test_attempts`
```sql
test_attempts (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES profiles(id),
  test_id UUID REFERENCES tests(id),
  score INT,
  submitted_at TIMESTAMP
)
```
Tracks user test attempts.

### 12. `question_attempts`
```sql
question_attempts (
  attempt_id UUID REFERENCES test_attempts(id),
  question_id UUID REFERENCES questions(id),
  selected_option INT,
  time_taken_seconds INT,
  is_correct BOOLEAN,
  PRIMARY KEY (attempt_id, question_id)
)
```
Per-question behavior during tests.

### 13. `daily_activity`
```sql
daily_activity (
  user_id UUID REFERENCES profiles(id),
  activity_date DATE,
  questions_solved INT,
  PRIMARY KEY (user_id, activity_date)
)
```
Daily activity summary (heatmap, streaks).

### 14. `platform_activity`
```sql
platform_activity (
  user_id UUID REFERENCES profiles(id),
  activity_date DATE,
  topic_id UUID REFERENCES topics(id),
  difficulty TEXT,
  time_spent_minutes INT,
  questions_solved INT,
  PRIMARY KEY (user_id, activity_date, topic_id)
)
```
Cross-platform learning activity time series.

### 15. `ai_usage`
```sql
ai_usage (
  user_id UUID REFERENCES profiles(id),
  date DATE,
  slm_requests INT,
  llm_requests INT,
  PRIMARY KEY (user_id, date)
)
```
Tracks AI usage and enforces daily limits.

### 16. `ai_query_gists`
```sql
ai_query_gists (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES profiles(id),
  topic_id UUID REFERENCES topics(id),
  gist TEXT,
  created_at TIMESTAMP
)
```
Condensed summaries of AI interactions.

### 17. `events`
```sql
events (
  id UUID PRIMARY KEY,
  name TEXT,
  phase TEXT,
  start_date DATE,
  end_date DATE
)
```
Competitions or college pilot events.

### 18. `event_participants`
```sql
event_participants (
  event_id UUID REFERENCES events(id),
  user_id UUID REFERENCES profiles(id),
  status TEXT,
  final_score INT,
  rank INT,
  PRIMARY KEY (event_id, user_id)
)
```
Tracks event participation and final outcomes.

## Leaderboard (Recommended Additions)

Leaderboards are best modeled as **materialized ranking snapshots**. This keeps read paths fast and prevents expensive rank recomputation for every dashboard load.

### 19. `leaderboards`
```sql
leaderboards (
  id UUID PRIMARY KEY,
  name TEXT,
  leaderboard_type TEXT, -- global | college | event | topic
  scope_id UUID NULL, -- college_id or event_id or topic_id when relevant
  window TEXT, -- daily | weekly | monthly | all_time
  metric TEXT, -- questions_solved | mastery_score | test_score
  starts_at TIMESTAMP NULL,
  ends_at TIMESTAMP NULL,
  created_at TIMESTAMP
)
```
Defines leaderboard configurations. `scope_id` ties it to a college, event, or topic when needed.

### 20. `leaderboard_entries`
```sql
leaderboard_entries (
  leaderboard_id UUID REFERENCES leaderboards(id),
  user_id UUID REFERENCES profiles(id),
  rank INT,
  score FLOAT,
  computed_at TIMESTAMP,
  PRIMARY KEY (leaderboard_id, user_id)
)
```
Stores rank snapshots for fast dashboard reads.

### Notes
- **Derived data:** Scores are computed from `daily_activity`, `platform_activity`, `user_topic_progress`, and `test_attempts`.
- **Indexing:** Add indexes on `(leaderboard_id, rank)` and `(leaderboard_id, score)` for paging.
- **Refresh strategy:** Update leaderboards on a schedule (cron) or on event boundaries (e.g., end of weekly cycle).
- **Fairness:** For cross-platform comparisons, normalize by difficulty or time spent if needed.

