# SUPABASE_AUTH_SETUP

## 1) Enable Email provider
1. Open Supabase Dashboard.
2. Go to **Authentication → Providers**.
3. Enable **Email** provider.

## 2) Disable confirm email (optional for MVP)
1. Go to **Authentication → Providers → Email**.
2. Turn off **Confirm email** for faster local iteration.

## 3) Copy project URL
1. Go to **Settings → API**.
2. Copy **Project URL**.

## 4) Copy anon key
1. In **Settings → API**, copy **anon public** key.

## 5) Copy service role key
1. In **Settings → API**, copy **service_role** key.
2. Keep this server-side only; never expose in frontend.

## 6) Environment variables

### Backend (`backend/.env`)
```env
PORT=8080
DATABASE_URL=postgres://...
SUPABASE_URL=https://<project-ref>.supabase.co
SUPABASE_AUDIENCE=authenticated
SUPABASE_SERVICE_ROLE_KEY=<service-role-key>
```

### Frontend (`frontend/.env.local`)
```env
NEXT_PUBLIC_SUPABASE_URL=https://<project-ref>.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=<anon-key>
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

## 7) Run SQL for RLS
1. Open **SQL Editor** in Supabase.
2. Run migration SQL from:
   - `backend/internal/database/migrations/008_enable_rls_user_owned_tables.sql`
3. Confirm migration applied.

## 8) Test RLS policies
1. Create two users (`user_a`, `user_b`).
2. Insert user-owned records for `user_a`.
3. Query using `user_b` JWT in SQL editor or API path.
4. Verify `user_b` cannot read/write `user_a` records.

## 9) Verify JWT in Supabase dashboard
1. Open **Authentication → Users** and copy a user ID.
2. Login via frontend (`/login`) and inspect access token.
3. Decode token at jwt.io (or local tooling) and verify:
   - `iss = https://<project-ref>.supabase.co/auth/v1`
   - `aud = authenticated`
   - `sub = <user-id>`
   - `exp` is valid
4. Call a protected backend route with token and confirm 200.
5. Call same route with missing/invalid token and confirm 401.
