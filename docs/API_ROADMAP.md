# API Roadmap (Based on Product Requirements)

## Current State
- Backend currently exposes only `GET /health`.
- AI service exposes `GET /health`, `POST /query`, `POST /index`.

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
- `POST /attempts/{attemptId}/answers`
- `POST /attempts/{attemptId}/submit`
- `GET /attempts/{attemptId}/result`

### Platform Sync
- `POST /platform-sync/trigger`
- `GET /platform-sync/jobs/{jobId}`

### AI
- `POST /ai/query`
- `POST /ai/code-helper/step`

## API Design Principles
- Idempotency keys for mutation endpoints where retries are likely (`/start`, `/submit`, `/trigger`).
- Strong request validation and explicit error codes.
- Standardized response envelope and trace IDs.
- AuthZ at user scope (`me`) and event scope.

