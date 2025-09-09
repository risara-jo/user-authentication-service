package main

import (
    "log"
    "os"
    "strconv"

    "smart-transit-system/internal/auth"
    "smart-transit-system/internal/config"
    "smart-transit-system/internal/database"
    "smart-transit-system/internal/handlers"
    "smart-transit-system/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
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

	// Test database connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

    log.Println("Successfully connected to database")

    // Auto-migrate minimal user-related tables
    if err := db.AutoMigrate(&models.User{}, &models.Organization{}, &models.UserOrgMembership{}, &models.UserAudit{}); err != nil {
        log.Fatalf("Failed to auto-migrate database: %v", err)
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

        // Auth middleware group (protects endpoints below)
        if cfg.AsgardeoIssuer != "" {
            cacheMin, _ := strconv.Atoi(cfg.JWKSCacheMinutes)
            authenticator, err := auth.New(cfg.AsgardeoIssuer, cfg.AsgardeoAudience, cacheMin)
            if err != nil {
                log.Printf("WARN: auth middleware disabled (issuer setup failed): %v", err)
            } else {
                protected := api.Group("")
                protected.Use(authenticator.Middleware())
                // /me endpoint requires basic scope (openid or profile). We'll not enforce a scope hard here.
                protected.GET("/me", handlers.Me)
            }
        } else {
            log.Printf("WARN: ASGARDEO_ISSUER not set; auth endpoints disabled")
        }
    }

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("im again saying Server starting on port  %s", port)
	log.Printf("im again saying for 2nd time Server starting on port  %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
