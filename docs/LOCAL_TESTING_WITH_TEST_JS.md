# Local API Testing Guide with `backend/test.js`

This guide helps you run the backend locally and validate the DB-backed diagnostic flow end-to-end using the provided script.

---

## 1) Prerequisites

- Go installed (for backend)
- Node.js 18+ installed (for `fetch` support in Node runtime)
- A PostgreSQL/Supabase database URL

---

## 2) Clone and setup

```bash
git clone <your-repo-url>
cd arnocodes/backend
```

Install Go dependencies (on your local machine):

```bash
go mod tidy
```

---

## 3) Environment setup

Create or update `backend/.env`:

```env
DATABASE_URL=postgres://<user>:<password>@<host>:5432/<db>?sslmode=require
PORT=8080
ENVIRONMENT=local
```

> Use your **Postgres connection string** from Supabase, not the anon HTTP URL.

---

## 4) Start backend

```bash
cd backend
go run cmd/api/main.go
```

On startup, backend will:
1. Connect DB
2. Run migrations (`001_init.sql`, `002_diagnostic_tables.sql`)
3. Run seed data

---

## 5) Ensure test user exists

The script uses a `USER_ID` header. That user must exist in `auth.users` and `profiles`.

Run in Supabase SQL editor:

```sql
INSERT INTO auth.users(id)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa')
ON CONFLICT DO NOTHING;

INSERT INTO profiles(id, full_name)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Local Test User')
ON CONFLICT DO NOTHING;
```

---

## 6) Run the API script

In a second terminal:

```bash
cd backend
node test.js
```

Optional custom config:

```bash
API_BASE=http://localhost:8080 \
USER_ID=aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa \
node test.js
```

---

## 7) What `test.js` verifies

- Health endpoint availability
- Profile status endpoint
- Dashboard lock behavior before diagnostic submit
- Diagnostic start (`/api/v1/diagnostic/start`)
- Question fetch loop (`/next`)
- MCQ/coding submissions (`/answer`, `/coding`)
- Attempt status (`/status`)
- Final submit (`/submit`)
- Dashboard unlock after successful submit
- Safe question payload (no `correct_option` returned by API)

---

## 8) Common issues and fixes

### A) `DIAGNOSTIC_BLOCKED`

You have reached retake limit (2 attempts in 48h) for that user.

**Fix:** use a fresh `USER_ID` and create corresponding rows in `auth.users` + `profiles`.

### B) Dashboard still locked after submit

Check:

```sql
SELECT id, user_id, status, submitted_at
FROM test_attempts
WHERE user_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
ORDER BY started_at DESC;
```

Ensure the latest attempt is `submitted`.

### C) Migrations not applied

Check:

```sql
SELECT * FROM schema_migrations ORDER BY applied_at DESC;
```

You should see `001_init.sql` and `002_diagnostic_tables.sql`.

### D) Seed data missing

Check:

```sql
SELECT id, type FROM tests;
SELECT id, question_type, order_index FROM questions ORDER BY order_index;
SELECT id, name FROM topics;
```

---

## 9) Suggested local workflow

1. Pull latest code
2. Run backend
3. Run `node test.js`
4. Fix issue
5. Re-run `node test.js` until green
6. Then proceed with frontend integration

This gives a repeatable smoke test during active backend development.
