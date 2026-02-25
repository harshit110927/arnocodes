# IDE_ASYNC_EVALUATION_TESTING_GUIDE

## 1. Testing Philosophy

- Backend validation must be completed before frontend integration because IDE evaluation is asynchronous, DB-driven, and worker-based; frontend can only reflect backend state and should never be used as the source of truth for evaluation correctness.
- The async worker path must be tested in isolation from UI flows to verify queue-claim semantics, crash recovery, idempotency, Docker sandbox constraints, and mastery update consistency under concurrent execution.
- Treat the test plan as layered gates:
  1. environment + infra readiness,
  2. deterministic functional behavior,
  3. concurrency and crash safety,
  4. security constraints,
  5. load and pre-production acceptance.

---

## 2. Environment Setup

### 2.1 Prerequisites

```bash
# Docker binary + daemon reachable
which docker
docker --version
docker info > /dev/null && echo "docker daemon reachable"

# Go toolchain
go version

# Optional node (for existing repo scripts)
node --version
```

### 2.2 Pre-pull required runtime images

```bash
docker pull gcc:latest
docker pull openjdk:17
docker pull python:3.11
docker pull node:18
```

### 2.3 Environment variables

Use at least:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:54322/postgres?sslmode=disable'
export PORT=8080
export ENVIRONMENT=local
```

### 2.4 Migration and seed confirmation

```bash
cd backend
go run cmd/api/main.go
```

In another terminal, verify migrations include IDE schema:

```sql
SELECT version, applied_at
FROM schema_migrations
WHERE version IN (
  '002_diagnostic_tables.sql',
  '007_coding_question_test_cases.sql'
)
ORDER BY version;
```

Verify IDE table exists:

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_name IN ('coding_submissions', 'coding_question_test_cases');
```

### 2.5 Worker startup validation

`cmd/api/main.go` starts the IDE worker with server startup. Validate process logs include server start and keep process running.

Sanity check that pending queue can be consumed:

```sql
SELECT evaluation_status, COUNT(*)
FROM coding_submissions
GROUP BY evaluation_status;
```

### 2.6 Verify Docker binary access from backend runtime account

```bash
id
which docker
docker run --rm --network=none --memory=128m --cpus=0.5 --pids-limit=64 --read-only --security-opt=no-new-privileges python:3.11 python -c "print('ok')"
```

---

## 3. Functional Test Cases

> Use a seeded user/topic/question or create a test user. All API calls below assume:
>
> `X-User-ID: 00000000-0000-0000-0000-000000000001`

### 3.0 Seed a deterministic coding test-case set

```sql
-- Replace question_id with a real coding question UUID from questions table.
-- Example seeded coding question in this repo: 55555555-5555-5555-5555-555555555556

INSERT INTO coding_question_test_cases (id, question_id, input, expected_output, is_sample, weight, order_index)
VALUES
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555556', '1 2\n', '3', true, 1, 1),
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555556', '10 20\n', '30', false, 2, 2)
ON CONFLICT DO NOTHING;
```

### 3.1 Happy Path Test

#### Step A — submit valid solution

```bash
curl -s -X POST http://localhost:8080/api/v1/ide/submit \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "question_id":"55555555-5555-5555-5555-555555555556",
    "language":"python",
    "code":"import sys\ndata=sys.stdin.read().strip().split()\nprint(int(data[0])+int(data[1]))"
  }'
```

Capture `submission_id`.

#### Step B — poll status

```bash
curl -s "http://localhost:8080/api/v1/ide/status?id=<submission_id>" \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001'
```

Expected progression: `pending -> processing -> completed`.

#### Step C — DB validation

```sql
SELECT id, evaluation_status, score, evaluated_at
FROM coding_submissions
WHERE id = '<submission_id>'::uuid;
```

Expected:
- `evaluation_status='completed'`
- `score > 0`
- `evaluated_at IS NOT NULL`

#### Step D — mastery side effect validation (practice submissions only)

```sql
-- only applicable if attempt_id is NULL and score >= pass threshold
SELECT id, attempt_id, user_id, question_id, score, evaluation_status
FROM coding_submissions
WHERE id = '<submission_id>'::uuid;
```

