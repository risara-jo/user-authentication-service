package handlers

import (
	"github.com/gin-gonic/gin"
	"smart-transit-system/internal/httpx"
)

// AuthNotConfigured responds when OIDC is not correctly configured at startup.
func AuthNotConfigured(c *gin.Context) {
	httpx.RespondError(c, 503, "auth_not_configured", "Asgardeo OIDC is not configured or discovery failed. Ensure ASGARDEO_ISSUER is set and reachable.", nil)
}
