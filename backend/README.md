# Backend (Go)

## Description
Backend service built with Go, PostgreSQL-ready infrastructure, and clean modular architecture.

## Structure
- `cmd/api/` - application entry point
- `config/` - configuration management
- `internal/database/` - connection, migrations, seeding
- `internal/handlers/` - HTTP handlers and route registration
- `internal/assessment/` - assessment repository skeleton
- `internal/course/` - read-only course access (unlock DAG + gated reads)
- `internal/learning/activity/` - learning activity write paths
- `internal/dashboard/` - dashboard repository skeleton


## Domain Boundaries
- `course` is read-only: computes derived unlock state from persisted mastery + prerequisites and never writes unlock flags.
- `learning/activity` owns learning question completion write paths used by async evaluation and mastery flows.
- `assessment` owns diagnostic state; `course` only gates reads on diagnostic completion.
- `ide` owns coding submission evaluation pipeline; it does not compute unlocks.
- Unlock remains derived state (not persisted), so external-sync mastery updates are reflected automatically on next course read.

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
Dashboard and course APIs are backend guarded. If diagnostic is incomplete they return:
```json
{ "error": "DIAGNOSTIC_REQUIRED" }
```
with HTTP `403`.

### Assessment (diagnostic APIs)
- `POST /api/v1/diagnostic/start`
- `GET /api/v1/diagnostic/{attemptId}/next`
- `POST /api/v1/diagnostic/{attemptId}/answer`
- `POST /api/v1/diagnostic/{attemptId}/coding`
- `GET /api/v1/diagnostic/{attemptId}/status`
- `POST /api/v1/diagnostic/{attemptId}/submit`

### Internal utilities
- `GET /api/v1/internal/api-catalog`
- `POST /api/v1/internal/api-smoke-check`
- `POST /api/v1/internal/recompute-dashboard`
- `POST /api/v1/internal/refresh-leaderboard`



## Supabase Local Testing
For full local clone + Supabase setup and troubleshooting, see `../docs/LOCAL_SUPABASE_SETUP.md`.


## Local diagnostic smoke test script
After backend is running, execute:
```bash
cd backend
node test.js
```
Full walkthrough: `../docs/LOCAL_TESTING_WITH_TEST_JS.md`.

## Docker-based IDE evaluation worker

### New IDE endpoints
- `POST /api/v1/ide/submit`
- `GET /api/v1/ide/status?id={submission_id}`
- `POST /api/v1/ide/run`

### Docker requirement
Docker Engine must be installed and running on worker hosts. The worker executes untrusted code in one-off containers.

Pre-pull required images:
```bash
docker pull gcc:latest
docker pull openjdk:17
docker pull python:3.11
docker pull node:18
```

### Security controls
The evaluator runs every test case with:
- `--network=none`
- `--memory=128m`
- `--cpus=0.5`
- `--rm`

This blocks outbound networking, limits resource abuse, and destroys containers after each run.

### Local worker run
The IDE worker is started from `cmd/api/main.go` with the API process.

Example env:
```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:54322/postgres?sslmode=disable'
export PORT=8080
```

Run backend (starts HTTP server + assessment worker + IDE worker):
```bash
cd backend
go run cmd/api/main.go
```

### Manual IDE test flow
Submit:
```bash
curl -s -X POST http://localhost:8080/api/v1/ide/submit \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{"question_id":"22222222-2222-2222-2222-222222222221","language":"python","code":"print(input())"}'
```

Check status:
```bash
curl -s 'http://localhost:8080/api/v1/ide/status?id=<submission_id>' -H 'X-User-ID: 00000000-0000-0000-0000-000000000001'
```

Sample run (non-persistent, no mastery update):
```bash
curl -s -X POST http://localhost:8080/api/v1/ide/run \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{"question_id":"22222222-2222-2222-2222-222222222221","language":"python","code":"print(input())"}'
```

### Production recommendations
- Run workers on dedicated judge nodes with Docker daemon isolation.
- Scale workers horizontally; each process polls DB with `FOR UPDATE SKIP LOCKED`.
- Keep worker goroutine count bounded (1 polling loop per process is sufficient initially).
- Tune DB pool for worker traffic (for this repo baseline: max 50 / min 10 connections).

## Course Access Layer

### Endpoints
- `GET /api/v1/course`
- `GET /api/v1/course/topic/{topic_id}`
- `GET /api/v1/course/subtopic/{subtopic_id}`

### Unlock rules
- Diagnostic submission is mandatory for all course reads.
- Topic unlock is derived at read time from persisted `mastery_score`:
  - `mastery_score >= 80` => `completed`
  - root topic (no prerequisites) => `unlocked`
  - all prerequisites `mastery_score >= 80` => `unlocked`
  - otherwise => `locked`

### Derived-state principle
- Unlock state is **never persisted** separately.
- APIs compute deterministic unlock state from database truth (`user_topic_progress` + `topic_prerequisites`).
- No frontend-provided unlock or completion input is trusted.

### External verification model
- External question solves are verified only by platform sync jobs.
- `external_question_activity` and `user_topic_progress.mastery_score` are persisted by sync/evaluation pipelines.
- Course APIs are read-only and automatically reflect updated mastery on next request.

### Security guarantees
- Server-side access control enforces topic/subtopic lock checks.
- Locked topic/subtopic access returns `403` with forbidden payload.
- No manual “mark solved” endpoint exists in course access APIs.

### Concurrency safety
- Course endpoints perform no writes.
- Reads are idempotent, stateless, restart-safe, and concurrency-safe.
