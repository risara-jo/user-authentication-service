package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// Ready returns a simple readiness payload.
func Ready(authConfigured bool, issuer string, authError string) gin.HandlerFunc {
    return func(c *gin.Context) {
        resp := gin.H{
            "auth_configured": authConfigured,
            "issuer":          issuer,
        }
        if !authConfigured && authError != "" {
            resp["error"] = authError
        }
        c.JSON(http.StatusOK, resp)
    }
}
