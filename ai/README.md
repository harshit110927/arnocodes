# AI Service (Python)

Production-safe AI orchestration service for ArnoCodes, implemented fully inside `/ai`.

## Implemented Modes

### 1) Chatbot Mode (`mode=chatbot`)
- Model: `gemini-2.5-flash`
- Per-user limit: **3 requests/hour**
- Behavior:
  - concise, direct answers
  - no follow-up questions
  - no conversational fluff
  - fallback text when context is insufficient
- Enforced output token cap via config.

### 2) Code Helper Mode (`mode=code_helper`)
- Model: `gemini-2.5-flash`
- Per-user limit: **5 requests/day**
- Strict required structure:
  - Problem Understanding
  - Intuition
  - Brute Force Approach
  - Optimization Path
  - Final Optimal Solution
  - Interview Simulation
- Enforced output token cap via config.

## Global Controls
- Global daily cap for all AI calls (`AI_GLOBAL_MAX_PER_DAY`, default `300`)
- Structured errors for:
  - rate limit exceeded
  - provider failure
  - invalid input
  - timeout
- Logs include:
  - `userId`
  - `mode`
  - timestamp
  - input/output token estimates
  - rate-limit hits

## Architecture

```
ai/src/
├── main.py                 # Flask routes + error handlers
├── service.py              # AI orchestration flow
├── mode_handler.py         # Per-mode limits and output token controls
├── rate_limiter.py         # Sliding-window per-user and global limiter
├── prompt_builder.py       # System instructions by mode
├── response_formatter.py   # Strict shaping per mode
├── logging_wrapper.py      # Structured JSON logging
├── errors.py               # Internal service error model
├── config.py               # Environment-backed config
└── providers/
    ├── base.py             # LLMProvider interface
    └── gemini.py           # Gemini 2.5 Flash adapter
```

Provider abstraction interface:
- `LLMProvider.generate(prompt, config)`

This keeps business logic decoupled from Gemini and swappable for future providers.

## API

### Health
- `GET /health`

### Query
- `POST /query`
- Request body:
```json
{
  "userId": "user-123",
  "mode": "chatbot",
  "text": "Explain memoization briefly"
}
```
- `mode` values: `chatbot` | `code_helper`
- Backwards compatibility: `query` field is also accepted when `text` is omitted.

### Index
- `POST /index`
- Returns `501` not implemented placeholder intentionally.

## Environment Variables

Copy `.env.example` to `.env` and set:
- `GEMINI_API_KEY`
- `GEMINI_MODEL` (default `gemini-2.5-flash`)
- rate/token controls as needed.

### How to add Gemini keys safely
1. Pull the repo locally:
   ```bash
   git clone <your-repo-url>
   cd arnocodes/ai
   ```
2. Create your local env file:
   ```bash
   cp .env.example .env
   ```
3. Open `.env` and set your real key (never commit it):
   ```dotenv
   GEMINI_API_KEY=AIza...your_real_key_here
   GEMINI_MODEL=gemini-2.5-flash
   ```
4. Confirm `.env` is ignored (`ai/.gitignore` already ignores local env files). Do not paste keys in PRs, logs, or screenshots.

## Run locally

```bash
cd ai
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python src/main.py
```

## Full local testing pipeline (from fresh pull)

Use this exact sequence on your local machine after pulling latest changes.

### 1) Setup and static sanity
```bash
cd arnocodes/ai
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python3 -m compileall src
```

### 2) Start service
```bash
python src/main.py
```
Service should listen on `http://localhost:5000`.

### 3) Health check
```bash
curl -s http://localhost:5000/health | jq
```
Expected: `{"status":"healthy","service":"ai-service"}`

### 4) Negative-path checks (no key, invalid input)

**A. Missing/invalid payload**
```bash
curl -s -X POST http://localhost:5000/query \
  -H 'Content-Type: application/json' \
  -d '{}' | jq
```
Expected: structured `INVALID_INPUT` error.

**B. Missing GEMINI_API_KEY (or wrong key)**
```bash
curl -s -X POST http://localhost:5000/query \
  -H 'Content-Type: application/json' \
  -d '{"userId":"u1","mode":"chatbot","text":"Hello"}' | jq
```
Expected: structured `PROVIDER_ERROR` without raw Gemini error leakage.

### 5) Chatbot mode functional check (with valid key)
```bash
curl -s -X POST http://localhost:5000/query \
  -H 'Content-Type: application/json' \
  -d '{"userId":"u-chat-1","mode":"chatbot","text":"Explain big-O in one sentence."}' | jq
```
Verify:
- Short/direct response
- `usage.inputTokenEstimate` and `usage.outputTokenEstimate` present

### 6) Code helper mode functional check (with valid key)
```bash
curl -s -X POST http://localhost:5000/query \
  -H 'Content-Type: application/json' \
  -d '{"userId":"u-code-1","mode":"code_helper","text":"Two Sum problem"}' | jq -r '.response'
```
Verify output contains all sections in order:
1. Problem Understanding
2. Intuition
3. Brute Force Approach
4. Optimization Path
5. Final Optimal Solution
6. Interview Simulation

### 7) Rate-limit tests

**A. Chatbot 3/hour per user**
```bash
for i in 1 2 3 4; do
  curl -s -X POST http://localhost:5000/query \
    -H 'Content-Type: application/json' \
    -d '{"userId":"limit-chat-user","mode":"chatbot","text":"ping"}' | jq '.error.code // "ok"'
done
```
Expected: first 3 are `ok`; 4th is `RATE_LIMIT_EXCEEDED`.

**B. Code helper 5/day per user**
```bash
for i in 1 2 3 4 5 6; do
  curl -s -X POST http://localhost:5000/query \
    -H 'Content-Type: application/json' \
    -d '{"userId":"limit-code-user","mode":"code_helper","text":"binary search"}' | jq '.error.code // "ok"'
done
```
Expected: first 5 are `ok`; 6th is `RATE_LIMIT_EXCEEDED`.

### 8) Global daily cap test (optional)
Set temporary low cap in `.env`, restart server:
```dotenv
AI_GLOBAL_MAX_PER_DAY=2
```
Then send 3 valid requests across any users/modes. Third should return `RATE_LIMIT_EXCEEDED` with global-limit message.

### 9) Index endpoint contract
```bash
curl -s -X POST http://localhost:5000/index \
  -H 'Content-Type: application/json' \
  -d '{"document":"test"}' | jq
```
Expected: HTTP `501` and message indicating indexing is intentionally not implemented in this AI mode service.

### 10) Log verification
Observe server logs while running requests. You should see structured JSON events:
- `ai_request`
- `rate_limit_hit`

Each request log should include `userId`, `mode`, timestamp, and token estimates.

## Troubleshooting
- **502 PROVIDER_ERROR**: verify `GEMINI_API_KEY` is present and valid.
- **TIMEOUT**: increase `AI_REQUEST_TIMEOUT_SECONDS`.
- **Unexpected 429 quickly**: verify per-user/global limit env values and restart after changing `.env`.
- **No sectioned output in code_helper**: confirm mode is exactly `code_helper`.
