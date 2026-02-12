# Backend (Go)

## Description
Backend service built with Go, PostgreSQL-ready infrastructure, and clean modular architecture.

## Structure
- `cmd/api/` - application entry point
- `config/` - configuration management
- `internal/database/` - connection, migrations, seeding
- `internal/handlers/` - HTTP handlers and route registration
- `internal/assessment/` - assessment repository skeleton
- `internal/learning/` - learning repository skeleton
- `internal/dashboard/` - dashboard repository skeleton

## Running Locally
```bash
cd backend
go mod tidy
go run cmd/api/main.go
```

## Environment Variables
- `PORT` - server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - reserved for future Redis usage
- `ENVIRONMENT` - environment name

## Startup Lifecycle
On startup backend now does:
1. DB connection initialization via `pgxpool`
2. Idempotent migration runner (`schema_migrations`)
3. Idempotent seed runner (sample diagnostic/test/topic records)
4. Repository initialization and handler wiring

## API Endpoints (skeleton + guards)
### Health
- `GET /health`
- `GET /api/v1/health`

### Profile
- `GET /api/v1/profiles/me`
- `PATCH /api/v1/profiles/me`
- `GET /api/v1/profiles/me/status`
- `GET /api/v1/profiles/me/platform-connections`
- `POST /api/v1/profiles/me/platform-connections`
- `DELETE /api/v1/profiles/me/platform-connections/{platform}`

### Protected resources
Dashboard and learning APIs are backend guarded. If diagnostic is incomplete they return:
```json
{ "error": "DIAGNOSTIC_REQUIRED" }
```
with HTTP `403`.

### Assessment (repository-backed stubs for business logic phase)
- `GET /api/v1/tests/{testId}`
- `POST /api/v1/tests/{testId}/start`
- `GET /api/v1/tests/{testId}/session`
- `POST /api/v1/test-attempts/{attemptId}/answers`
- `POST /api/v1/test-attempts/{attemptId}/submit`
- `GET /api/v1/test-attempts/{attemptId}/result`
- `GET /api/v1/test-attempts/{attemptId}/next-question`
- `POST /api/v1/test-attempts/{attemptId}/expire`
- `POST /api/v1/test-attempts/{attemptId}/resume`

### Internal utilities
- `GET /api/v1/internal/api-catalog`
- `POST /api/v1/internal/api-smoke-check`
- `POST /api/v1/internal/recompute-dashboard`
- `POST /api/v1/internal/refresh-leaderboard`



## Supabase Local Testing
For full local clone + Supabase setup and troubleshooting, see `../docs/LOCAL_SUPABASE_SETUP.md`.
