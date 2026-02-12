# Frontend Helper Guide (Pre-DB Workflow Testing)

This guide explains exactly what frontend needs to test the end-to-end flow **before database integration**.

## 1) Base URL + Versioning
Use backend base URL with API versioning:
- `BACKEND_BASE_URL=http://localhost:8080`
- `API_PREFIX=/api/v1`

All feature calls should go to `${BACKEND_BASE_URL}${API_PREFIX}/...`.

## 2) Recommended Frontend API Client Shape
Create a thin client wrapper with:
- `request(method, path, body?)`
- automatic JSON headers
- `X-User-ID` header support (required for stateful assessment simulation in local mode)
- typed response envelope: `{ status, message, data }`
- centralized error handling for `4xx/5xx`

## 3) Use Backend Skeleton Introspection Endpoints
Backend now provides:
- `GET /api/v1/internal/api-catalog` → machine-readable endpoint catalog
- `POST /api/v1/internal/api-smoke-check` → run all API checks in-process and return pass/fail report

Use these to verify your frontend route map and to gate CI for mock-integration phase.

## 4) Workflow Test Checklist (No DB)

### Profile
- Fetch profile: `GET /api/v1/profiles/me`
- Fetch status: `GET /api/v1/profiles/me/status`
- Update profile: `PATCH /api/v1/profiles/me`
- Connect/list/disconnect platform

### Learning
- Fetch DAG in one shot: `GET /api/v1/course/structure`
- Fetch topics/subtopics details
- Complete learning question
- Complete subtopic (send `mastery_score`)
  - expect `422` when below threshold
  - expect `202` when threshold met

### Assessment
- Start test → load session
- Submit answer per question
- Get next question
- Submit attempt
- Expire/resume attempt

### Platform sync
- Trigger sync
- Check sync job status

### AI
- Ask AI
- Use code helper step
- Read usage quota/status

### Internal jobs (dev/testing only)
- Recompute dashboard
- Refresh leaderboard

## 5) Suggested Frontend Mock Screens
Before DB integration, add temporary QA screens:
1. **API Explorer Panel**: list endpoints from catalog and allow one-click trigger.
2. **Assessment Session Panel**: test start/session/next-question/submit flow quickly.
3. **Learning Validation Panel**: submit subtopic completion with variable `mastery_score`.
4. **Ops Panel**: run smoke check endpoint and show pass/fail table.

## 6) Minimal CI for Frontend During Skeleton Phase
- Smoke test that app can call `GET /api/v1/health`
- Validate API catalog endpoint returns non-empty list
- Validate smoke-check endpoint returns `failed = 0`

## 7) Integration Contract Notes
- Treat current responses as placeholders; only rely on:
  - endpoint path
  - HTTP method
  - status-code expectations
  - common response envelope shape
- Replace placeholder assumptions with typed DTO contracts once services are wired.


## 8) Companion backend test doc
For backend-first API testing commands (curl + JSON payloads), see `docs/API_LOCAL_TESTING.md`.

## 9) Protected resource behavior
- Dashboard/learning APIs are backend-guarded.
- Before diagnostic submit, expect `403` with `{ "error": "DIAGNOSTIC_REQUIRED" }`.
- Frontend should show lock icon/state using `/profiles/me/status` and route users to diagnostic flow.
