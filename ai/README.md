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

## Run

```bash
cd ai
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python src/main.py
```
