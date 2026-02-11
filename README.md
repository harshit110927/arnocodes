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

## License

This project is licensed under the terms specified in the LICENSE file.
