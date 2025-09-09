#!/bin/bash
# Smart Transit System Setup Script
# This script sets up the complete Smart Transit System development environment

echo ""
echo "            ______"
echo "            _\\ _~-\\___"
echo "    =  = ==(____AA____D"
echo "                \\_____\\___________________,-~~~~~~~\`-.._"
echo "                /     o O o o o o O O o o o o o o O o  |\\_" 
echo "                \`~-.__        ___..----..                  )"
echo "                      \`---~~\\___________/------------\`\`\`\`"
echo "                      =  ===(________D"
echo ""
echo "      AAA        AAA       SSSSS      LLL     "
echo "     A   A      A   A     S          LLL     "
echo "    AAAAAAA    AAAAAAA     SSSS      LLL     "
echo "   A       A  A       A        S     LLL     "
echo "  A         AA         A  SSSSS      LLLLLLL"
echo ""
echo "                    dev env"
echo ""
echo "🚌 Smart Transit System Setup Starting..."
echo "========================================="

# Check if Docker and Docker Compose are installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

echo "✅ Docker and Docker Compose are installed"

# Stop any existing containers
echo "🛑 Stopping any existing containers..."
docker-compose -f docker-compose.dev.yml down --volumes --remove-orphans

# Build and start the development environment
echo "🔨 Building and starting the Smart Transit System..."
docker-compose -f docker-compose.dev.yml up --build -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are running
if docker-compose -f docker-compose.dev.yml ps | grep -q "Up"; then
    echo ""
    echo ""
    echo "            ______"
    echo "            _\\ _~-\\___"
    echo "    =  = ==(____AA____D"
    echo "                \\_____\\___________________,-~~~~~~~\`-.._"
    echo "                /     o O o o o o O O o o o o o o O o  |\\_" 
    echo "                \`~-.__        ___..----..                  )"
    echo "                      \`---~~\\___________/------------\`\`\`\`"
    echo "                      =  ===(________D"
    echo ""
    echo "      AAA        AAA       SSSSS      LLL     "
    echo "     A   A      A   A     S          LLL     "
    echo "    AAAAAAA    AAAAAAA     SSSS      LLL     "
    echo "   A       A  A       A        S     LLL     "
    echo "  A         AA         A  SSSSS      LLLLLLL"
    echo ""
    echo "                    dev env"
    echo ""
    echo "🎉 Smart Transit System is now running!"
    echo "========================================="
    echo "🌐 Application: http://localhost:8080"
    echo "🗄️  PostgreSQL: localhost:5432"
    echo "   - Database: myapp_dev"
    echo "   - Username: postgres"
    echo "   - Password: password"
    echo ""
    echo "📋 Useful Commands:"
    echo "   - View logs: docker-compose -f docker-compose.dev.yml logs -f"
    echo "   - Stop system: docker-compose -f docker-compose.dev.yml down"
    echo "   - Database shell: docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev"
    echo ""
    echo "🔥 Hot reload is enabled - make changes to Go files and they'll auto-restart!"
else
    echo "❌ Something went wrong. Check the logs with:"
    echo "   docker-compose -f docker-compose.dev.yml logs"
fi
