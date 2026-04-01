# ArnoCodes Monorepo

A modern monorepo containing Backend (Go), Frontend (Next.js), and AI (Python) services.

## Project Structure

```
arnocodes/
├── backend/          # Go API service
│   ├── cmd/
│   │   └── api/      # Application entry point
│   ├── internal/     # Internal packages
│   │   ├── handlers/ # HTTP handlers
│   │   └── database/ # Database connections
│   ├── config/       # Configuration management
│   └── Dockerfile
├── frontend/         # Next.js application
│   ├── src/
│   │   ├── app/      # Next.js App Router pages
│   │   └── components/
│   └── Dockerfile
├── ai/               # Python AI service
│   ├── src/
│   │   └── main.py   # Flask application with RAG service
│   └── Dockerfile
└── docker-compose.yml
```

## Services

### Backend (Go)
- **Port:** 8080
- **Tech Stack:** Go, HTTP server
- **Features:**
  - Clean modular architecture (cmd/api, internal, config)
  - Health check endpoint (`/health`)
  - Environment-based configuration
  - PostgreSQL and Redis connection placeholders

### Frontend (Next.js)
- **Port:** 3000
- **Tech Stack:** Next.js, React, TypeScript
- **Features:**
  - Next.js 14+ with App Router
  - TypeScript support
  - Minimal page setup

### AI (Python)
- **Port:** 5000
- **Tech Stack:** Python, Flask
- **Features:**
  - RAG (Retrieval-Augmented Generation) service placeholder
  - Health check endpoint
  - Query and document indexing endpoints

## Quick Start

### Using Docker Compose (Recommended)

First, copy the environment file and customize if needed:
```bash
cp .env.example .env
```

Then start all services:
```bash
# Start all services
docker compose up -d

# Stop all services
docker compose down

# View logs
docker compose logs -f
```

Services will be available at:
- Backend: http://localhost:8080
- Frontend: http://localhost:3000
- AI Service: http://localhost:5000
- PostgreSQL: localhost:5432
- Redis: localhost:6379

### Running Services Individually

#### Backend
```bash
cd backend
go mod download
go run cmd/api/main.go
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

#### AI Service
```bash
cd ai
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
python src/main.py
```


## Run Entire Project Locally (Backend + Frontend + AI)

### 1) Start infrastructure dependencies
Use Docker for Postgres and Redis:

```bash
docker compose up -d postgres redis
```

### 2) Start backend

```bash
cd backend
go mod download
go run cmd/api/main.go
```

Backend runs at `http://localhost:8080`.

### 3) Start AI service

```bash
cd ai
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python src/main.py
```

AI service runs at `http://localhost:5000`.

### 4) Start frontend
Create `frontend/.env.local` with Supabase and API settings:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_SUPABASE_URL=YOUR_SUPABASE_URL
NEXT_PUBLIC_SUPABASE_ANON_KEY=YOUR_SUPABASE_ANON_KEY
```

Then run:

```bash
cd frontend
npm install
npm run dev
```

Frontend runs at `http://localhost:3000`.

### 5) Validate local API behavior

```bash
curl -s http://localhost:8080/api/v1/health | jq
curl -s http://localhost:8080/api/v1/profiles/me/status -H 'X-User-ID: 00000000-0000-0000-0000-000000000001' | jq
```

For complete local testing flows, use:
- `docs/API_LOCAL_TESTING.md`
- `docs/LOCAL_TESTING_WITH_TEST_JS.md`
- `docs/FRONTEND_IMPLEMENTATION_CONTEXT.md`

## Environment Variables

Each service has an `.env.example` file. Copy it to `.env` and update the values:

- **Backend:** `backend/.env`
- **AI Service:** `ai/.env`
- **Frontend:** `frontend/.env.local` (if needed)

## API Endpoints

### Backend
- `GET /health` - Health check

### AI Service
- `GET /health` - Health check
- `POST /query` - Query RAG service
- `POST /index` - Index a document

## Development

Each service has its own README with detailed setup instructions:
- [Backend README](./backend/README.md)
- [Frontend README](./frontend/README.md)
- [AI Service README](./ai/README.md)

## Product Documentation

- [Product Overview](./docs/PRODUCT.md)
- [Database Schema](./docs/DB_SCHEMA.md)
- [Development Cycle](./docs/DEVELOPMENT_CYCLE.md)
- [Product Alignment Review](./docs/PRODUCT_ALIGNMENT.md)
- [API Roadmap](./docs/API_ROADMAP.md)
- [Frontend Helper Guide](./docs/FRONTEND_HELPER.md)
- [Assessment Engine Guide](./docs/ASSESSMENT_ENGINE.md)
- [API Local Testing Playbook](./docs/API_LOCAL_TESTING.md)
- [Codebase Guide](./docs/CODEBASE_GUIDE.md)
- [Local Supabase Setup Guide](./docs/LOCAL_SUPABASE_SETUP.md)
- [Local Script Testing Guide (test.js)](./docs/LOCAL_TESTING_WITH_TEST_JS.md)

## License

This project is licensed under the terms specified in the LICENSE file.
