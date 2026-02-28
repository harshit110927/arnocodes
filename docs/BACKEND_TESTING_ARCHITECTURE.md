# Backend Testing Architecture Blueprint

## Audience & Scope
This document defines the production testing architecture for the Go backend powering the DSA mastery platform. It is implementation-ready guidance for engineering and QA teams.

---

## 1) Testing Philosophy

### Why layered testing is required
A production backend with auth, async workers, DB transactions, and sandboxed code execution must be tested in layers because each layer validates different failure modes:

- **Unit tests** validate pure logic and fast edge-case iteration.
- **Integration tests** validate DB behavior, transactions, migrations, worker idempotency, and external boundary assumptions.
- **HTTP tests** validate middleware, auth propagation, status codes, and schema-level contract behavior.
- **End-to-end tests** validate full user journeys and cross-domain invariants.

### Why unit tests alone are insufficient
Unit tests do not reveal:
- SQL query regressions.
- transaction isolation bugs.
- middleware + handler wiring issues.
- async retry/idempotency behavior.
- auth token verification runtime edge cases.

### Why integration tests are critical
This system’s core correctness depends on:
- Postgres transaction semantics.
- explicit `WHERE user_id = $1` filtering.
- migration correctness.
- worker restart behavior.
- RLS policy presence.

These are only verifiable through integration tests against a real database.

### Why HTTP layer tests must exist
Middleware and handlers are the enforcement boundary for:
- JWT verification.
- context user propagation.
- 401/403/404/422 behavior.
- JSON response contract.

Repository/service unit tests cannot guarantee these behaviors.

---

## 2) Unit Test Coverage Matrix

### Coverage targets (minimum)
- `internal/middleware`: **95%+**
- `internal/course`: **90%+**
- `internal/assessment` service logic: **85%+**
- `internal/ide` service + evaluator command construction: **85%+**
- `internal/dashboard` derived logic: **85%+**
- overall backend line coverage: **80%+**

### Package-level matrix

| Package | Unit test focus | Critical edge cases |
|---|---|---|
| `internal/middleware` | JWT parse + claims validation + JWKS cache behavior | expired token, wrong issuer, wrong audience, wrong alg, missing kid, invalid signature, JWKS refresh failure fallback |
| `internal/course` | derived unlock DAG logic | 79.99 vs 80 boundary, no prerequisites, multi-prerequisite lock, empty topic set, missing progress rows |
| `internal/assessment` | diagnostic orchestration state transitions | invalid attempt state, duplicate submit, empty answer payloads, nil optional fields |
| `internal/dashboard` | snapshot/streak/mastery display derivation | missing snapshot rows, zero counts, day boundary handling |
| `internal/ide` | submission service, status behavior, worker orchestration contracts | invalid language, question missing, duplicate finalize guard, stale processing reset |
| `internal/learning/activity` | write-entrypoint transaction flow | nil repo, tx begin/commit failures |
| `internal/handlers` | request validation + error mapping | malformed JSON, missing required params, incorrect method |

### Boundary values
- mastery threshold checks: **79.99 / 80.00 / 80.01**
- score threshold checks (ide pass): **79.99 / 80.00 / 80.01**
- timeout durations: exact boundary and +1ms scenarios.

### Nil pointer risk zones
Add explicit unit tests for constructors and nil dependencies:
- nil repository injection.
- nil evaluator.
- nil auth middleware in route registration.
- nil status providers.

### Unlock cycle edge case
Even with DAG assumption, add defensive unit test for cyclic prerequisite input to ensure code does not panic/hang and returns deterministic locked states.

### Concurrency unsafe zones (unit-level simulation)
- JWKS cache read/write interleavings.
- worker finalization guard (second finalize no-op).

---

## 3) Integration Test Architecture

### Environment strategy
Use ephemeral integration environment per test run:
- Postgres test DB (container or Supabase test project).
- optional Redis container (if rate limiting/abuse logic depends on it).

### Test DB bootstrap flow
1. Create isolated DB/schema per run.
2. Execute migrations programmatically.
3. Seed deterministic fixtures.
4. Run test suite.
5. Drop schema/database.

### Deterministic seeding
Seed stable IDs and deterministic timestamps for:
- users, topics, prerequisites, subtopics.
- diagnostic questions + attempts.
- learning questions and IDE test cases.

### Isolation strategy
Preferred options:
- **Per-test transaction rollback** for read/write tests.
- **Per-test schema** for worker/concurrency tests that require committed state.

### DB cleanup strategy
- Truncate user-owned tables between tests (or rollback transaction).
- reset sequence/UUID fixture maps where needed.

### Example integration boot helper (concept)
```go
func NewIntegrationDB(t *testing.T) *pgxpool.Pool {
  // create schema/test db
  // run migrations
  // run seed
  // return pool
}
```

---

