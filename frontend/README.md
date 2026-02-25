# Frontend (Next.js)

## Description
Frontend application built with Next.js, React, and TypeScript.

## Structure
- `src/app/` - Next.js App Router pages
- `src/components/` - Reusable React components

## Running Locally

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

## Environment Variables
Create `.env.local` file for environment-specific configuration.

## Features
- Next.js 16+ with App Router
- TypeScript support
- React 18+
- Minimal page setup


## Auth Flow
- Supabase client is initialized in `src/lib/supabase.ts` using:
  - `NEXT_PUBLIC_SUPABASE_URL`
  - `NEXT_PUBLIC_SUPABASE_ANON_KEY`
- Login and signup pages live at `/login` and `/signup` (email/password).
- Session persistence is handled by Supabase JS client.
- Route guard (`src/components/AuthGuard.tsx`) redirects unauthenticated users to `/login` for all routes except `/login` and `/signup`.
- API calls should use `src/lib/api.ts`, which forwards `Authorization: Bearer <access_token>` to backend.
- Logout is implemented via `supabase.auth.signOut()` and redirects to `/login`.
