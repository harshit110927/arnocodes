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
- `GET /profiles/me`
- `PATCH /profiles/me`
- `GET /profiles/me/platform-connections`
- `POST /profiles/me/platform-connections`
- `DELETE /profiles/me/platform-connections/{platform}`

### Dashboard
- `GET /dashboard/summary`
- `GET /dashboard/heatmap`
- `GET /dashboard/leaderboards`

### Learning
- `GET /topics`
- `GET /topics/{topicId}`
- `GET /topics/{topicId}/unlock-status`
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

### Platform sync
- `POST /platform-sync/trigger`
- `GET /platform-sync/jobs/{jobId}`

### AI gateway
- `POST /ai/query`
- `POST /ai/code-helper/step`
- `GET /ai/usage`

### Internal (cron/worker)
- `POST /internal/recompute-dashboard`
- `POST /internal/refresh-leaderboard`

> Note: most endpoints currently return structured placeholder responses and are ready to be wired to services/repositories.
