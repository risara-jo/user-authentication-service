package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"smart-transit-system/internal/audit"
	"smart-transit-system/internal/auth"
	"smart-transit-system/internal/config"
	"smart-transit-system/internal/database"
	"smart-transit-system/internal/handlers"
	mid "smart-transit-system/internal/middleware"
	"smart-transit-system/internal/models"
	"smart-transit-system/internal/repository"
	"smart-transit-system/internal/scim"
	adminsvc "smart-transit-system/internal/services/admin"

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
			if err := db.AutoMigrate(&models.UserProfile{}, &models.AuditLog{}); err != nil {
				log.Printf("WARN: Auto-migrate failed: %v", err)
			}
		}
	}

	var userRepo repository.ProfileRepository
	var auditRepo repository.AuditRepository
	var auditLogger *audit.Logger
	if db != nil {
		userRepo = repository.NewUserProfileRepository(db)
		auditRepo = repository.NewAuditRepository(db)
		auditLogger = audit.NewLogger(auditRepo)
	}

	var scimClient *scim.Client
	if cfg.AsgardeoSCIMBaseURL != "" {
		client, err := scim.New(scim.Config{
			BaseURL:      cfg.AsgardeoSCIMBaseURL,
			StaticToken:  cfg.AsgardeoSCIMToken,
			ClientID:     cfg.AsgardeoSCIMClientID,
			ClientSecret: cfg.AsgardeoSCIMClientSecret,
		})
		if err != nil {
			log.Printf("WARN: SCIM client init failed: %v", err)
		} else {
			scimClient = client
		}
	}

	var adminService *adminsvc.Service
	if userRepo != nil {
		adminService = adminsvc.NewService(userRepo, scimClient, auditLogger)
	}
	var adminHandler *handlers.AdminHandler
	if adminService != nil {
		adminHandler = handlers.NewAdminHandler(adminService)
	}

	if adminService != nil && cfg.BootstrapAdminEmail != "" {
		ensureBootstrapAdmin(context.Background(), adminService, userRepo, scimClient, cfg)
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
			authenticator, err := auth.New(cfg.AsgardeoIssuer, cfg.AsgardeoAudience, cfg.JWKSCacheMinutes)
			if err != nil {
				log.Printf("WARN: auth setup failed; /me will return 503: %v", err)
				api.GET("/me", handlers.AuthNotConfigured)
				authErrMsg = err.Error()
			} else {
				protected := api.Group("")
				protected.Use(authenticator.Middleware())
				protected.GET("/me", handlers.Me)

				if adminHandler != nil {
					adminRead := protected.Group("")
					adminRead.Use(auth.ScopeGuardAny("users.manage", "org.manage"))
					adminRead.GET("/roles", mid.AdminReadRateLimiter(), adminHandler.GetRoles)
					adminRead.GET("/profiles", mid.AdminReadRateLimiter(), adminHandler.GetProfiles)

					adminWrite := protected.Group("")
					adminWrite.Use(auth.ScopeGuard("users.manage"))
					adminWrite.POST("/admin/users", mid.AdminWriteRateLimiter(), adminHandler.CreateUser)
					adminWrite.PATCH("/admin/users/:id", mid.AdminWriteRateLimiter(), adminHandler.UpdateUser)
					adminWrite.PUT("/admin/users/:id/roles", mid.AdminWriteRateLimiter(), adminHandler.ReplaceRoles)
					adminWrite.POST("/admin/users/:id/status", mid.AdminWriteRateLimiter(), adminHandler.UpdateStatus)
				}

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

func ensureBootstrapAdmin(ctx context.Context, service *adminsvc.Service, repo repository.ProfileRepository, scimClient *scim.Client, cfg *config.Config) {
	if service == nil || repo == nil || scimClient == nil {
		return
	}
	email := strings.TrimSpace(cfg.BootstrapAdminEmail)
	if email == "" {
		return
	}
	if _, err := repo.GetByEmail(ctx, email); err == nil {
		return
	} else if err != nil && !errors.Is(err, repository.ErrProfileNotFound) {
		log.Printf("WARN: bootstrap admin lookup failed: %v", err)
		return
	}
	name := strings.TrimSpace(cfg.BootstrapAdminName)
	if name == "" {
		name = "Super Admin"
	}
	roles := []string{"admin"}
	actor := auth.ContextValues{
		Subject: "bootstrap",
		Email:   "bootstrap@system",
		Roles:   []string{"system"},
		Scopes:  []string{"users.manage"},
	}
	var tempPass *string
	if cfg.BootstrapAdminTempPass != "" {
		temp := cfg.BootstrapAdminTempPass
		tempPass = &temp
	}
	if _, err := service.CreateUser(ctx, actor, adminsvc.CreateUserInput{
		Email:             email,
		Name:              name,
		Roles:             roles,
		TemporaryPassword: tempPass,
	}); err != nil {
		log.Printf("WARN: bootstrap admin creation failed: %v", err)
	} else {
		log.Printf("Bootstrap admin ensured for %s", email)
	}
}
