# Local Clone + Supabase Setup Guide (Backend)

This guide is for running this repo on your own machine and ensuring migrations/seeds actually update your Supabase PostgreSQL schema.

## Why data was not updating before
If backend uses a local/stub DB driver, SQL never reaches Supabase.
This repo is now configured to use the real `pgx` dependency (no local stub replacement).

## 1) Clone and enter backend
```bash
git clone <your-repo-url>
cd arnocodes/backend
```

## 2) Configure environment
Create `backend/.env` with at least:
```env
PORT=8080
ENVIRONMENT=development
DATABASE_URL=postgresql://<user>:<password>@<host>:5432/postgres?sslmode=require
```

For Supabase, use the connection string from **Project Settings → Database**.
Prefer the pooler/IPv4-compatible URL if your ISP/network has IPv6 restrictions.

## 3) Install dependencies (local machine)
```bash
go mod tidy
```

If your corporate network blocks `proxy.golang.org`, try:
```bash
GOPROXY=direct GOSUMDB=off go mod tidy
```

## 4) Start backend
```bash
go run cmd/api/main.go
```

On startup backend does:
1. connect DB
2. run migrations (`internal/database/migrations/*.sql`)
3. run seeds

## 5) Verify migration + seed in Supabase SQL editor
Run these checks:
```sql
SELECT * FROM schema_migrations ORDER BY applied_at DESC;
SELECT id, type FROM tests;
SELECT id, test_id, question_type, order_index FROM questions ORDER BY order_index;
SELECT id, name FROM topics;
SELECT id, topic_id, title FROM subtopics;
```

You should see:
- migration `001_init.sql` present
- diagnostic test seeded
- slides + mcq questions seeded
- sample topics/subtopics seeded

## 6) Re-run safety
Migrations are idempotent via `schema_migrations`.
Seeds use `ON CONFLICT DO NOTHING`.
So restarting backend is safe.

## 7) API-level verification
```bash
curl -s http://localhost:8080/api/v1/health | jq
curl -s http://localhost:8080/api/v1/profiles/me/status -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' | jq
```

## 8) Common troubleshooting

### A) `missing go.sum entry` / dependency errors
Run:
```bash
go mod tidy
```

### B) Cannot download modules (403 from proxy)
Use:
```bash
GOPROXY=direct GOSUMDB=off go mod tidy
```

### C) Supabase permission error
Ensure DB user in `DATABASE_URL` can `CREATE TABLE`, `CREATE TYPE`, and `CREATE EXTENSION`.
If extension creation is restricted, execute required extension setup from Supabase SQL editor first.

### D) No data in Supabase after startup
Check backend logs for migration/seed failures.
Then manually query `schema_migrations` in Supabase.
If empty, connection string likely points to wrong DB/project.

