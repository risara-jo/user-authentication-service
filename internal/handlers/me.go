package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "smart-transit-system/internal/auth"
)

// Me returns authenticated user's token-derived profile.
func Me(c *gin.Context) {
    claims, ok := auth.FromContext(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "no auth context"})
        return
    }
    resp := gin.H{
        "sub":    claims.Subject(),
        "email":  claims.Email(),
        "scopes": claims.Scopes(),
        "roles":  claims.Roles(),
        "claims": claims, // include full map for now; can trim later
    }
    c.JSON(http.StatusOK, resp)
}

