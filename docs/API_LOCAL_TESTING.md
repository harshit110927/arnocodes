# API Local Testing Playbook (No DB / Skeleton Mode)

This is a beginner-friendly, command-first guide to test every current backend API locally.

## 1) Start backend locally
```bash
cd backend
go run cmd/api/main.go
```

Base URL: `http://localhost:8080`  
Version prefix: `/api/v1`

## 2) Quick sanity checks
```bash
curl -s http://localhost:8080/api/v1/health | jq
curl -s http://localhost:8080/api/v1/internal/api-catalog | jq
curl -s -X POST http://localhost:8080/api/v1/internal/api-smoke-check | jq
```

## 3) Profile APIs
```bash
curl -s http://localhost:8080/api/v1/profiles/me | jq
curl -s -X PATCH http://localhost:8080/api/v1/profiles/me \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Alice","college":"ABC College","graduation_year":2027}' | jq
curl -s http://localhost:8080/api/v1/profiles/me/platform-connections | jq
curl -s -X POST http://localhost:8080/api/v1/profiles/me/platform-connections \
  -H 'Content-Type: application/json' \
  -d '{"platform":"leetcode","platform_handle":"alice_dev"}' | jq
curl -s -X DELETE http://localhost:8080/api/v1/profiles/me/platform-connections/leetcode | jq
```

## 4) Dashboard APIs
```bash
curl -s http://localhost:8080/api/v1/dashboard/summary | jq
curl -s 'http://localhost:8080/api/v1/dashboard/heatmap?from=2026-01-01&to=2026-01-31' | jq
curl -s 'http://localhost:8080/api/v1/dashboard/leaderboards?scope=global&window=weekly' | jq
```

## 5) Learning APIs
```bash
curl -s http://localhost:8080/api/v1/course/structure | jq
curl -s http://localhost:8080/api/v1/topics | jq
curl -s http://localhost:8080/api/v1/topics/topic-1 | jq
curl -s http://localhost:8080/api/v1/topics/topic-1/unlock-status | jq
curl -s http://localhost:8080/api/v1/subtopics/subtopic-1 | jq
curl -s -X POST http://localhost:8080/api/v1/learning/questions/question-1/complete | jq
```

Subtopic completion validation:
```bash
curl -s -X POST http://localhost:8080/api/v1/subtopics/subtopic-1/complete \
  -H 'Content-Type: application/json' \
  -d '{"mastery_score":0.5}' | jq

curl -s -X POST http://localhost:8080/api/v1/subtopics/subtopic-1/complete \
  -H 'Content-Type: application/json' \
  -d '{"mastery_score":0.9}' | jq
```

## 6) Assessment APIs (stateful flow)
Use a fixed user id header to simulate login state:
```bash
USER_ID=demo-user
```

Step A: Fetch diagnostic test for selected topics
```bash
curl -s 'http://localhost:8080/api/v1/tests/diagnostic-1?topics=arrays,strings' \
  -H "X-User-ID: $USER_ID" | jq
```

Step B: Start attempt with selected topics
```bash
START=$(curl -s -X POST http://localhost:8080/api/v1/tests/diagnostic-1/start \
  -H "X-User-ID: $USER_ID" \
  -H 'Content-Type: application/json' \
  -d '{"topics_known":["arrays","strings"]}')

echo "$START" | jq
ATTEMPT_ID=$(echo "$START" | jq -r '.data.attempt_id')
```

Step C: Load session
```bash
curl -s "http://localhost:8080/api/v1/tests/diagnostic-1/session?attempt_id=$ATTEMPT_ID" \
  -H "X-User-ID: $USER_ID" | jq
```

Step D: Get next question and submit answer
```bash
NEXT=$(curl -s "http://localhost:8080/api/v1/test-attempts/$ATTEMPT_ID/next-question" \
  -H "X-User-ID: $USER_ID")
echo "$NEXT" | jq
QUESTION_ID=$(echo "$NEXT" | jq -r '.data.question.id')

curl -s -X POST "http://localhost:8080/api/v1/test-attempts/$ATTEMPT_ID/answers" \
  -H "X-User-ID: $USER_ID" \
  -H 'Content-Type: application/json' \
  -d "{\"question_id\":\"$QUESTION_ID\",\"selected_option\":2}" | jq
```

Step E: Submit attempt and read result
```bash
curl -s -X POST "http://localhost:8080/api/v1/test-attempts/$ATTEMPT_ID/submit" \
  -H "X-User-ID: $USER_ID" | jq

curl -s "http://localhost:8080/api/v1/test-attempts/$ATTEMPT_ID/result" \
  -H "X-User-ID: $USER_ID" | jq
```

Optional expire/resume:
```bash
curl -s -X POST "http://localhost:8080/api/v1/test-attempts/$ATTEMPT_ID/expire" \
  -H "X-User-ID: $USER_ID" | jq

curl -s -X POST "http://localhost:8080/api/v1/test-attempts/$ATTEMPT_ID/resume" \
  -H "X-User-ID: $USER_ID" | jq
```

## 7) Platform sync APIs
```bash
curl -s -X POST http://localhost:8080/api/v1/platform-sync/trigger | jq
curl -s http://localhost:8080/api/v1/platform-sync/jobs/job-1 | jq
```

## 8) AI APIs
```bash
curl -s -X POST http://localhost:8080/api/v1/ai/query \
  -H 'Content-Type: application/json' \
  -d '{"topic":"graphs","query":"Explain BFS"}' | jq
curl -s -X POST http://localhost:8080/api/v1/ai/code-helper/step \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"s1","question":"two sum","step":1}' | jq
curl -s http://localhost:8080/api/v1/ai/usage | jq
```

## 9) Internal jobs (dev only)
```bash
curl -s -X POST http://localhost:8080/api/v1/internal/recompute-dashboard | jq
curl -s -X POST http://localhost:8080/api/v1/internal/refresh-leaderboard | jq
```

## 10) What this validates
- route registration
- method restrictions
- stateful assessment flow behavior
- response envelope shape

It does not validate DB persistence yet.
