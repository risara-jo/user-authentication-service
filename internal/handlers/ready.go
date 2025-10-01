package handlers

import (
	"github.com/gin-gonic/gin"

	"smart-transit-system/internal/httpx"
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
		httpx.RespondSuccess(c, 200, resp, "ready")
	}
}
