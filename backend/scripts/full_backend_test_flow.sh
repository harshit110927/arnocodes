#!/usr/bin/env bash
set -euo pipefail

# Full backend flow validator for ArnoCodes.
# This script runs logical step-by-step checks from process boot -> auth -> diagnostic -> course -> IDE -> dashboard.
#
# Required env:
#   DATABASE_URL
#   SUPABASE_URL
# Optional env:
#   SUPABASE_AUDIENCE (default: authenticated)
#   TEST_JWT (Supabase access token for authenticated flow)
#   PORT (default: 18080)
#
# Usage:
#   DATABASE_URL=... SUPABASE_URL=... ./backend/scripts/full_backend_test_flow.sh
#   DATABASE_URL=... SUPABASE_URL=... TEST_JWT=... ./backend/scripts/full_backend_test_flow.sh

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
PORT="${PORT:-18080}"
BASE_URL="http://127.0.0.1:${PORT}"
LOG_FILE="${ROOT_DIR}/.backend_test_flow.log"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[FAIL] Required command not found: $1"
    exit 1
  fi
}

log_step() {
  echo
  echo "========== $1 =========="
}

expect_status() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "[FAIL] ${label}: expected HTTP ${expected}, got ${actual}"
    exit 1
  fi
  echo "[PASS] ${label}: HTTP ${actual}"
}

json_extract() {
  local body="$1"
  local expr="$2"
  python - <<PY
import json
body = '''$body'''
expr = "$expr"
obj = json.loads(body)
cur = obj
for part in expr.split('.'):
    if not part:
        continue
    if isinstance(cur, dict):
        cur = cur.get(part)
    else:
        cur = None
        break
print(cur if cur is not None else "")
PY
}

start_server() {
  log_step "Starting backend server"
  : > "$LOG_FILE"
  (
    cd "$BACKEND_DIR"
    PORT="$PORT" go run ./cmd/api/main.go >>"$LOG_FILE" 2>&1
  ) &
  SERVER_PID=$!
  echo "[INFO] Server PID: $SERVER_PID"

  for _ in $(seq 1 60); do
    code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || true)
    if [[ "$code" == "200" ]]; then
      echo "[PASS] Server is healthy"
      return
    fi
    sleep 1
  done

  echo "[FAIL] Server did not become healthy. Last logs:"
  tail -n 80 "$LOG_FILE" || true
  exit 1
}

stop_server() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" >/dev/null 2>&1 || true
  fi
}

trap stop_server EXIT

# ---------- Preconditions ----------
require_cmd curl
require_cmd go
require_cmd python

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "[FAIL] DATABASE_URL is required"
  exit 1
fi
if [[ -z "${SUPABASE_URL:-}" ]]; then
  echo "[FAIL] SUPABASE_URL is required"
  exit 1
fi

log_step "Pre-flight"
echo "[INFO] BASE_URL=$BASE_URL"
echo "[INFO] SUPABASE_URL=${SUPABASE_URL}"

# ---------- Optional static checks ----------
log_step "Static sanity checks"
(
  cd "$BACKEND_DIR"
  gofmt -l ./cmd ./config ./internal | sed 's/^/[gofmt] /' || true
)

# ---------- Runtime checks ----------
start_server

log_step "Public health checks"
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
expect_status "200" "$code" "GET /health"
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/health")
expect_status "200" "$code" "GET /api/v1/health"

log_step "Auth guard checks"
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/profiles/me/status")
expect_status "401" "$code" "GET /api/v1/profiles/me/status without token"

code=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer invalid.token.value" "$BASE_URL/api/v1/profiles/me/status")
expect_status "401" "$code" "GET /api/v1/profiles/me/status with invalid token"

if [[ -z "${TEST_JWT:-}" ]]; then
  log_step "Authenticated flow skipped"
  echo "[WARN] TEST_JWT not provided; ran public/auth-guard checks only."
  echo "[INFO] To run full flow, set TEST_JWT from Supabase login session."
  exit 0
fi

AUTH_HEADER=( -H "Authorization: Bearer ${TEST_JWT}" )

log_step "Authenticated profile status"
profile_resp=$(curl -s "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/profiles/me/status")
profile_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/profiles/me/status")
expect_status "200" "$profile_code" "GET /api/v1/profiles/me/status"
echo "[INFO] profile response: $profile_resp"

log_step "Diagnostic flow"
start_payload='{"selected_topics":["22222222-2222-2222-2222-222222222221"]}'
start_body=$(curl -s "${AUTH_HEADER[@]}" -H "Content-Type: application/json" -X POST "$BASE_URL/api/v1/diagnostic/start" -d "$start_payload")
start_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" -H "Content-Type: application/json" -X POST "$BASE_URL/api/v1/diagnostic/start" -d "$start_payload")
if [[ "$start_code" == "201" ]]; then
  echo "[PASS] POST /api/v1/diagnostic/start -> 201"
  attempt_id=$(json_extract "$start_body" "data.attempt_id")
  if [[ -z "$attempt_id" ]]; then
    echo "[FAIL] Could not parse diagnostic attempt_id"
    echo "$start_body"
    exit 1
  fi
  echo "[PASS] Diagnostic attempt created: $attempt_id"

  next_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/diagnostic/${attempt_id}/next")
  expect_status "200" "$next_code" "GET /api/v1/diagnostic/{attempt}/next"

  submit_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" -X POST "$BASE_URL/api/v1/diagnostic/${attempt_id}/submit")
  if [[ "$submit_code" != "202" && "$submit_code" != "403" ]]; then
    echo "[FAIL] POST /diagnostic/{attempt}/submit expected 202 or 403, got ${submit_code}"
    exit 1
  fi
  echo "[PASS] POST /diagnostic/{attempt}/submit -> ${submit_code}"
