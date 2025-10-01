package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smart-transit-system/internal/apperror"
)

// RespondSuccess standardizes successful API responses.
func RespondSuccess(c *gin.Context, status int, data any, message string) {
	if message == "" {
		message = "ok"
	}
	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// RespondError standardizes error responses to match the API contract.
func RespondError(c *gin.Context, status int, code string, message string, details any) {
	if message == "" {
		message = code
	}
	c.JSON(status, gin.H{
		"success": false,
		"code":    code,
		"message": message,
		"details": details,
	})
}

func RespondAppError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if appErr, ok := err.(*apperror.Error); ok {
		status := appErr.Status
		if status == 0 {
			status = http.StatusInternalServerError
		}
		RespondError(c, status, appErr.Code, appErr.Message, appErr.Details)
		return
	}
	RespondError(c, http.StatusInternalServerError, "internal_error", "unexpected error", err.Error())
}
