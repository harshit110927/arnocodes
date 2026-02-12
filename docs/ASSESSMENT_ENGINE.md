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
The backend now includes an in-memory assessment engine for local development (no DB dependency yet):
- topic-filtered test fetch
- diagnostic attempt start with one-attempt policy
- next-question navigation
- answer submission with order enforcement
- submit + score computation
- result fetch, expire, resume endpoints

This lets frontend test complete onboarding assessment flow before DB wiring.

## 10) Backend access control guard
Protected APIs (dashboard, learning content, leaderboards) must enforce diagnostic completion in backend.
If diagnostic is incomplete, return:
```json
{ "error": "DIAGNOSTIC_REQUIRED" }
```
with `403 Forbidden`.

Frontend lock icon is UX only; backend check is the actual authorization control.
