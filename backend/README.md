# Backend (Go)

## Description
Backend service built with Go, featuring a clean modular architecture.

## Structure
- `cmd/api/` - Application entry point
- `internal/` - Internal packages
  - `handlers/` - HTTP handlers
  - `database/` - Database connections (PostgreSQL and Redis placeholders)
- `config/` - Configuration management

## Running Locally

```bash
# Install dependencies
go mod download

# Run the server
go run cmd/api/main.go
```

## Environment Variables
Copy `.env.example` to `.env` and update the values:
- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `ENVIRONMENT` - Environment name (development/production)

## Endpoints (current placeholders + health)

### Health
- `GET /health`

### Profile
- `GET /api/v1/profiles/me`
- `PATCH /api/v1/profiles/me`
- `GET /api/v1/profiles/me/status`
- `GET /api/v1/profiles/me/platform-connections`
- `POST /api/v1/profiles/me/platform-connections`
- `DELETE /api/v1/profiles/me/platform-connections/{platform}`

### Dashboard
- `GET /api/v1/dashboard/summary`
- `GET /api/v1/dashboard/heatmap`
- `GET /api/v1/dashboard/leaderboards`

### Learning
- `GET /api/v1/course/structure`
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

### Platform sync
- `POST /api/v1/platform-sync/trigger`
- `GET /api/v1/platform-sync/jobs/{jobId}`

### AI gateway
- `POST /api/v1/ai/query`
- `POST /api/v1/ai/code-helper/step`
- `GET /api/v1/ai/usage`

### Internal (cron/worker)
- `POST /api/v1/internal/recompute-dashboard`
- `POST /api/v1/internal/refresh-leaderboard`
- `GET /api/v1/internal/api-catalog`
- `POST /api/v1/internal/api-smoke-check`

> Note: most endpoints currently return structured placeholder responses and are ready to be wired to services/repositories.

### Validation policy note
- `POST /api/v1/subtopics/{subtopicId}/complete` is threshold-validated and should not allow blind completion writes.


## API Skeleton Check (No DB Required)
Use the skeleton verification endpoint to validate all registered APIs before wiring database/services:

```bash
curl -X POST http://localhost:8080/api/v1/internal/api-smoke-check
```

The response includes total/passed/failed and endpoint-level status checks.


### Assessment implementation note
- Assessment endpoints are now backed by an in-memory service (`internal/assessment`) for local development.
- Use `X-User-ID` header to simulate per-user attempt state during local testing.
- `POST /api/v1/tests/diagnostic-1/start` accepts `{"topics_known": ["arrays", "strings"]}` and enforces diagnostic one-attempt policy.

- Protected dashboard/learning endpoints return `403` with `DIAGNOSTIC_REQUIRED` until the diagnostic test is submitted for that user.
