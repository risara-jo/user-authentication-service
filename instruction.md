# Go + Gin + PostgreSQL Docker Development Guide

## Project Structure
```
project-name/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   └── health.go
│   ├── models/
│   │   └── user.go
│   └── database/
│       └── postgres.go
├── pkg/
│   └── utils/
├── migrations/
│   └── 001_initial.sql
├── docker/
│   ├── Dockerfile
│   └── Dockerfile.dev
├── docker-compose.yml
├── docker-compose.dev.yml
├── .env.example
├── .env
├── .dockerignore
├── go.mod
└── go.sum
```

## 1. Docker Configuration

### Dockerfile (Production)
```dockerfile
# docker/Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Expose port
EXPOSE 8080

CMD ["./main"]
```

### Dockerfile.dev (Development)
```dockerfile
# docker/Dockerfile.dev
FROM golang:1.25-alpine

WORKDIR /app

# Install air for hot reload
RUN go install github.com/cosmtrek/air@latest

# Install postgresql-client for migrations
RUN apk add --no-cache postgresql-client

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
```

### .dockerignore
```
.git
.gitignore
README.md
Dockerfile*
.dockerignore
node_modules
npm-debug.log
.env
.env.local
```

## 2. Docker Compose Configuration

### docker-compose.yml (Production)
```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=password
      - DB_NAME=myapp
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - app-network

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: myapp
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

volumes:
  postgres_data:

networks:
  app-network:
    driver: bridge
```

### docker-compose.dev.yml (Development)
```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: docker/Dockerfile.dev
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=password
      - DB_NAME=myapp_dev
    volumes:
      - .:/app
      - go_modules:/go/pkg/mod
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - app-network

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: myapp_dev
    ports:
      - "5432:5432"
    volumes:
      - postgres_dev_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

volumes:
  postgres_dev_data:
  go_modules:

networks:
  app-network:
    driver: bridge
```

## 3. Air Configuration for Hot Reload

### .air.toml
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main cmd/api/main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

## 4. Go Application Setup

### go.mod
```go
module your-project-name

go 1.25

require (
    github.com/gin-gonic/gin v1.10.0
    github.com/lib/pq v1.10.9
    github.com/joho/godotenv v1.5.1
    gorm.io/gorm v1.25.12
    gorm.io/driver/postgres v1.5.9
)
```

### cmd/api/main.go
```go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "your-project-name/internal/config"
    "your-project-name/internal/database"
    "your-project-name/internal/handlers"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // Load config
    cfg := config.Load()

    // Initialize database
    db, err := database.NewPostgresDB(cfg)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Setup Gin router
    r := gin.Default()

    // Health check endpoint
    r.GET("/health", handlers.HealthCheck)

    // API routes
    api := r.Group("/api/v1")
    {
        api.GET("/ping", func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "pong"})
        })
    }

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

### internal/config/config.go
```go
package config

import "os"

type Config struct {
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    Port       string
}

