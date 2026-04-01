# Frontend (Next.js)

## Description
Frontend dashboard built with Next.js, React, and TypeScript. The UI is now aligned to `frontend/reference.html` and connected to real backend endpoints.

## Structure
- `src/app/` - App Router pages
- `src/components/` - Reusable components
- `src/lib/api.ts` - Authenticated API client + endpoint wrappers

## Environment Variables
Create `frontend/.env.local`:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_SUPABASE_URL=YOUR_SUPABASE_URL
NEXT_PUBLIC_SUPABASE_ANON_KEY=YOUR_SUPABASE_ANON_KEY
```

## Running Locally

```bash
npm install
npm run dev
```

Open `http://localhost:3000`.

## Backend endpoints used by frontend
- `GET /api/v1/profiles/me/status`
- `GET /api/v1/dashboard`
- `GET /api/v1/course/structure`
- `GET /api/v1/profiles/me/platform-connections`
- `POST /api/v1/profiles/me/platform-connections`
- `POST /api/v1/platform-sync/trigger`

## Debugging
For implementation context and debugging runbook, see:
- `docs/FRONTEND_IMPLEMENTATION_CONTEXT.md`
- `docs/API_LOCAL_TESTING.md`
- `docs/LOCAL_TESTING_WITH_TEST_JS.md`
