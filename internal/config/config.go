package config

import "os"

type Config struct {
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    Port       string
    // Auth / Asgardeo
    AsgardeoIssuer   string // e.g., https://<org>.asgard.<region>.asgardeo.io/t/<tenant>/oauth2/token (issuer base)
    AsgardeoAudience string // optional expected audience; leave empty to skip aud check
    JWKSCacheMinutes string // optional, minutes to cache JWKS before refresh
}

func Load() *Config {
    return &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", "password"),
        DBName:     getEnv("DB_NAME", "myapp"),
        Port:       getEnv("PORT", "8080"),
        AsgardeoIssuer:   getEnv("ASGARDEO_ISSUER", ""),
        AsgardeoAudience: getEnv("ASGARDEO_AUDIENCE", ""),
        JWKSCacheMinutes: getEnv("JWKS_CACHE_MINUTES", "60"),
    }
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
