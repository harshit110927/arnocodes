# API Roadmap (Backend + Frontend Integration Contract)

## Current Backend State

- PostgreSQL infra is in place (connection, migrations, seed).
- Diagnostic subsystem is DB-backed with stateful endpoints.
- Protected resources are backend-gated until diagnostic completion.
- Local development identity is passed via `X-User-ID`.

---

## Versioning

All endpoints are versioned under:
- `/api/v1`

Do not consume unversioned feature routes from frontend.

---

## Core v1 Endpoint Surface

### Health
- `GET /api/v1/health`

### Profile
- `GET /api/v1/profiles/me`
- `PATCH /api/v1/profiles/me`
- `GET /api/v1/profiles/me/status`
- `GET /api/v1/profiles/me/platform-connections`
- `POST /api/v1/profiles/me/platform-connections`
- `DELETE /api/v1/profiles/me/platform-connections/{platform}`

### Dashboard (Protected)
- `GET /api/v1/dashboard/summary`
- `GET /api/v1/dashboard/heatmap?from=&to=`
- `GET /api/v1/dashboard/leaderboards?scope=&window=`

### Learning (Protected)
- `GET /api/v1/course/structure`
- `GET /api/v1/topics`
- `GET /api/v1/topics/{topicId}`
- `GET /api/v1/topics/{topicId}/unlock-status`
- `GET /api/v1/subtopics/{subtopicId}`
- `POST /api/v1/learning/questions/{questionId}/complete`
- `POST /api/v1/subtopics/{subtopicId}/complete`

### Diagnostic Assessment (Stateful)
- `POST /api/v1/diagnostic/start`
- `GET /api/v1/diagnostic/{attemptId}/next`
- `POST /api/v1/diagnostic/{attemptId}/answer`
- `POST /api/v1/diagnostic/{attemptId}/coding`
- `GET /api/v1/diagnostic/{attemptId}/status`
- `POST /api/v1/diagnostic/{attemptId}/submit`

### Platform Sync
- `POST /api/v1/platform-sync/trigger`
- `GET /api/v1/platform-sync/jobs/{jobId}`

### AI
- `POST /api/v1/ai/query`
- `POST /api/v1/ai/code-helper/step`
- `GET /api/v1/ai/usage`

### Internal (Dev/Ops only)
- `GET /api/v1/internal/api-catalog`
- `POST /api/v1/internal/api-smoke-check`
- `POST /api/v1/internal/recompute-dashboard`
- `POST /api/v1/internal/refresh-leaderboard`

---

## Authorization + Guard Rules

- `X-User-ID` required for user-scoped calls in local/dev.
- Protected routes return:
  - `403 {"error":"DIAGNOSTIC_REQUIRED"}` until diagnostic is submitted.
- Diagnostic retake guard returns:
  - `403 {"error":"DIAGNOSTIC_BLOCKED"}` when attempt limit is exceeded.

---

## Response and Error Contract

### Success envelope
```json
{
  "status": "ok",
  "message": "...",
  "data": {}
}
```

### Common error payload
```json
{
  "error": "ERROR_CODE"
}
```

Expected frontend error handling:
- `401 UNAUTHORIZED`
- `403 DIAGNOSTIC_REQUIRED`
- `403 DIAGNOSTIC_BLOCKED`
- `404 NOT_FOUND`
- `422 UNPROCESSABLE_ENTITY`

---

## Frontend Integration Notes

- Use `GET /api/v1/profiles/me/status` as the single source of truth for lock/unlock UX.
- Never depend on frontend-only lock logic for security.
- Use tokenized design system from `docs/FRONTEND_HELPER.md` for theme consistency.

---

## Local Validation

Use:
- `docs/LOCAL_TESTING_WITH_TEST_JS.md`
- `backend/test.js`
- `docs/API_LOCAL_TESTING.md`

to validate full diagnostic flow before frontend release.