Then validate learning progress changed according to existing mastery update behavior (project-specific table/fields used by current learning flow).

---

### 3.2 Compile Error Test

#### Invalid C++

```bash
curl -s -X POST http://localhost:8080/api/v1/ide/submit \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "question_id":"55555555-5555-5555-5555-555555555556",
    "language":"cpp",
    "code":"#include <bits/stdc++.h>\nint main(){ BROKEN SYNTAX }"
  }'
```

#### Invalid Java

```bash
curl -s -X POST http://localhost:8080/api/v1/ide/submit \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "question_id":"55555555-5555-5555-5555-555555555556",
    "language":"java",
    "code":"public class Main { public static void main(String[] args) { BROKEN } }"
  }'
```

Expected:
- `evaluation_status='failed'`
- worker continues to process subsequent submissions

Validation:

```sql
SELECT id, evaluation_status, score, evaluated_at
FROM coding_submissions
WHERE question_id='55555555-5555-5555-5555-555555555556'::uuid
ORDER BY created_at DESC
LIMIT 5;
```

---

### 3.3 Runtime Error Test

#### Python exception

```bash
curl -s -X POST http://localhost:8080/api/v1/ide/submit \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "question_id":"55555555-5555-5555-5555-555555555556",
    "language":"python",
    "code":"raise RuntimeError(\"boom\")"
  }'
```

#### Java division by zero

```bash
curl -s -X POST http://localhost:8080/api/v1/ide/submit \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "question_id":"55555555-5555-5555-5555-555555555556",
    "language":"java",
    "code":"public class Main { public static void main(String[] a){ int x=1/0; System.out.println(x);} }"
  }'
```

Expected failed rows recorded with evaluated timestamp.

---

### 3.4 Infinite Loop Test

```bash
curl -s -X POST http://localhost:8080/api/v1/ide/submit \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "question_id":"55555555-5555-5555-5555-555555555556",
    "language":"python",
    "code":"while True: pass"
  }'
```

Expected:
- timeout-driven failure path
- worker remains responsive for later submissions
- no container leaks

Check no zombie containers remain:

```bash
docker ps --format '{{.ID}} {{.Image}} {{.Command}}'
```

Submit a normal solution immediately after and confirm it completes.

---

### 3.5 Sample Run Test (`/api/v1/ide/run`)

```bash
curl -s -X POST http://localhost:8080/api/v1/ide/run \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "question_id":"55555555-5555-5555-5555-555555555556",
    "language":"python",
    "code":"import sys\ndata=sys.stdin.read().split(); print(int(data[0])+int(data[1]))"
  }'
```

Expected:
- returns immediate sample result only
- **no mastery update**
- **no new async queue row** for sample run path

Validation:

```sql
SELECT COUNT(*)
FROM coding_submissions
WHERE created_at > NOW() - INTERVAL '1 minute'
  AND language='python'
  AND question_id='55555555-5555-5555-5555-555555555556'::uuid;
```

(Count should change only for `/submit`, not `/run`.)

---

## 4. Concurrency Testing

### 4.1 Parallel submit burst (10–20)

```bash
for i in $(seq 1 20); do
  curl -s -X POST http://localhost:8080/api/v1/ide/submit \
    -H 'Content-Type: application/json' \
    -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' \
    -d "{\"question_id\":\"55555555-5555-5555-5555-555555555556\",\"language\":\"python\",\"code\":\"print(3)\"}" >/tmp/ide-submit-$i.json &
done
wait
```

### 4.2 Verify no duplicate processing / race anomalies

```sql
SELECT evaluation_status, COUNT(*)
FROM coding_submissions
WHERE created_at > NOW() - INTERVAL '5 minutes'
GROUP BY evaluation_status;
```

```sql
-- detect any impossible double-finalization symptoms
SELECT id, COUNT(*)
FROM coding_submissions
WHERE created_at > NOW() - INTERVAL '5 minutes'
GROUP BY id
HAVING COUNT(*) > 1;
```

```sql
-- latest submission per question for user
SELECT id, question_id, created_at, evaluation_status, score
FROM coding_submissions
WHERE user_id='00000000-0000-0000-0000-000000000001'::uuid
ORDER BY question_id, created_at DESC;
```

