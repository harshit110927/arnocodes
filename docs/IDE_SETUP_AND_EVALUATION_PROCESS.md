# IDE Setup, Question Authoring, and Evaluation Process

## 1. Overview
This guide explains how to run and operate the async IDE evaluator, how to add coding questions and test cases (visible + hidden), and how scoring works.

## 2. Prerequisites
- Docker daemon installed and running.
- Backend environment configured (`DATABASE_URL`, `SUPABASE_URL`, etc.).
- Required evaluator images pre-pulled:
  - `gcc:latest`
  - `openjdk:17`
  - `python:3.11`
  - `node:18`

## 3. How IDE Evaluation Works
1. `POST /api/v1/ide/submit` inserts a row into `coding_submissions` with `evaluation_status='pending'`.
2. Worker (`StartIDEWorker`) periodically claims one pending submission (`FOR UPDATE SKIP LOCKED`) and marks it `processing`.
3. Evaluator (`DockerEvaluator`) writes code to temp dir and executes test cases in isolated one-off containers.
4. Finalization updates submission status/score atomically and conditionally applies learning activity side effects.

## 4. Runtime Sandbox Controls
Each test execution runs with:
- `--network=none`
- `--memory=128m`
- `--cpus=0.5`
- `--pids-limit=64`
- `--read-only`
- `--security-opt=no-new-privileges`

## 5. Supported Languages
- C++ (`cpp`)
- Java (`java`)
- Python (`python`)
- JavaScript (`javascript`, `js`)

## 6. Add Coding Questions
Insert questions into `questions` with `question_type='coding'`.

Example:
```sql
INSERT INTO questions (id, test_id, question_type, content, marks, order_index)
VALUES (
  gen_random_uuid(),
  '11111111-1111-1111-1111-111111111111',
  'coding',
  'Write a function to reverse an array.',
  10,
  50
);
```

## 7. Add Test Cases (Visible and Hidden)
Use `coding_question_test_cases`:
- `is_sample=true` => visible/sample test cases used by `/api/v1/ide/run`
- `is_sample=false` => hidden test cases used by async scoring in `/api/v1/ide/submit`

Example:
```sql
INSERT INTO coding_question_test_cases
(id, question_id, input, expected_output, is_sample, weight, order_index, created_at)
VALUES
(gen_random_uuid(), '<question_uuid>', '1 2 3', '3 2 1', true, 1, 1, NOW()),
(gen_random_uuid(), '<question_uuid>', '10 20 30', '30 20 10', false, 2, 2, NOW());
```

## 8. Scoring Model
- Score is weighted percentage:
  - `score = (sum(weights of passed tests) / sum(all test weights)) * 100`
- Missing/zero weight defaults to `1`.
- No test cases => score `0` and completed result.

## 9. Visible vs Hidden Evaluation
- `/api/v1/ide/run`:
  - evaluates only `is_sample=true` test cases.
  - does not enqueue async submission.
- `/api/v1/ide/submit`:
  - enqueues async job.
  - worker evaluates with `is_sample=false` (full scoring set).

## 10. Worker Operations
- Reset stuck processing rows (`> 5 minutes`) back to `pending`.
- Claim one pending row at a time safely.
- Finalize with idempotent guard (`WHERE evaluation_status='processing'`).

## 11. Validation Commands
```bash
# syntax check script
bash -n backend/scripts/full_backend_test_flow.sh

# run end-to-end backend flow
DATABASE_URL=postgres://... \
SUPABASE_URL=https://<project>.supabase.co \
TEST_JWT=<access_token> \
./backend/scripts/full_backend_test_flow.sh
```


## 12. Docker connectivity checks
Run before production rollout:
```bash
docker info
docker run --rm --network=none --read-only --security-opt=no-new-privileges python:3.11 python -c "print(1)"
```
If these fail, IDE worker submissions will remain pending/failed.
