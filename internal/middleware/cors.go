package middleware

import (
    "net/http"
    "os"
    "strings"

    "github.com/gin-gonic/gin"
)

// CORS returns a simple CORS middleware.
func CORS() gin.HandlerFunc {
    allow := os.Getenv("CORS_ALLOW_ORIGINS")
    var origins []string
    if allow == "" {
        origins = []string{"*"}
    } else {
        for _, p := range strings.Split(allow, ",") {
            origins = append(origins, strings.TrimSpace(p))
        }
    }

    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        allowOrigin := "*"
        if len(origins) > 0 && origins[0] != "*" && origin != "" {
            for _, o := range origins {
                if o == origin {
                    allowOrigin = origin
                    break
                }
            }
            if allowOrigin != origin {
                // not allowed, but still send common headers
                allowOrigin = origins[0]
            }
        }

        c.Header("Access-Control-Allow-Origin", allowOrigin)
        c.Header("Vary", "Origin")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

        if c.Request.Method == http.MethodOptions {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        c.Next()
    }
}

