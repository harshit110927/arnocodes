# API Roadmap (Based on Product Requirements)

## Current State
- Backend now has PostgreSQL infrastructure wiring (connection, migrations, seed) and repository skeletons.
- Assessment endpoints are repository-backed stubs in this phase (business logic pending).
- AI service exposes `GET /health`, `POST /query`, `POST /index` (RAG placeholder implementation).

## Required v1 API Surface

### Profile
- `GET /api/v1/profiles/me`
- `PATCH /api/v1/profiles/me`
- `GET /api/v1/profiles/me/status`
- `POST /api/v1/profiles/me/platform-connections`
- `DELETE /api/v1/profiles/me/platform-connections/{platform}`

### Dashboard
- `GET /api/v1/dashboard/summary`
- `GET /api/v1/dashboard/heatmap?from=&to=`
- `GET /api/v1/dashboard/leaderboards?scope=&window=`

### Learning
- `GET /api/v1/course/structure` (optional but recommended single-call DAG read model)
- `GET /api/v1/topics`
- `GET /api/v1/topics/{topicId}`
- `GET /api/v1/topics/{topicId}/unlock-status`
- `GET /api/v1/subtopics/{subtopicId}`
- `POST /api/v1/learning/questions/{questionId}/complete`
- `POST /api/v1/subtopics/{subtopicId}/complete`

### Assessment
- `GET /api/v1/tests/{testId}`
- `POST /api/v1/tests/{testId}/start`
- `GET /api/v1/tests/{testId}/session`
- `POST /api/v1/test-attempts/{attemptId}/answers`
- `POST /api/v1/test-attempts/{attemptId}/submit`
- `GET /api/v1/test-attempts/{attemptId}/result`
- `GET /api/v1/test-attempts/{attemptId}/next-question`
- `POST /api/v1/test-attempts/{attemptId}/expire`
- `POST /api/v1/test-attempts/{attemptId}/resume`

### Platform Sync
- `GET /api/v1/profiles/me/platform-connections`
- `POST /api/v1/platform-sync/trigger`
- `GET /api/v1/platform-sync/jobs/{jobId}`

### AI
- `POST /api/v1/ai/query`
- `POST /api/v1/ai/code-helper/step`
- `GET /api/v1/ai/usage`

### Internal (Worker/Cron only)
- `POST /api/v1/internal/recompute-dashboard`
- `POST /api/v1/internal/refresh-leaderboard`
- `GET /api/v1/internal/api-catalog`
- `POST /api/v1/internal/api-smoke-check`

## API Design Principles
- Idempotency keys for mutation endpoints where retries are likely (`/start`, `/submit`, `/trigger`).
- Strong request validation and explicit error codes.
- Standardized response envelope and trace IDs.
- AuthZ at user scope (`me`) and event scope.


## Validation and Integrity Notes
- `POST /api/v1/subtopics/{subtopicId}/complete` must be system-validated; completion should only be accepted when mastery threshold is met.
- In production, mastery should come from server-side scoring signals (question activity, correctness, time profile), not blind client trust.


## Diagnostic Test Engine Rollout
1. `GET /api/v1/tests/{testId}`
2. `POST /api/v1/tests/{testId}/start`
3. `GET /api/v1/test-attempts/{attemptId}/next-question`
4. `POST /api/v1/test-attempts/{attemptId}/answers`
5. `POST /api/v1/test-attempts/{attemptId}/submit`
6. Dashboard gate (`403 TEST_REQUIRED`) until diagnostic submit


## Assessment request examples (local)
- Start attempt: `POST /api/v1/tests/diagnostic-1/start` with `{ "topics_known": ["arrays", "strings"] }`.
- Submit answer: `POST /api/v1/test-attempts/{attemptId}/answers` with `{ "question_id": "q-1", "selected_option": 2 }`.
- Session restore: `GET /api/v1/tests/diagnostic-1/session?attempt_id={attemptId}`.


## Protected resources policy
- Dashboard and learning resources are backend-protected until diagnostic test is submitted.
- If not completed, APIs return `403` with `{ "error": "DIAGNOSTIC_REQUIRED" }`.
- Frontend should read `GET /api/v1/profiles/me/status` and render lock states accordingly.


## Infrastructure-first rollout note
- Phase complete: DB connection + migration + seed + repository skeletons.
- Next phase: move assessment workflow logic from stubs into repository/service layer using raw SQL with transactions.

## Diagnostic APIs (DB-backed)

- `POST /api/v1/diagnostic/start`
  - Starts diagnostic attempt (validates prerequisite-respecting topic selection and retake limit).
  - `403 {"error":"DIAGNOSTIC_BLOCKED"}` if >2 attempts in 48h.
- `GET /api/v1/diagnostic/{attemptID}/next`
  - Returns next safe question payload (never includes correct answer).
- `POST /api/v1/diagnostic/{attemptID}/answer`
  - Submits MCQ or coding answer payload.
- `POST /api/v1/diagnostic/{attemptID}/coding`
  - Alias for coding answer submit (same schema as `/answer`).
- `GET /api/v1/diagnostic/{attemptID}/status`
  - Returns attempt status, progress, total allowed seconds.
- `POST /api/v1/diagnostic/{attemptID}/submit`
  - Finalizes attempt, computes topic-wise results and updates topic progress using highest-score rule.

All endpoints require `X-User-ID` and enforce attempt ownership.
