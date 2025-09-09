# 🚌 Smart Transit System

A comprehensive bus booking and management system built with Go, PostgreSQL, and Docker.

## 🚀 Quick Start

### Prerequisites
- Docker Desktop (includes Docker Compose)
- Git

### One-Command Setup

**Windows:**
```bash
setup.bat
```

**Linux/macOS:**
```bash
chmod +x setup.sh
./setup.sh
```

**Manual Setup:**
```bash
# Clone and navigate to project
git clone <your-repo-url>
cd test-dev-env

# Start the system
docker-compose -f docker-compose.dev.yml up --build
```

## 🌐 Access Points

- **Application**: http://localhost:8080
- **PostgreSQL**: localhost:5432
  - Database: `myapp_dev`
  - Username: `postgres`
  - Password: `password`

## 📊 Database Schema

The system includes a complete database schema with:

### Core Tables
- **Users**: Firebase-integrated user management
- **Bus Owners**: Company/owner registration
- **Buses**: Vehicle management
- **Staff**: Driver and conductor management
- **Routes & Stops**: Route planning and stop management
- **Schedules & Trips**: Trip scheduling and execution
- **Bookings**: Passenger reservations
- **Payments**: Payment processing and tracking

### Advanced Features
- **Real-time Tracking**: GPS location tracking
- **Digital Wallets**: In-app payment system
- **Lounges**: Waiting area bookings
- **Ratings & Reviews**: User feedback system
- **Notifications**: Multi-channel messaging
- **Promotions**: Discount and coupon system
- **Maintenance**: Vehicle maintenance tracking
- **Incidents**: Problem reporting and resolution
- **Audit Logs**: System activity tracking

## 🛠️ Development

### Hot Reload
The development environment includes Air for hot reload:
- Make changes to any `.go` file
- Application automatically rebuilds and restarts
- Check logs: `docker-compose -f docker-compose.dev.yml logs -f app`

### Database Access
```bash
# Connect to database
docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev

# View all tables
\dt

# Check a specific table
\d users
```

### Useful Commands

```bash
# View all logs
docker-compose -f docker-compose.dev.yml logs -f

# View app logs only
docker-compose -f docker-compose.dev.yml logs -f app

# View database logs
docker-compose -f docker-compose.dev.yml logs -f postgres

# Stop system
docker-compose -f docker-compose.dev.yml down

# Reset database (removes all data)
docker-compose -f docker-compose.dev.yml down --volumes
docker-compose -f docker-compose.dev.yml up --build

# Build for production
docker-compose up --build
```

## 📁 Project Structure

```
.
├── cmd/api/main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── database/             # Database connection
│   ├── handlers/             # HTTP handlers
│   └── models/               # Data models
├── migrations/               # Database migrations
│   ├── 001_initial.sql
│   ├── 002_init_db.sql
│   └── 003_smart_transit_complete.sql
├── docker/
│   ├── Dockerfile            # Production Docker image
│   └── Dockerfile.dev        # Development Docker image
├── docker-compose.yml        # Production compose
├── docker-compose.dev.yml    # Development compose
├── .air.toml                 # Hot reload configuration
├── setup.sh                 # Linux/macOS setup script
└── setup.bat                 # Windows setup script
```

## 🔧 Configuration

### Environment Variables
The system uses these environment variables (set in docker-compose.dev.yml):

- `DB_HOST=postgres`
- `DB_PORT=5432`
- `DB_USER=postgres`
- `DB_PASSWORD=password`
- `DB_NAME=myapp_dev`

### Air Configuration
Hot reload is configured in `.air.toml`:
- Watches all `.go` files
- Excludes `tmp/` and `vendor/` directories
- Polling enabled for Docker compatibility
- Builds to `./tmp/main`

## 🚨 Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   docker-compose -f docker-compose.dev.yml down
   docker-compose -f docker-compose.dev.yml up --build
   ```

2. **Database connection issues**
   - Wait for PostgreSQL to fully initialize (check logs)
   - Verify environment variables in docker-compose.dev.yml

3. **Hot reload not working**
   - Ensure `.air.toml` has `poll = true`
   - Check that source files are properly mounted in container

4. **Build failures**
   ```bash
   # Clean rebuild
   docker-compose -f docker-compose.dev.yml down --volumes
   docker system prune -f
   docker-compose -f docker-compose.dev.yml up --build
   ```

## 📝 Database Migrations

Migrations run automatically when the PostgreSQL container starts:

1. `001_initial.sql` - Basic user table
2. `002_init_db.sql` - Transit users with Firebase integration  
3. `003_smart_transit_complete.sql` - Complete system schema

**Note**: Migrations only run on first container start. To re-run migrations, use:
```bash
docker-compose -f docker-compose.dev.yml down --volumes
docker-compose -f docker-compose.dev.yml up --build
```

## 🎯 Next Steps

1. Implement API endpoints in `internal/handlers/`
2. Add data models in `internal/models/`
3. Configure Firebase authentication
4. Add API documentation
5. Implement frontend application
6. Add unit tests
7. Set up CI/CD pipeline

## 📄 License

[Add your license here]
