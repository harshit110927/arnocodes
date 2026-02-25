# Assessment Engine (Diagnostic Test) - Architecture Guide

This document defines how to build the **first-login diagnostic test engine** in clean phases.

## 1) Why this module is special
Test-taking is:
- Stateful
- Time-constrained
- Business-critical

So we treat assessment as a strict state machine instead of generic CRUD.

## 2) Existing schema support (already enough)
Current core tables already support the engine:
- `tests`
- `questions`
- `test_attempts`
- `question_attempts`

## 3) Required schema refinement
To guarantee deterministic question ordering, add `order_index` to `questions`.

```sql
ALTER TABLE questions
ADD COLUMN order_index INT;
```

## 4) Diagnostic test rule-set
For onboarding test (`tests.type = 'diagnostic'`):
- One-time mandatory attempt
- Dashboard locked until submitted
- No second attempt unless admin reset policy exists

## 5) Phased backend implementation

### Phase 1 — Static test load (read only)
Implement `GET /api/v1/tests/{testId}`:
- Return slides + questions + duration
- No write operations

### Phase 2 — Start attempt
Implement `POST /api/v1/tests/{testId}/start`:
- Verify diagnostic one-attempt policy
- Insert `test_attempts` with `status='in_progress'`, `started_at`

### Phase 3 — Next-question engine
Implement `GET /api/v1/test-attempts/{attemptId}/next-question`:
- Validate attempt status/time
- Return next question by `order_index`
- Never return `correct_option`

### Phase 4 — Submit answer
Implement `POST /api/v1/test-attempts/{attemptId}/answers`:
- Validate attempt active
- Validate question eligibility
- Enforce time limit server-side
- Write `question_attempts`

### Phase 5 — Submit attempt
Implement `POST /api/v1/test-attempts/{attemptId}/submit`:
- Validate completion/time-expiry
- Compute score
- Mark attempt submitted and lock

### Phase 6 — Dashboard gating
For onboarding users:
- If no submitted diagnostic attempt, block dashboard APIs with `403 TEST_REQUIRED`

## 6) Slide handling
Use `questions.question_type = 'slide'` for explainers:
- Slides are part of test payload
- Slides are not written to `question_attempts`
- Slides are not timed

## 7) Timer enforcement
Server is source of truth:
- Track per-question start timestamp (service cache/state)
- On answer submit: compute elapsed seconds
- If limit exceeded: reject or auto-mark incorrect (policy-driven)

## 8) Scope control (what to build later)
Do **not** mix these into first implementation:
- Leaderboards
- Anti-cheat AI
- Full analytics
- Coding engine

Ship MCQ diagnostic engine first.


## 9) Current backend status
Backend has moved to PostgreSQL infrastructure phase:
- DB connection uses `pgxpool`
- migrations + seed run at startup
- assessment repository skeleton exists
- assessment HTTP endpoints remain stubs until service+repository business logic is added

This keeps architecture clean while preparing production data flow.

## 10) Backend access control guard
Protected APIs (dashboard, learning content, leaderboards) must enforce diagnostic completion in backend.
If diagnostic is incomplete, return:
```json
{ "error": "DIAGNOSTIC_REQUIRED" }
```
with `403 Forbidden`.

Frontend lock icon is UX only; backend check is the actual authorization control.

## Implemented DB-backed Diagnostic Engine

- Source-of-truth is PostgreSQL (`test_attempts`, `question_attempts`, `diagnostic_attempt_questions`, `coding_submissions`, `diagnostic_topic_results`).
- Retake rule enforced server-side: maximum 2 attempts in the last 48 hours.
- Resume behavior: next question is computed from `answered_at` and `order_index`.
- Global expiry enforced from `test_attempts.expires_at`; expired attempts are marked `expired`.
- Coding evaluation is async through worker polling of `coding_submissions` where `evaluation_status='pending'`.
- Submission finalization computes per-topic score and updates `user_topic_progress` by keeping highest mastery.

### Timer policy

- MCQ question slot: 30 seconds.
- Coding question slot: 30 minutes.
- Attempt-level limit is persisted in `test_attempts.expires_at`; APIs reject actions once expired.

## Diagnostic Hardening Pass (Concurrency + Idempotency)

The backend now enforces stricter production safety in repository SQL:

- Attempt validation (`status='in_progress'` and `NOW() <= expires_at`) is enforced before next-question and all mutation paths.
- `CompleteDiagnosticAttempt` is idempotent and guarded with `UPDATE ... WHERE status='in_progress'`.
- Coding worker claim step uses transactional `FOR UPDATE SKIP LOCKED` and marks rows `processing` before returning them.
- Topic scoring counts only the latest coding submission per question (lateral latest row).
- State updates check `RowsAffected()` to prevent silent no-op races.

These changes harden multi-request and multi-worker concurrency without changing public API contracts.

## Dashboard Read Model + External Mastery Integration

- External solved questions are ingested into `external_question_activity` with idempotent uniqueness (`user_id + platform + platform_question_id`).
- `user_topic_progress` stores `external_solved_count`, `total_external_questions`, and persisted `mastery_score` (write-time update only).
- Unlock rule is transactional: a topic unlocks only when every prerequisite has mastery >= 80.
- Completion rule is transactional: status transitions to `completed` when mastery >= 80.
- Diagnostic finalization updates topic mastery from diagnostic percentages and re-runs unlock checks.
- Dashboard reads from precomputed tables: `dashboard_daily_snapshots`, `daily_activity`, `user_activity_feed`, `user_topic_progress`, and `events`.
