# Profile Sync and Public Rate Limiting Guide

## 1. Profile Sync Status
Current implementation supports:
- Connect platform handles
- List connections
- Disconnect handles
- Trigger sync job and query job status

Supported platforms:
- `leetcode`
- `gfg`
- `codeforces`
- `hackerrank`
- `codechef`

## 2. API Endpoints
- `GET /api/v1/profiles/me/platform-connections`
- `POST /api/v1/profiles/me/platform-connections`
- `DELETE /api/v1/profiles/me/platform-connections/{platform}`
- `POST /api/v1/platform-sync/trigger`
- `GET /api/v1/platform-sync/jobs/{job_id}`
- `GET /api/v1/platform-sync/overview`

## 3. Connect Request Format
```json
{
  "platform": "leetcode",
  "handle": "my_handle"
}
```

## 4. Sync Job Behavior
Triggering sync currently creates/claims/completes a sync job transactionally. The API and job lifecycle are in place for platform ingestion workflows.

Retry policy for trigger attempts: 3 retries with exponential backoff (1s, 2s, 4s). If all retries fail, API returns a temporary failure response asking the user to retry after 5–10 seconds.

## 5. Rate Limiting Strategy
Rate limiting is intentionally expected at deployment ingress (API gateway / CDN / load balancer), not in backend process memory, to avoid shared-network false positives.

## 6. Ingress Policy Guidance
Recommended initial controls at edge:
- stricter limits on `POST /api/v1/ide/submit` and auth-sensitive endpoints
- moderate limits on read APIs
- temporary block on repeated abusive bursts

## 7. Testing Profile Sync APIs
Use:
```bash
DATABASE_URL=postgres://... \
SUPABASE_URL=https://<project>.supabase.co \
TEST_JWT=<access_token> \
./backend/scripts/full_backend_test_flow.sh
```
The script validates profile connect/list/disconnect and platform sync trigger/job/overview APIs.
