# Backend (Go)

## Description
Backend service built with Go, featuring a clean modular architecture.

## Structure
- `cmd/api/` - Application entry point
- `internal/` - Internal packages
  - `handlers/` - HTTP handlers
  - `database/` - Database connections (PostgreSQL and Redis placeholders)
- `config/` - Configuration management

## Running Locally

```bash
# Install dependencies
go mod download

# Run the server
go run cmd/api/main.go
```

## Environment Variables
Copy `.env.example` to `.env` and update the values:
- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `ENVIRONMENT` - Environment name (development/production)

## Endpoints
- `GET /health` - Health check endpoint
