@echo off
REM Smart Transit System Setup Script for Windows
REM This script sets up the complete Smart Transit System development environment

REM Set UTF-8 encoding for proper Unicode display
chcp 65001 >nul

echo.
echo            ______
echo            _\ _~-\___
echo    =  = ==(____AA____D
echo                \_____\___________________,-~~~~~~~`-.._
echo                /     o O o o o o O O o o o o o o O o  ^|\_ 
echo                `~-.__        ___..----..                  ^)
echo                      `---~~\___________/------------`````
echo                      =  ===^(_________D
echo.
echo     █████╗      █████╗      ███████╗    ██╗     
echo    ██╔══██╗    ██╔══██╗    ██╔════╝    ██║     
echo    ███████║    ███████║    ███████╗    ██║     
echo    ██╔══██║    ██╔══██║    ╚════██║    ██║     
echo    ██║  ██║    ██║  ██║    ███████║    ███████╗
echo    ╚═╝  ╚═╝    ╚═╝  ╚═╝    ╚══════╝    ╚══════╝
echo                         dev env
echo.
echo 🚌 Smart Transit System Setup Starting...
echo =========================================

REM Check if Docker is installed
docker --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker is not installed. Please install Docker Desktop first.
    pause
    exit /b 1
)

REM Check if Docker Compose is installed
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker Compose is not installed. Please install Docker Desktop with Compose.
    pause
    exit /b 1
)

echo ✅ Docker and Docker Compose are installed

REM Stop any existing containers
echo 🛑 Stopping any existing containers...
docker-compose -f docker-compose.dev.yml down --volumes --remove-orphans

REM Build and start the development environment
echo 🔨 Building and starting the Smart Transit System...
docker-compose -f docker-compose.dev.yml up --build -d

REM Wait for services to be ready
echo ⏳ Waiting for services to be ready...
timeout /t 10 /nobreak > nul

echo.
echo.
echo                                      ______
echo                                      _\ _~-\___
echo                              =  = ==(____AA____D
echo                                          \_____\___________________,-~~~~~~~`-.._
echo                                          /     o O o o o o O O o o o o o o O o  ^|\_ 
echo                                          `~-.__        ___..----..                  ^)
echo                                                `---~~\___________/------------`````
echo                                                =  ===^(_________D
echo.
echo.
echo     █████╗      █████╗      ███████╗    ██╗     
echo    ██╔══██╗    ██╔══██╗    ██╔════╝    ██║     
echo    ███████║    ███████║    ███████╗    ██║     
echo    ██╔══██║    ██╔══██║    ╚════██║    ██║     
echo    ██║  ██║    ██║  ██║    ███████║    ███████╗
echo    ╚═╝  ╚═╝    ╚═╝  ╚═╝    ╚══════╝    ╚══════╝    dev env
echo                         
echo.
echo 🎉 Smart Transit System is now running!
echo =========================================
echo 🌐 Application: http://localhost:8080
echo 🗄️  PostgreSQL: localhost:5432
echo    - Database: myapp_dev
echo    - Username: postgres
echo    - Password: password
echo.
echo 📋 Useful Commands:
echo    - View logs: docker-compose -f docker-compose.dev.yml logs -f
echo    - Stop system: docker-compose -f docker-compose.dev.yml down
echo    - Database shell: docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d myapp_dev
echo.
echo 🔥 Hot reload is enabled - make changes to Go files and they'll auto-restart!
echo.
pause
