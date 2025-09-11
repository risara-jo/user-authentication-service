package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// AuthNotConfigured responds when OIDC is not correctly configured at startup.
func AuthNotConfigured(c *gin.Context) {
    c.JSON(http.StatusServiceUnavailable, gin.H{
        "error": "auth_not_configured",
        "message": "Asgardeo OIDC is not configured or discovery failed. Ensure ASGARDEO_ISSUER is set and reachable.",
    })
}

