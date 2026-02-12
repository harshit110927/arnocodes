# Codebase Guide (Beginner Friendly)

This document helps a new developer quickly understand the repository and development flow.

## 1) Monorepo layout
- `backend/` → Go API service (current focus)
- `frontend/` → Next.js app
- `ai/` → Python AI service
- `docs/` → product, architecture, schema, and testing docs

## 2) Backend architecture (current skeleton)
- Entry point: `backend/cmd/api/main.go`
- Route registration: `backend/internal/handlers/routes.go`
- API handlers: `backend/internal/handlers/api_v1.go`
- Skeleton endpoint catalog + smoke checks: `backend/internal/skeleton/catalog.go`

## 3) How API development should proceed
1. Add endpoint contract to `docs/API_ROADMAP.md`
2. Add handler placeholder + route
3. Add local testing command to `docs/API_LOCAL_TESTING.md`
4. Add/adjust unit tests
5. Wire DB/service logic in next iteration

## 4) Recommended implementation order
1. Diagnostic test engine (business-critical)
2. Dashboard gating
3. Profile/platform sync persistence
4. Learning progress persistence
5. AI policy enforcement

## 5) Key docs to read first
1. `docs/PRODUCT.md`
2. `docs/DB_SCHEMA.md`
3. `docs/ASSESSMENT_ENGINE.md`
4. `docs/API_ROADMAP.md`
5. `docs/API_LOCAL_TESTING.md`
6. `docs/FRONTEND_HELPER.md`

