package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smart-transit-system/internal/auth"
	"smart-transit-system/internal/httpx"
)

// Me returns authenticated user's token-derived profile.
func Me(c *gin.Context) {
	claims, ok := auth.FromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "auth.no_context", "authentication required", nil)
		return
	}
	resp := gin.H{
		"sub":    claims.Subject(),
		"email":  claims.Email(),
		"scopes": claims.Scopes(),
		"roles":  claims.Roles(),
		"claims": claims,
	}
	httpx.RespondSuccess(c, http.StatusOK, resp, "profile fetched")
}
