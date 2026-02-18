# Frontend Integration Helper (Backend-Aligned + Theme System)

This guide is the frontend implementation reference for building ArnoCodes UI against the current backend and schema docs.

Use this with:
- `docs/API_ROADMAP.md`
- `docs/DB_SCHEMA.md`
- `docs/PRODUCT_ALIGNMENT.md`
- `docs/LOCAL_TESTING_WITH_TEST_JS.md`

---

## 1) Current Integration Baseline

- API base URL: `http://localhost:8080`
- API prefix: `/api/v1`
- Auth header in local/dev: `X-User-ID: <uuid>`
- Backend enforces diagnostic lock for protected resources (dashboard/learning)

If diagnostic is incomplete, protected endpoints return:
```json
{ "error": "DIAGNOSTIC_REQUIRED" }
```
with HTTP `403`.

Frontend responsibility:
- render lock states and redirects
- never assume unlock based on local UI state only

---

## 2) Required Theme Tokens (Single Source of Truth)

Create a single token file (example: `src/styles/tokens.css`) and do **not** hardcode color values in components.

### Light Theme
```css
:root,
[data-theme="light"] {
  --bg: #F8FAFC;
  --surface: #FFFFFF;
  --primary: #4F46E5;
  --secondary: #2563EB;
  --accent: #F59E0B;

  --text-primary: #0F172A;
  --text-secondary: #475569;
  --border: #E2E8F0;
}
```

### Dark Theme
```css
[data-theme="dark"] {
  --bg: #020617;
  --surface: #0F172A;
  --primary: #6366F1;
  --secondary: #3B82F6;
  --accent: #FBBF24;

  --text-primary: #F8FAFC;
  --text-secondary: #CBD5E1;
  --border: #1E293B;
}
```

### Usage rules
- `--bg` only for page background
- `--surface` for cards/panels/tables/header container
- `--primary` only for primary CTA
- `--accent` sparse usage for highlights/badges
- Never mix raw hex colors directly in component styles

---

## 3) Typography System

- Font family: `Inter, sans-serif`
- Weights:
  - 400: body
  - 500: labels / secondary actions
  - 600: buttons / table headers
  - 700: headings
- Scale:
  - Page heading: `32px–36px`
  - Section heading: `18px–20px`
  - Body: `14px–16px`
  - Supporting: `12px–13px`

---

## 4) Layout and Component Design Rules

### Container and spacing
- Max content width: `1100px` centered
- Spacing scale: `8 / 16 / 24 / 32 / 48`

### Border radii
- Inputs: `8px`
- Buttons: `10px`
- Tables: `12px`
- Cards/forms: `14px`
- Badges: pill/full rounded

### Components
- Header: `var(--surface)` + bottom border `var(--border)`, no shadow
- Primary button: `var(--primary)`, white text
- Secondary button: transparent + `var(--border)` outline
- Card: `var(--surface)` + `var(--border)`
- Input/form: bg `var(--bg)`, border `var(--border)`
- Table: `var(--surface)` with clear borders

### Responsiveness
- Desktop: max-width layout
- Tablet: reduce columns
- Mobile: single-column, reduced padding, slightly smaller headings
- Use CSS grid auto-fit patterns, no fixed widths/heights

---

## 5) Theme Switching (SSR-safe)

- Control theme with `data-theme="dark"` on root element
- Keep switching logic minimal (class/attribute toggle only)
- Let CSS variables resolve colors
- Avoid JS-heavy per-component color logic

---

## 6) Frontend API Client Contract

Create one reusable client utility:

```ts
type APIEnvelope<T> = { status: string; message: string; data?: T };

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  // attaches Content-Type + X-User-ID
  // parses envelope or { error: string }
  // centralized 4xx/5xx handling
}
```

### Error handling map
- `401 UNAUTHORIZED` → session/user header missing
- `403 DIAGNOSTIC_REQUIRED` → lock dashboard + redirect diagnostic
- `403 DIAGNOSTIC_BLOCKED` → show retry/retake-limit message
- `404 NOT_FOUND` → show empty/fallback state
- `422 UNPROCESSABLE_ENTITY` → form/input validation state

---

## 7) Endpoint Mapping for Frontend Screens

### Profile & status
- `GET /api/v1/profiles/me`
- `PATCH /api/v1/profiles/me`
- `GET /api/v1/profiles/me/status`

### Dashboard (protected)
- `GET /api/v1/dashboard/summary`
- `GET /api/v1/dashboard/heatmap`
- `GET /api/v1/dashboard/leaderboards`

### Learning (protected)
- `GET /api/v1/course/structure`
- `GET /api/v1/topics`
- `GET /api/v1/topics/{topicId}`
- `GET /api/v1/topics/{topicId}/unlock-status`
- `GET /api/v1/subtopics/{subtopicId}`
- `POST /api/v1/learning/questions/{questionId}/complete`
- `POST /api/v1/subtopics/{subtopicId}/complete`

### Diagnostic assessment (stateful)
- `POST /api/v1/diagnostic/start`
- `GET /api/v1/diagnostic/{attemptId}/next`
- `POST /api/v1/diagnostic/{attemptId}/answer`
- `POST /api/v1/diagnostic/{attemptId}/coding`
- `GET /api/v1/diagnostic/{attemptId}/status`
- `POST /api/v1/diagnostic/{attemptId}/submit`

### AI & platform
- `POST /api/v1/ai/query`
- `POST /api/v1/ai/code-helper/step`
- `GET /api/v1/ai/usage`
- `GET /api/v1/profiles/me/platform-connections`
- `POST /api/v1/platform-sync/trigger`

---

## 8) Frontend Folder Structure Suggestion

```txt
src/
  app/
  components/
    ui/                 # Button, Card, Badge, Table, Input
    dashboard/
    learning/
    diagnostic/
  lib/
    api/
      client.ts
      endpoints.ts
      types.ts
  styles/
    tokens.css
    typography.css
    globals.css
```

Rules:
- UI primitives consume tokens only
- Feature components do not declare new spacing/color systems
- Shared API types live in one place (`lib/api/types.ts`)

---

## 9) Diagnostic UX Flow (What to build)

1. Load `GET /profiles/me/status`
2. If `dashboard_unlocked=false`:
   - show locked dashboard cards
   - route to diagnostic start CTA
3. Start diagnostic (`/diagnostic/start`)
4. Render sequential question flow from `/diagnostic/{id}/next`
5. Submit answer by question type (`/answer` for mcq, `/coding` for coding)
6. Poll status (`/status`) as needed
7. Submit attempt (`/submit`)
8. Refresh profile status and unlock dashboard routes

---

## 10) Local QA and handoff checklist

Before frontend handoff:
- Theme tokens file implemented and used by all UI primitives
- Light/dark mode parity verified
- No hardcoded colors/spacing in feature components
- Diagnostic flow tested via `backend/test.js`
- Error states mapped for `401/403/404/422`
- Protected route behavior verified using real API responses

For local backend verification runbook use:
- `docs/LOCAL_TESTING_WITH_TEST_JS.md`
- `docs/API_LOCAL_TESTING.md`
