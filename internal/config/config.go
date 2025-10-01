package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Port       string

	// Auth / Asgardeo
	AsgardeoIssuer         string
	AsgardeoAudience       string
	JWKSCacheMinutes       int
	AsgardeoClientID       string
	AsgardeoClientSecret   string
	AsgardeoRedirectURI    string
	AsgardeoTokenEndpoint  string
	AsgardeoRevokeEndpoint string

	// SCIM / Admin APIs
	AsgardeoSCIMBaseURL      string
	AsgardeoSCIMToken        string
	AsgardeoSCIMClientID     string
	AsgardeoSCIMClientSecret string

	// Bootstrap user
	BootstrapAdminEmail    string
	BootstrapAdminName     string
	BootstrapAdminTempPass string
}

func Load() *Config {
	return &Config{
		DBHost:                   getEnv("DB_HOST", "localhost"),
		DBPort:                   getEnv("DB_PORT", "5432"),
		DBUser:                   getEnv("DB_USER", "postgres"),
		DBPassword:               getEnv("DB_PASSWORD", "password"),
		DBName:                   getEnv("DB_NAME", "myapp"),
		Port:                     getEnv("PORT", "8080"),
		AsgardeoIssuer:           getEnv("ASGARDEO_ISSUER", ""),
		AsgardeoAudience:         getEnv("ASGARDEO_AUDIENCE", ""),
		JWKSCacheMinutes:         getEnvInt("JWKS_CACHE_MINUTES", 60),
		AsgardeoClientID:         getEnv("ASGARDEO_CLIENT_ID", ""),
		AsgardeoClientSecret:     getEnv("ASGARDEO_CLIENT_SECRET", ""),
		AsgardeoRedirectURI:      getEnv("ASGARDEO_REDIRECT_URI", ""),
		AsgardeoTokenEndpoint:    getEnv("ASGARDEO_TOKEN_ENDPOINT", ""),
		AsgardeoRevokeEndpoint:   getEnv("ASGARDEO_REVOKE_ENDPOINT", ""),
		AsgardeoSCIMBaseURL:      getEnv("ASGARDEO_SCIM_BASE_URL", ""),
		AsgardeoSCIMToken:        getEnv("ASGARDEO_SCIM_TOKEN", ""),
		AsgardeoSCIMClientID:     getEnv("ASGARDEO_SCIM_CLIENT_ID", ""),
		AsgardeoSCIMClientSecret: getEnv("ASGARDEO_SCIM_CLIENT_SECRET", ""),
		BootstrapAdminEmail:      getEnv("BOOTSTRAP_ADMIN_EMAIL", ""),
		BootstrapAdminName:       getEnv("BOOTSTRAP_ADMIN_NAME", ""),
		BootstrapAdminTempPass:   getEnv("BOOTSTRAP_ADMIN_TEMP_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
