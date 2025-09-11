package main

import (
    "log"
    "os"
    "strconv"
    "strings"

    "smart-transit-system/internal/auth"
    "smart-transit-system/internal/config"
    "smart-transit-system/internal/database"
    "smart-transit-system/internal/handlers"
    mid "smart-transit-system/internal/middleware"
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

    // Initialize database (non-fatal so health endpoint still works)
    db, err := database.NewPostgresDB(cfg)
    if err != nil {
        log.Printf("WARN: DB connect failed: %v", err)
    } else {
        sqlDB, err := db.DB()
        if err != nil {
            log.Printf("WARN: DB instance error: %v", err)
        } else if err := sqlDB.Ping(); err != nil {
            log.Printf("WARN: DB ping failed: %v", err)
        } else {
            log.Println("Successfully connected to database")
            if err := db.AutoMigrate(&models.User{}, &models.Organization{}, &models.UserOrgMembership{}, &models.UserAudit{}); err != nil {
                log.Printf("WARN: Auto-migrate failed: %v", err)
            }
        }
    }

    // Setup Gin router
    r := gin.Default()
    // CORS for SPA calls
    r.Use(mid.CORS())

    // Health check endpoint
    r.GET("/health", handlers.HealthCheck)
    r.GET("/health2", handlers.HealthCheck)

    // API routes
    api := r.Group("/api/v1")
    {
        api.GET("/ping", func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "pong"})
        })

        // Auth helper for SPA: returns authorize URL template
        api.GET("/auth/login", handlers.AuthLogin)
        // Optional: redirect helper when given PKCE params
        api.GET("/auth/authorize", handlers.AuthAuthorize)

        // Auth middleware group (protects /me) or fallback if misconfigured
        authReady := false
        var authErrMsg string
        if cfg.AsgardeoIssuer != "" {
            cacheMin, _ := strconv.Atoi(cfg.JWKSCacheMinutes)
            authenticator, err := auth.New(cfg.AsgardeoIssuer, cfg.AsgardeoAudience, cacheMin)
            if err != nil {
                log.Printf("WARN: auth setup failed; /me will return 503: %v", err)
                api.GET("/me", handlers.AuthNotConfigured)
                authErrMsg = err.Error()
            } else {
                protected := api.Group("")
                protected.Use(authenticator.Middleware())
                protected.GET("/me", handlers.Me)
                authReady = true
            }
        } else {
            log.Printf("WARN: ASGARDEO_ISSUER not set; /me will return 503")
            api.GET("/me", handlers.AuthNotConfigured)
            authErrMsg = "ASGARDEO_ISSUER not set"
        }

        // Readiness endpoint shows auth configuration detected at startup
        api.GET("/ready", handlers.Ready(authReady, cfg.AsgardeoIssuer, authErrMsg))
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

// retain strings import usage
var _ = strings.TrimSpace