elif [[ "$start_code" == "403" ]]; then
  echo "[WARN] POST /api/v1/diagnostic/start -> 403 (diagnostic already blocked/completed for this user)"
else
  echo "[FAIL] POST /api/v1/diagnostic/start expected 201 or 403, got ${start_code}"
  exit 1
fi

log_step "Course + dashboard access"
course_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/course")
if [[ "$course_code" != "200" && "$course_code" != "403" ]]; then
  echo "[FAIL] GET /api/v1/course expected 200 or 403, got ${course_code}"
  exit 1
fi
echo "[PASS] GET /api/v1/course -> ${course_code}"

dash_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/dashboard/summary")
if [[ "$dash_code" != "200" && "$dash_code" != "403" ]]; then
  echo "[FAIL] GET /api/v1/dashboard/summary expected 200 or 403, got ${dash_code}"
  exit 1
fi
echo "[PASS] GET /api/v1/dashboard/summary -> ${dash_code}"


log_step "Platform connection + sync APIs"
platform_payload='{"platform":"leetcode","handle":"sample_user"}'
platform_connect_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" -H "Content-Type: application/json" -X POST "$BASE_URL/api/v1/profiles/me/platform-connections" -d "$platform_payload")
if [[ "$platform_connect_code" != "202" ]]; then
  echo "[FAIL] POST /api/v1/profiles/me/platform-connections expected 202, got ${platform_connect_code}"
  exit 1
fi
echo "[PASS] POST /api/v1/profiles/me/platform-connections -> ${platform_connect_code}"

platform_list_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/profiles/me/platform-connections")
expect_status "200" "$platform_list_code" "GET /api/v1/profiles/me/platform-connections"

sync_trigger_resp=$(mktemp)
sync_trigger_code=$(curl -s -o "$sync_trigger_resp" -w "%{http_code}" "${AUTH_HEADER[@]}" -X POST "$BASE_URL/api/v1/platform-sync/trigger")
expect_status "202" "$sync_trigger_code" "POST /api/v1/platform-sync/trigger"
sync_trigger_body=$(cat "$sync_trigger_resp")
rm -f "$sync_trigger_resp"
job_id=$(json_extract "$sync_trigger_body" "data.job_id")
if [[ -n "$job_id" ]]; then
  job_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/platform-sync/jobs/${job_id}")
  expect_status "200" "$job_code" "GET /api/v1/platform-sync/jobs/{job_id}"
fi

overview_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/platform-sync/overview")
expect_status "200" "$overview_code" "GET /api/v1/platform-sync/overview"

disconnect_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" -X DELETE "$BASE_URL/api/v1/profiles/me/platform-connections/leetcode")
if [[ "$disconnect_code" != "202" ]]; then
  echo "[FAIL] DELETE /api/v1/profiles/me/platform-connections/leetcode expected 202, got ${disconnect_code}"
  exit 1
fi
echo "[PASS] DELETE /api/v1/profiles/me/platform-connections/leetcode -> ${disconnect_code}"

log_step "IDE sample run + async submission"
run_payload='{"question_id":"55555555-5555-5555-5555-555555555556","language":"python","code":"print(1)"}'
run_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" -H "Content-Type: application/json" -X POST "$BASE_URL/api/v1/ide/run" -d "$run_payload")
if [[ "$run_code" != "200" && "$run_code" != "404" ]]; then
  echo "[FAIL] POST /api/v1/ide/run expected 200 or 404, got ${run_code}"
  exit 1
fi
echo "[PASS] POST /api/v1/ide/run -> ${run_code}"

submit_payload='{"question_id":"55555555-5555-5555-5555-555555555556","language":"python","code":"print(1)"}'
ide_submit_body=$(curl -s "${AUTH_HEADER[@]}" -H "Content-Type: application/json" -X POST "$BASE_URL/api/v1/ide/submit" -d "$submit_payload")
ide_submit_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" -H "Content-Type: application/json" -X POST "$BASE_URL/api/v1/ide/submit" -d "$submit_payload")
if [[ "$ide_submit_code" != "202" && "$ide_submit_code" != "404" ]]; then
  echo "[FAIL] POST /api/v1/ide/submit expected 202 or 404, got ${ide_submit_code}"
  exit 1
fi
echo "[PASS] POST /api/v1/ide/submit -> ${ide_submit_code}"

if [[ "$ide_submit_code" == "202" ]]; then
  sub_id=$(json_extract "$ide_submit_body" "data.submission_id")
  if [[ -n "$sub_id" ]]; then
    echo "[INFO] Polling IDE submission status for $sub_id"
    for _ in $(seq 1 15); do
      status_body=$(curl -s "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/ide/status?id=${sub_id}")
      status_code=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/ide/status?id=${sub_id}")
      if [[ "$status_code" == "200" ]]; then
        eval_status=$(json_extract "$status_body" "data.evaluation_status")
        echo "[INFO] evaluation_status=${eval_status}"
        if [[ "$eval_status" == "completed" || "$eval_status" == "failed" ]]; then
          echo "[PASS] IDE async evaluation reached terminal state"
          break
        fi
      fi
      sleep 2
    done
  fi
fi

log_step "Backend flow test complete"
echo "[PASS] Script completed successfully"
