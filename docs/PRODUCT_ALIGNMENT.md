# Product-to-Implementation Alignment Review

## Scope
This document maps required product actions to the **current repository implementation** (schema docs + available APIs) and highlights gaps and recommended next steps.

## Current API Inventory (Repository)

### Backend (Go)
- `GET /health`

### AI Service (Python)
- `GET /health`
- `POST /query`
- `POST /index`

> Observation: Backend now has versioned placeholder APIs for core domains; business logic wiring, persistence, and policy enforcement remain pending.

## Flow Coverage Matrix

| Product Action | Schema Coverage | API Coverage | Status | Notes |
|---|---|---|---|---|
| Profile → Update profile | `profiles` exists | Available (placeholder) | Partial | Need persistence, validation, audit trail |
| Profile → Connect platform | Missing dedicated table | Available (placeholder) | Partial | Add `platform_connections` + sync job model |
| Learning → Complete learning question | `user_learning_question_activity` exists | Available (placeholder) | Partial | Add idempotency + scoring integration |
| Learning → Complete subtopic | `user_subtopic_progress` exists | Available (placeholder) | Partial | Threshold validation present; add server-side mastery derivation |
| Assessment → Start test | `tests`, `test_attempts`, `questions` exist | Available (in-memory logic) | Partial | DB persistence + auth middleware pending |
| Assessment → Submit test | `test_attempts`, `question_attempts` exist | Available (in-memory logic) | Partial | DB persistence + deterministic timers pending |
| Platform sync → Trigger sync | `platform_activity` exists | Available (placeholder) | Partial | Need `platform_connections` + `sync_jobs` persistence |
| AI → Ask AI | `ai_usage`, `ai_query_gists` exist | Implemented only in AI service | Partial | Topic-restriction and quota policy enforcement missing |
| AI → Use code helper step | Missing dedicated tables | Available (placeholder) | Partial | Need session/step telemetry model |

## Critical Gaps by Domain

### 1) Assessment lifecycle is under-modeled
- `test_attempts` and `question_attempts` currently lack explicit state machine semantics.
- Without enums and transitions, edge cases become inconsistent (resume test, auto-submit on timeout, partial save).

### 2) Platform connection/sync control plane is missing
- Current schema has `platform_activity` (result data), but lacks source-of-truth for auth, sync state, and retries.

### 3) Dashboard read model is still emerging
- `dashboard_daily_snapshots` is recommended; endpoint layer and freshness strategy are still needed.

### 4) API surface needs service-layer wiring
- Backend endpoint coverage exists, but handlers are still placeholders and need domain service/repository integration.

## Recommended Rollout Sequence (Corporate Team Style)

1. **Foundation (Schema contracts + enums + migrations)**
   - Finalize enums for assessment and progress states.
   - Add missing control-plane tables (`platform_connections`, `platform_sync_jobs`, `ai_code_helper_sessions`).

2. **Core APIs (Backend v1 with services)**
   - Keep endpoints under `/api/v1/*` and wire handlers to service/repository layers.
   - Enforce state-machine transitions for assessment and sync jobs.
   - Add auth, idempotency keys, and quota checks before production rollout.

3. **Read Models + Performance**
   - Enable snapshot jobs for dashboard and leaderboards.
   - Add observability and SLOs on latency/error budgets.

4. **Policy + Guardrails**
   - AI quota enforcement in backend gateway layer.
   - Topic-scoped AI checks and prompt safety.

## Definition of Done (for each feature)
- Schema migration with rollback plan
- API contract + validation + error model
- Unit + integration tests
- Telemetry and dashboards
- Updated docs in `/docs`



## Architecture Decisions Added
- All new endpoints are versioned under `/api/v1` to support backward-compatible evolution.
- Subtopic completion is system-validated (mastery threshold) and should not accept blind completion writes.
- Added optional `GET /api/v1/course/structure` as a single-call DAG read model to simplify onboarding and reduce client orchestration complexity.