## 4) HTTP Layer Testing

### Test harness
- Use `httptest.NewServer` for full middleware + routing behavior.
- Build server with real handlers and in-memory/test dependencies.

### JWT injection strategy
- generate test RSA keypair.
- issue RS256 tokens with adjustable claims (`iss`, `aud`, `exp`, `sub`, `kid`).
- serve JWKS from local test HTTP endpoint.

### Required status-code coverage
- `401`: missing/invalid token.
- `403`: auth valid but domain forbidden (e.g., diagnostic required/locked).
- `404`: unknown resources.
- `422`: validation failures.

### JSON schema validation
Use JSON schema assertions for stable payload shape:
- required fields present.
- error envelope structure:
  ```json
  {"error":"..."}
  ```

---

## 5) Authentication Testing Plan

Validate all JWT failure and success modes:

1. Missing `Authorization` header -> 401.
2. Malformed bearer prefix -> 401.
3. Expired token (`exp` in past) -> 401.
4. Wrong issuer (`iss`) -> 401.
5. Wrong audience (`aud`) -> 401.
6. Wrong algorithm (`HS256`) -> 401.
7. Missing `kid` header -> 401.
8. Invalid signature -> 401.
9. JWKS refresh failure with cached key available -> request still succeeds.
10. JWKS refresh failure with no cached key -> 401.
11. Valid token with correct `sub` -> request succeeds and `user_id` propagated.

---

## 6) Rate Limiting Tests

> If rate limiting middleware is introduced/active, validate this matrix.

### Scenarios
- Per-IP limit enforcement.
- Per-user limit enforcement.
- Burst requests at window start.
- Sliding window boundary rollover.
- concurrent requests from same user.
- distributed instance simulation (shared Redis backend).

### Assertions
- proper `429` status.
- retry headers (if implemented).
- no false-positive throttle under normal usage.

---

## 7) Abuse Detection Tests

### Cases
- excessive IDE submissions in short interval.
- repeated diagnostic start attempts.
- AI quota overflow behavior.
- temporary user block escalation.
- security event log emission.

### Assertions
- mitigation triggered deterministically.
- block expiration respected.
- logging includes user_id, route, reason, timestamp.

---

## 8) IDE & Docker Evaluator Tests

### Functional safety tests
- compile failure (C++/Java syntax error) -> failed status.
- runtime exception -> failed status.
- infinite loop -> timeout -> failed.
- memory overuse -> killed -> failed.
- large output overflow -> bounded handling.
- huge code payload -> validated/rejected gracefully.

### Sandbox tests
Verify docker invocation includes:
- `--network=none`
- memory limit
- cpu limit
- read-only FS
- pids limit
- no-new-privileges

### Cleanup tests
- container removed after each test case.
- no zombie processes after timeout.

---

## 9) Worker Testing

### Core validations
- claim-next semantics process one job at a time per worker (`SKIP LOCKED`).
- retry or reset behavior for stuck processing records.
- idempotent finalization (double finalize no-op).
- duplicate job attempts do not duplicate mastery updates.
- crash mid-processing preserves DB consistency.

### Crash recovery scenarios
1. worker claims job, crashes before finalize.
2. stale processing reset job runs.
3. new worker reclaims and finalizes once.

---

## 10) Concurrency & Race Tests

### Parallel test scenarios
- parallel diagnostic submissions for same user.
- parallel subtopic completion requests.
- concurrent IDE submissions for same `(user, question)`.
- concurrent dashboard snapshot updates.

### Tooling
- `go test -race ./...`
- custom parallel integration tests using goroutines + waitgroups.

### Assertions
- no data races.
- no deadlocks.
- transactional invariants preserved.

---

## 11) Failure Scenario Testing

Simulate and assert graceful behavior for:
- Postgres unavailable at startup and runtime.
- Redis unavailable (if used by rate limiting/abuse).
- JWKS endpoint unavailable/timeout.
- Docker daemon unavailable.
- migration failure.
- seed failure.
- context deadlines/timeouts on DB calls.

Expected: fail-fast at startup where required; controlled error responses at runtime.

---

## 12) RLS & Authorization Testing

Even with service-role backend connection, validate:
- explicit repository filtering prevents cross-user reads.
- user A cannot access user B records via API.
- all user-owned queries include `WHERE user_id = $1` semantics.
- RLS policies exist and are syntactically valid.

Add SQL policy existence checks in migration tests:
```sql
SELECT * FROM pg_policies WHERE tablename = 'user_topic_progress';
```

---

## 13) Full End-to-End User Journey Test

### Journey script
1. Authenticate user (Supabase JWT).
2. Start diagnostic.
3. Submit answers (mcq/coding).
4. Finalize diagnostic.
5. Access course root.
6. Verify unlock state transition.
7. Submit IDE code.
8. Poll IDE status until completion.
9. Fetch dashboard summary.

