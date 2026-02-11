# API Roadmap (Based on Product Requirements)

## Current State
- Backend now exposes health plus placeholder endpoints for profile, dashboard, learning, assessment, platform sync, AI gateway, and internal recompute jobs.
- AI service exposes `GET /health`, `POST /query`, `POST /index` (RAG placeholder implementation).

## Required v1 API Surface

### Profile
- `GET /api/v1/profiles/me`
- `PATCH /api/v1/profiles/me`
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

## API Design Principles
- Idempotency keys for mutation endpoints where retries are likely (`/start`, `/submit`, `/trigger`).
- Strong request validation and explicit error codes.
- Standardized response envelope and trace IDs.
- AuthZ at user scope (`me`) and event scope.


## Validation and Integrity Notes
- `POST /api/v1/subtopics/{subtopicId}/complete` must be system-validated; completion should only be accepted when mastery threshold is met.
- In production, mastery should come from server-side scoring signals (question activity, correctness, time profile), not blind client trust.
