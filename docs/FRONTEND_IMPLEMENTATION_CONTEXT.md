# Frontend Implementation Context and Debugging Guide

## What was implemented

The Next.js frontend now uses the `frontend/reference.html` dashboard style as the visual foundation and maps real backend API responses into dashboard cards, topic lists, activity, weak topics, and platform-connection workflows.

## API contracts used

All calls are authenticated through Supabase session access token and sent as `Authorization: Bearer <token>`.

- `GET /api/v1/profiles/me/status`
- `GET /api/v1/dashboard`
- `GET /api/v1/course/structure`
- `GET /api/v1/profiles/me/platform-connections`
- `POST /api/v1/profiles/me/platform-connections`
- `POST /api/v1/platform-sync/trigger`

## Error handling behavior

Centralized error handling is in `frontend/src/lib/api.ts`:

- Missing auth token redirects to `/login` and throws `UNAUTHORIZED`.
- Non-2xx HTTP responses throw a structured `APIError` with:
  - `status`
  - `code` (if backend returned `{ error: string }`)
  - `payload` (raw parsed response)

Current page-level handling in `frontend/src/app/page.tsx`:

- `DIAGNOSTIC_REQUIRED` marks dashboard as locked and renders lock card.
- Other errors show a compact HTTP-status message.

## Frontend state loading sequence

1. Fetch profile status.
2. Fetch course structure and platform connections in parallel.
3. If status indicates unlocked dashboard, fetch dashboard payload.
4. Render summary cards + weak topics + recent activity.

## Debugging checklist

### 1) Verify auth first

- Login via `/login`.
- In browser devtools, ensure Supabase session exists and contains `access_token`.

### 2) Verify API base URL

In `frontend/.env.local`:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_SUPABASE_URL=...
NEXT_PUBLIC_SUPABASE_ANON_KEY=...
```

### 3) Verify backend and DB availability

- Backend must run on `:8080`.
- DB must include migrated and seeded data.

### 4) Diagnose `DIAGNOSTIC_REQUIRED`

This is expected before a user completes diagnostic. Run diagnostic flow with API playbooks in:

- `docs/API_LOCAL_TESTING.md`
- `docs/LOCAL_TESTING_WITH_TEST_JS.md`

### 5) Diagnose platform sync/connect issues

- Validate platform key in allowed list (`leetcode`, `gfg`, `codeforces`, `hackerrank`, `codechef`).
- Confirm user has a valid profile row.

## Notes on styling and reference parity

The app pulls the multi-theme visual token approach from `frontend/reference.html` and uses a similar shell:

- Topbar
- Sidebar
- Card-grid dashboard
- Accent-based status and metric display

Theme switching is now live in the topbar and toggles the `data-theme` attribute on the document root.