Expected:
- no deadlocks
- worker keeps draining queue
- mastery side effect occurs only for latest qualifying practice submission

---

## 5. Crash Recovery Testing

### 5.1 Simulate worker crash mid-processing

1. Submit a long-running job (infinite loop).
2. Kill API/worker process immediately after status becomes `processing`.

```bash
pkill -f "cmd/api/main.go"
```

3. Restart backend.

```bash
cd backend
go run cmd/api/main.go
```

### 5.2 Force stuck-row simulation via SQL

```sql
UPDATE coding_submissions
SET evaluation_status='processing',
    score=NULL,
    evaluated_at=NULL,
    created_at=NOW() - INTERVAL '10 minutes'
WHERE id='<submission_id>'::uuid;
```

### 5.3 Validate stuck reset behavior

```sql
SELECT id, evaluation_status, score, evaluated_at, created_at
FROM coding_submissions
WHERE id='<submission_id>'::uuid;
```

Expected after worker cycle:
- row transitions back to `pending`, then reprocessed once
- final row reaches terminal state once
- mastery applied at most once (latest-qualifying only)

---

## 6. Idempotency Testing

### 6.1 Double finalization protection

Validate same submission is not repeatedly finalized by replay worker cycle.

```sql
SELECT id, evaluation_status, evaluated_at
FROM coding_submissions
WHERE id='<submission_id>'::uuid;
```

Restart worker multiple times and re-check row remains stable.

### 6.2 Duplicate worker cycle safety

Run two backend instances against same DB (locally if possible), then submit batch.

Expected:
- `FOR UPDATE SKIP LOCKED` prevents duplicate claims
- each submission finalized once

Anomaly check:

```sql
SELECT evaluation_status, COUNT(*)
FROM coding_submissions
WHERE created_at > NOW() - INTERVAL '10 minutes'
GROUP BY evaluation_status;
```

---

## 7. Security Validation

### 7.1 Confirm effective runtime restrictions

```bash
# observe docker run args from process execution logs or audit wrapper
# manual probe with same flags:
docker run --rm --network=none --memory=128m --cpus=0.5 --pids-limit=64 --read-only --security-opt=no-new-privileges python:3.11 python -c "print('ok')"
```

### 7.2 Network isolation check

```bash
docker run --rm --network=none python:3.11 python -c "import socket; socket.gethostbyname('google.com')"
```

Expected: failure due to no network.

### 7.3 Read-only filesystem check

```bash
docker run --rm --read-only python:3.11 python -c "open('/tmp/x','w').write('x')"
```

Expected: write failure.

### 7.4 PID limit check

```bash
docker run --rm --pids-limit=64 python:3.11 python -c "import os,time; [os.fork() for _ in range(200)]"
```

Expected: process creation constrained.

### 7.5 Privilege escalation check

```bash
docker run --rm --security-opt=no-new-privileges python:3.11 id
```

Expected: no elevated privileges.

---

## 8. Load Simulation

### 8.1 Basic concurrent submission script

```bash
#!/usr/bin/env bash
set -euo pipefail
BASE_URL="http://localhost:8080"
USER_ID="00000000-0000-0000-0000-000000000001"
QUESTION_ID="55555555-5555-5555-5555-555555555556"

for i in $(seq 1 100); do
  curl -s -X POST "$BASE_URL/api/v1/ide/submit" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: $USER_ID" \
    -d "{\"question_id\":\"$QUESTION_ID\",\"language\":\"python\",\"code\":\"print(1)\"}" >/dev/null &
  if (( i % 20 == 0 )); then wait; fi
done
wait
echo "load burst completed"
```

### 8.2 Observability during load

```bash
# docker resource usage
docker stats --no-stream

# backend process CPU/memory
top -pid $(pgrep -f "cmd/api/main.go" | head -n1)
```

```sql
-- processing rate snapshot
SELECT evaluation_status, COUNT(*)
FROM coding_submissions
WHERE created_at > NOW() - INTERVAL '15 minutes'
GROUP BY evaluation_status;
```

### 8.3 Throughput estimation guidance