### Assertions
- unlock/diagnostic gating correct.
- submission lifecycle transitions valid.
- dashboard reflects persisted activity.

---

## 14) Load & Stress Testing Plan

### Tooling
- `k6` or `vegeta`.

### Initial target profile
- warmup: 50 RPS for 2m.
- baseline: 100 RPS for 5m.
- stress: ramp to 300 RPS for 10m.

### Monitor
- p50/p95/p99 latency.
- error rate.
- DB connection pool saturation.
- worker lag and queue depth.
- CPU/memory of backend and DB.

### Bottleneck detection
- identify lock contention.
- slow SQL traces.
- middleware overhead.

---

## 15) Fuzz Testing Strategy

### Fuzz targets
- HTTP JSON payload bodies (all POST/PATCH endpoints).
- JWT token parser inputs.
- IDE code payload and language fields.
- diagnostic answer payload structure.

### Execution
- use Go fuzzing for parser/validator-heavy paths.
- run nightly fuzz budget with corpus retention.

---

## 16) CI/CD Testing Pipeline Design

### GitHub Actions stages
1. **Static checks**: `go vet`, formatting check, lint.
2. **Unit tests**: `go test ./...` excluding integration tags.
3. **Race tests**: `go test -race ./...`.
4. **Integration tests**: Postgres + Redis service containers.
5. **Coverage**: aggregate and enforce thresholds.
6. **Artifact publish**: junit/coverage reports.

### Service containers
- Postgres container for integration tests.
- Redis container for abuse/rate-limit tests (if enabled).

### Fail build conditions
- any failing test stage.
- race detector failures.
- coverage drops below thresholds.
- migration test failures.

### Example commands
```bash
go test ./...
go test -race ./...
go test -tags=integration ./internal/...
go vet ./...
```

---

## 17) Observability Validation

### Logging validation
- assert structured logs include: request_id, user_id, path, status, latency.
- security failures (401, invalid token, abuse blocks) logged with reason.
- worker failures logged with job identifiers.

### Panic recovery validation
- intentionally trigger handler panic in test-only endpoint/harness.
- verify middleware recovers and returns 500 without crashing process.

### Metrics/trace validation
- request counts by route/status.
- auth failure rate.
- worker queue/processing durations.
- evaluator timeout and failure metrics.

---

## Recommended Test Directory Structure

```text
backend/
  internal/
    middleware/
      auth_test.go
    handlers/
      api_v1_test.go
      auth_http_test.go
    assessment/
      service_test.go
      integration_test.go
    course/
      service_test.go
      integration_test.go
    dashboard/
      repository_test.go
      integration_test.go
    ide/
      service_test.go
      repository_test.go
      worker_integration_test.go
      docker_evaluator_test.go
  test/
    integration/
      fixtures/
      jwt/
      helpers/
```

---

## Execution Playbook (Developer)

```bash
# unit
cd backend
go test ./internal/... -count=1

# race
cd backend
go test -race ./internal/... -count=1

# integration (requires DATABASE_URL)
cd backend
DATABASE_URL=postgres://... go test -tags=integration ./internal/... -count=1

# coverage
cd backend
go test ./... -coverprofile=coverage.out

go tool cover -func=coverage.out
```

---

## Acceptance Criteria for “Production-Ready Testing”

- Unit + integration + HTTP + E2E suites exist and run in CI.
- Auth/JWT negative test matrix fully covered.
- Worker idempotency and crash recovery proven by tests.
- DB migrations tested on clean DB every pipeline run.
- Explicit user isolation validated across all user-owned APIs.
- Race tests pass consistently.
- Coverage thresholds met for critical packages.


## Appendix: One-Command Backend Flow Script

Run the step-by-step validator:

```bash
DATABASE_URL=postgres://... \
SUPABASE_URL=https://<project-ref>.supabase.co \
TEST_JWT=<supabase-access-token> \
./backend/scripts/full_backend_test_flow.sh
```

Notes:
- If `TEST_JWT` is omitted, the script still validates health + unauthorized guards.
- With `TEST_JWT`, it validates authenticated diagnostic/course/dashboard/IDE flow checks.
- Script location: `backend/scripts/full_backend_test_flow.sh`.


### Script flow stages (real-user-like order)
1. Health + auth guard verification
2. Authenticated profile status
3. Diagnostic start/next/submit path (or blocked handling)
4. Course/dashboard gated reads
5. Platform connect/list/sync/disconnect
6. IDE sample run + async submit + polling


### Platform Sync Retry Verification
- Validate trigger retries by simulating transient DB failure (or dependency timeout) and confirming 3-attempt exponential backoff behavior.
- Expected response after final failure: temporary error with guidance to retry after 5–10 seconds.
- Validate `GET /api/v1/platform-sync/overview` for audit-ready status counts and recent job list.
