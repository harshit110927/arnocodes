# API Local Testing Playbook (PostgreSQL Infrastructure Phase)

## 1) Start backend
```bash
cd backend
go run cmd/api/main.go
```

## 2) Infra checks
```bash
curl -s http://localhost:8080/api/v1/health | jq
curl -s http://localhost:8080/api/v1/internal/api-catalog | jq
curl -s -X POST http://localhost:8080/api/v1/internal/api-smoke-check | jq
```

## 3) Profile + status
```bash
curl -s http://localhost:8080/api/v1/profiles/me | jq
curl -s http://localhost:8080/api/v1/profiles/me/status -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' | jq
```

## 4) Guard behavior for protected resources
Before diagnostic completion in DB, expect 403:
```bash
curl -s http://localhost:8080/api/v1/dashboard/summary -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' | jq
curl -s http://localhost:8080/api/v1/topics -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' | jq
```

Expected body:
```json
{ "error": "DIAGNOSTIC_REQUIRED" }
```

## 5) Assessment endpoints (currently repository-backed stubs)
```bash
curl -s http://localhost:8080/api/v1/tests/diagnostic-1 | jq
curl -s -X POST http://localhost:8080/api/v1/tests/diagnostic-1/start | jq
curl -s http://localhost:8080/api/v1/test-attempts/attempt-1/next-question | jq
```

> Note: these routes are now prepared for DB-backed implementation and intentionally kept as stubs in this infra phase.

## 6) Migration + seed validation (manual SQL)
Use psql and verify:
- `schema_migrations` has `001_init.sql`
- `tests` contains diagnostic seed
- `questions` contains slide+mcq seed rows
- `topics`, `subtopics`, `learning_questions` contain seed rows



For local machine + Supabase DB connection setup, see `docs/LOCAL_SUPABASE_SETUP.md`.

## Diagnostic API local test flow (DB-backed)

Set user header:

```bash
export API=http://localhost:8080
export USER_ID=00000000-0000-0000-0000-000000000001
```

1) Start diagnostic:
```bash
curl -i -X POST "$API/api/v1/diagnostic/start" \
  -H "X-User-ID: $USER_ID" -H "Content-Type: application/json" \
  -d '{"selected_topics":["22222222-2222-2222-2222-222222222221","22222222-2222-2222-2222-222222222222"]}'
```

2) Get next question:
```bash
curl -i -H "X-User-ID: $USER_ID" "$API/api/v1/diagnostic/<attempt_id>/next"
```

3) Submit MCQ answer:
```bash
curl -i -X POST "$API/api/v1/diagnostic/<attempt_id>/answer" \
  -H "X-User-ID: $USER_ID" -H "Content-Type: application/json" \
  -d '{"question_id":"<question_id>","question_type":"mcq","selected_option":2}'
```

4) Submit coding answer:
```bash
curl -i -X POST "$API/api/v1/diagnostic/<attempt_id>/coding" \
  -H "X-User-ID: $USER_ID" -H "Content-Type: application/json" \
  -d '{"question_id":"<question_id>","question_type":"coding","language":"go","code":"package main\nfunc solve(){}"}'
```

5) Check status and submit:
```bash
curl -i -H "X-User-ID: $USER_ID" "$API/api/v1/diagnostic/<attempt_id>/status"
curl -i -X POST -H "X-User-ID: $USER_ID" "$API/api/v1/diagnostic/<attempt_id>/submit"
```