- Measure avg completion latency over N submissions.
- Compute worker throughput as `N / total_time_seconds`.
- Increase concurrency gradually (20 -> 50 -> 100) and monitor DB pool saturation / worker lag.

---

## 9. Edge Case Testing

1. **Zero test cases**
   - remove all test cases for a question and submit.
   - expect deterministic handled outcome (currently completed with score 0 by evaluator behavior).

```sql
DELETE FROM coding_question_test_cases WHERE question_id='55555555-5555-5555-5555-555555555556'::uuid;
```

2. **Very large input**
   - insert large `input` payload in test case and submit standard parser solution.

3. **Large output**
   - expected output with large body; verify worker remains stable and finalizes.

4. **Rapid resubmission same question**
   - submit same question repeatedly in short time; ensure latest-only mastery semantics hold.

5. **Attempt-bound submissions (attempt_id not NULL)**
   - submit with attempt_id and high score; verify mastery hook does not run for attempt-linked rows.

---

## 10. Database Integrity Checks

```sql
-- 10.1 No rows stuck in processing beyond threshold
SELECT id, user_id, question_id, created_at
FROM coding_submissions
WHERE evaluation_status='processing'
  AND NOW() - created_at > INTERVAL '5 minutes';
```

```sql
-- 10.2 Completed/failed rows must have evaluated_at
SELECT id, evaluation_status, evaluated_at
FROM coding_submissions
WHERE evaluation_status IN ('completed','failed')
  AND evaluated_at IS NULL;
```

```sql
-- 10.3 Orphan check (question FK)
SELECT cs.id
FROM coding_submissions cs
LEFT JOIN questions q ON q.id = cs.question_id
WHERE q.id IS NULL;
```

```sql
-- 10.4 Sanity distribution
SELECT evaluation_status, COUNT(*)
FROM coding_submissions
GROUP BY evaluation_status;
```

---

## 11. Pre-Production Checklist

- [ ] Worker stable under 50+ concurrent submissions in staging
- [ ] No container leaks (`docker ps` clean after workload)
- [ ] No DB deadlocks observed in logs
- [ ] Mastery side effects consistent with latest-submission rule
- [ ] Stuck reset functioning (`processing` rows older than 5m get recovered)
- [ ] Logs clean (no repetitive evaluator crash loops)
- [ ] Docker resource/security flags verified in runtime execution
- [ ] Sample run path verified to not enqueue async scoring rows
- [ ] Compile/runtime/timeout failures correctly transition to terminal states
- [ ] Restart recovery tested successfully

---

## 12. Post-Deployment Monitoring Recommendations

Track and alert on:

1. **Submission lifecycle metrics**
   - pending count
   - processing count
   - completion latency (submit -> evaluated)
   - failed rate by language

2. **Worker health**
   - polling loop errors
   - stuck reset count per hour
   - finalize-with-mastery transaction failures

3. **DB health**
   - lock wait times on `coding_submissions`
   - transaction rollback spikes
   - connection pool saturation

4. **Docker/judge health**
   - container start failure rate
   - timeout frequency
   - OOM kill frequency

5. **Mastery integrity signals**
   - count of completed submissions eligible for mastery vs applied mastery updates
   - anomaly checks for duplicate or missing progression updates

Suggested recurring checks:

```sql
-- queue pressure trend
SELECT date_trunc('minute', created_at) AS minute_bucket,
       COUNT(*) FILTER (WHERE evaluation_status='pending') AS pending,
       COUNT(*) FILTER (WHERE evaluation_status='processing') AS processing,
       COUNT(*) FILTER (WHERE evaluation_status='completed') AS completed,
       COUNT(*) FILTER (WHERE evaluation_status='failed') AS failed
FROM coding_submissions
WHERE created_at > NOW() - INTERVAL '2 hours'
GROUP BY 1
ORDER BY 1;
```

```sql
-- prolonged processing watchdog
SELECT COUNT(*) AS stuck_candidates
FROM coding_submissions
WHERE evaluation_status='processing'
  AND NOW() - created_at > INTERVAL '5 minutes'
  AND evaluated_at IS NULL
  AND score IS NULL;
```
