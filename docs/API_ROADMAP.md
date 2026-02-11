# API Roadmap (Based on Product Requirements)

## Current State
- Backend now exposes health plus placeholder endpoints for profile, dashboard, learning, assessment, platform sync, AI gateway, and internal recompute jobs.
- AI service exposes `GET /health`, `POST /query`, `POST /index` (RAG placeholder implementation).

## Required v1 API Surface

### Profile
- `GET /profiles/me`
- `PATCH /profiles/me`
- `POST /profiles/me/platform-connections`
- `DELETE /profiles/me/platform-connections/{platform}`

### Dashboard
- `GET /dashboard/summary`
- `GET /dashboard/heatmap?from=&to=`
- `GET /dashboard/leaderboards?scope=&window=`

### Learning
- `GET /topics`
- `GET /topics/{topicId}`
- `GET /subtopics/{subtopicId}`
- `POST /learning/questions/{questionId}/complete`
- `POST /subtopics/{subtopicId}/complete`

### Assessment
- `GET /tests/{testId}`
- `POST /tests/{testId}/start`
- `GET /tests/{testId}/session`
- `POST /test-attempts/{attemptId}/answers`
- `POST /test-attempts/{attemptId}/submit`
- `GET /test-attempts/{attemptId}/result`
- `GET /test-attempts/{attemptId}/next-question`
- `POST /test-attempts/{attemptId}/expire`
- `POST /test-attempts/{attemptId}/resume`

### Platform Sync
- `GET /profiles/me/platform-connections`
- `POST /platform-sync/trigger`
- `GET /platform-sync/jobs/{jobId}`

### AI
- `POST /ai/query`
- `POST /ai/code-helper/step`
- `GET /ai/usage`

### Internal (Worker/Cron only)
- `POST /internal/recompute-dashboard`
- `POST /internal/refresh-leaderboard`

## API Design Principles
- Idempotency keys for mutation endpoints where retries are likely (`/start`, `/submit`, `/trigger`).
- Strong request validation and explicit error codes.
- Standardized response envelope and trace IDs.
- AuthZ at user scope (`me`) and event scope.