func Load() *Config {
    return &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", "password"),
        DBName:     getEnv("DB_NAME", "myapp"),
        Port:       getEnv("PORT", "8080"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### internal/database/postgres.go
```go
package database

import (
    "fmt"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "your-project-name/internal/config"
)

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    return db, nil
}
```

### internal/handlers/health.go
```go
package handlers

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "message": "Service is running",
    })
}
```

## 5. Environment Configuration

### .env.example
```
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=myapp_dev

# Server
PORT=8080
GIN_MODE=debug
```

### .env (copy from .env.example and modify as needed)

## 6. Database Migration

### migrations/001_initial.sql
```sql
-- Initial database schema
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

## 7. Development Commands

### Getting Started
```bash
# Clone the repository
git clone <repository-url>
cd project-name

# Copy environment file
cp .env.example .env

# Start development environment
docker-compose -f docker-compose.dev.yml up --build

# Or run in background
docker-compose -f docker-compose.dev.yml up --build -d
```

### Useful Commands
```bash
# View logs
docker-compose -f docker-compose.dev.yml logs -f app

# Stop services
docker-compose -f docker-compose.dev.yml down

# Remove volumes (reset database)
docker-compose -f docker-compose.dev.yml down -v

# Execute commands in container
docker-compose -f docker-compose.dev.yml exec app go mod tidy
docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev

# Run tests
docker-compose -f docker-compose.dev.yml exec app go test ./...

# Build production image
docker-compose build

# Run production
docker-compose up -d
```

### Database Operations
```bash
# Connect to database
docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev

# Run migrations manually
docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev -f /docker-entrypoint-initdb.d/001_initial.sql

# Backup database
docker-compose -f docker-compose.dev.yml exec postgres pg_dump -U postgres myapp_dev > backup.sql
```

## 8. Team Development Workflow

### Daily Development
1. **Start development environment:**
   ```bash
   docker-compose -f docker-compose.dev.yml up -d
   ```

2. **Make code changes** - Hot reload will automatically restart the server

3. **Test your changes:**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/api/v1/ping
   ```

4. **Stop when done:**
   ```bash
   docker-compose -f docker-compose.dev.yml down
   ```

### Adding Dependencies
```bash
# Add new dependency
docker-compose -f docker-compose.dev.yml exec app go get github.com/new/package

# Update go.mod and go.sum
docker-compose -f docker-compose.dev.yml exec app go mod tidy
```

### Code Quality
```bash
# Format code
docker-compose -f docker-compose.dev.yml exec app go fmt ./...

# Run linter (install golangci-lint in Dockerfile.dev if needed)
docker-compose -f docker-compose.dev.yml exec app golangci-lint run

# Run tests with coverage
docker-compose -f docker-compose.dev.yml exec app go test -v -cover ./...
```

## 9. Production Deployment

### Build and Deploy
```bash
# Build production image
docker build -f docker/Dockerfile -t your-app:latest .

# Run production stack
docker-compose up -d

# Check logs
docker-compose logs -f app
```

## 10. Working with Your Database Script

### Step-by-Step Integration

1. **Copy your script to migrations folder:**
   ```bash
   cp your_database_script.sql migrations/002_your_schema.sql
   ```

2. **Review your script for Docker compatibility:**
   - Remove any `CREATE DATABASE` statements (Docker creates the DB)
   - Add `IF NOT EXISTS` to table creation if not present
   - Wrap in transaction block if needed

3. **Test the migration:**
   ```bash
   # Start fresh database
   docker-compose -f docker-compose.dev.yml down -v
   docker-compose -f docker-compose.dev.yml up -d
   
   # Check if your tables were created
   docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev -c "\dt"
   ```

4. **If your script has issues:**
   ```bash
   # View PostgreSQL logs
   docker-compose -f docker-compose.dev.yml logs postgres
   
   # Connect and debug manually
   docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev
   ```

### Common Script Modifications Needed

**If your script has:**
- `CREATE DATABASE myapp;` → Remove this line
- `USE myapp;` → Remove this line (PostgreSQL doesn't use USE)
- `CREATE TABLE table_name (...)` → Change to `CREATE TABLE IF NOT EXISTS table_name (...)`

**For large scripts:**
```sql
-- Add at the beginning of your script
\timing on
\echo 'Starting database migration...'

-- Your existing script content here

\echo 'Database migration completed!'
```

### Testing Your Database Integration

After setting up your script, verify it works:

```bash
# 1. Start the application
docker-compose -f docker-compose.dev.yml up --build

# 2. Check if your API can connect to DB with your schema
curl http://localhost:8080/health

# 3. Connect to DB and verify your tables
docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev -c "
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public';"
```

## 11. Troubleshooting

### Common Issues

**Container won't start:**
- Check Docker logs: `docker-compose -f docker-compose.dev.yml logs app`
- Verify environment variables in `.env`

**Database connection failed:**
- Ensure PostgreSQL container is healthy: `docker-compose -f docker-compose.dev.yml ps`
- Check database credentials in `.env`

**Hot reload not working:**
- Ensure volume mounting is correct in docker-compose.dev.yml
- Check .air.toml configuration

**Port already in use:**
- Change ports in docker-compose files
- Kill existing processes: `sudo lsof -ti:8080 | xargs kill -9`

### Performance Tips
- Use multi-stage builds for smaller production images
- Leverage Docker layer caching
- Use `.dockerignore` to exclude unnecessary files
- Mount Go modules as volumes for faster rebuilds

## 11. Next Steps

1. **Add authentication middleware**
2. **Implement API versioning**
3. **Add request validation**
4. **Setup logging and monitoring**
5. **Add unit and integration tests**
6. **Configure CI/CD pipeline**
7. **Setup database migrations tool (golang-migrate)**
8. **Add API documentation (Swagger)**

## Quick Reference

| Command | Description |
|---------|-------------|
| `docker-compose -f docker-compose.dev.yml up --build` | Start development |
| `docker-compose -f docker-compose.dev.yml down` | Stop development |
| `docker-compose -f docker-compose.dev.yml logs -f app` | View app logs |
| `docker-compose -f docker-compose.dev.yml exec app go mod tidy` | Update dependencies |
| `docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev` | Connect to DB |

The application will be available at `http://localhost:8080` with hot reload enabled for development.