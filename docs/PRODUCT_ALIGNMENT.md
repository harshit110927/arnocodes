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

> Observation: Product-critical APIs for profile, learning, assessment, sync, and dashboard are not yet implemented in the backend service.

## Flow Coverage Matrix

| Product Action | Schema Coverage | API Coverage | Status | Notes |
|---|---|---|---|---|
| Profile → Update profile | `profiles` exists | Missing | Partial | Need `PATCH /profiles/me` + validation/audit |
| Profile → Connect platform | Missing dedicated table | Missing | Gap | Add platform connections + sync job model |
| Learning → Complete learning question | `user_learning_question_activity` exists | Missing | Partial | Need idempotent completion endpoint |
| Learning → Complete subtopic | `user_subtopic_progress` exists | Missing | Partial | Should auto-derive topic progress |
| Assessment → Start test | `tests`, `test_attempts`, `questions` exist | Missing | Partial | Needs attempt status + timing state |
| Assessment → Submit test | `test_attempts`, `question_attempts` exist | Missing | Partial | Needs explicit lifecycle and grading pipeline |
| Platform sync → Trigger sync | `platform_activity` exists | Missing | Partial | Need `platform_connections` + `sync_jobs` |
| AI → Ask AI | `ai_usage`, `ai_query_gists` exist | Implemented only in AI service | Partial | Topic-restriction and quota policy enforcement missing |
| AI → Use code helper step | Missing dedicated tables | Missing | Gap | Need session/step telemetry model |

## Critical Gaps by Domain

### 1) Assessment lifecycle is under-modeled
- `test_attempts` and `question_attempts` currently lack explicit state machine semantics.
- Without enums and transitions, edge cases become inconsistent (resume test, auto-submit on timeout, partial save).

### 2) Platform connection/sync control plane is missing
- Current schema has `platform_activity` (result data), but lacks source-of-truth for auth, sync state, and retries.

### 3) Dashboard read model is still emerging
- `dashboard_daily_snapshots` is recommended; endpoint layer and freshness strategy are still needed.

### 4) API surface is far from product flow
- Backend currently exposes only `/health`; all core product APIs are still pending.

## Recommended Rollout Sequence (Corporate Team Style)

1. **Foundation (Schema contracts + enums + migrations)**
   - Finalize enums for assessment and progress states.
   - Add missing control-plane tables (`platform_connections`, `platform_sync_jobs`, `ai_code_helper_sessions`).

2. **Core APIs (Backend v1)**
   - Profile: `GET/PATCH /profiles/me`
   - Learning: `POST /learning/questions/{id}/complete`, `POST /subtopics/{id}/complete`
   - Assessment: `POST /tests/{id}/start`, `POST /attempts/{id}/answer`, `POST /attempts/{id}/submit`
   - Sync: `POST /platform-sync/trigger`

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

