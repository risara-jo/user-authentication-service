# Smart Transit System

A Go + Gin + PostgreSQL Docker development environment for a smart transit system.

## Project Structure

```
test-dev-env/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handlers/
│   │   └── health.go            # HTTP handlers
│   ├── models/
│   │   └── user.go              # Data models
│   └── database/
│       └── postgres.go          # Database connection
├── pkg/
│   └── utils/                   # Shared utilities
├── migrations/
│   ├── 001_initial.sql          # Initial schema
│   └── 002_init_db.sql          # Complete smart transit schema
├── docker/
│   ├── Dockerfile               # Production Docker image
│   └── Dockerfile.dev           # Development Docker image
├── docker-compose.yml           # Production docker-compose
├── docker-compose.dev.yml       # Development docker-compose
├── .env.example                 # Environment variables template
├── .env                         # Environment variables
├── .dockerignore               # Docker ignore file
├── .air.toml                   # Hot reload configuration
├── go.mod                      # Go module file
├── go.sum                      # Go dependencies
└── README.md                   # This file
```

## Quick Start

1. **Start development environment:**
   ```bash
   docker-compose -f docker-compose.dev.yml up --build
   ```

2. **Test the API:**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/api/v1/ping
   ```

3. **Connect to database:**
   ```bash
   docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev
   ```

4. **View tables:**
   ```sql
   \dt
   ```

5. **Stop services:**
   ```bash
   docker-compose -f docker-compose.dev.yml down
   ```

## Features

- **Hot Reload**: Code changes automatically reload the server
- **PostgreSQL**: Full Smart Transit System database schema
- **Docker**: Containerized development and production environments
- **GORM**: Go ORM for database operations
- **Gin**: Fast HTTP web framework
- **Environment Variables**: Configurable via .env files

## Database

The project includes a comprehensive Smart Transit System database schema with:

- Users and authentication (Firebase integration ready)
- Bus owners and companies
- Buses and routes management
- Booking and payment systems
- Real-time tracking
- Notifications and feedback
- Maintenance records
- Audit logs

## Development Workflow

1. Make code changes (hot reload will restart the server)
2. Add new dependencies with: `docker-compose -f docker-compose.dev.yml exec app go get <package>`
3. Update go.mod: `docker-compose -f docker-compose.dev.yml exec app go mod tidy`
4. Run tests: `docker-compose -f docker-compose.dev.yml exec app go test ./...`

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/ping` - Simple ping endpoint

## Environment Variables

See `.env.example` for all available configuration options.

## Production Deployment

```bash
# Build production image
docker build -f docker/Dockerfile -t smart-transit-system:latest .

# Run production stack
docker-compose up -d
```
