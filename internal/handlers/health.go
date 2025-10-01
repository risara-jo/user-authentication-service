package handlers

import (
	"github.com/gin-gonic/gin"

	"smart-transit-system/internal/httpx"
)

func HealthCheck(c *gin.Context) {
	httpx.RespondSuccess(c, 200, gin.H{"status": "healthy"}, "service is running")
}
