package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"smart-transit-system/internal/auth"
	"smart-transit-system/internal/httpx"
	"smart-transit-system/internal/models"
	"smart-transit-system/internal/services/admin"
)

type AdminHandler struct {
	service *admin.Service
}

func NewAdminHandler(service *admin.Service) *AdminHandler {
	return &AdminHandler{service: service}
}

// GetRoles godoc
// @Summary List available roles
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200
// @Router /roles [get]
func (h *AdminHandler) GetRoles(c *gin.Context) {
	httpx.RespondSuccess(c, http.StatusOK, gin.H{"roles": h.service.ListRoles()}, "roles fetched")
}

// GetProfiles godoc
// @Summary List user profiles
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param q query string false "search term"
// @Param role query string false "role filter"
// @Param company_id query string false "company id filter"
// @Param lounge_id query string false "lounge id filter"
// @Param status query string false "status filter"
// @Param page query int false "page number"
// @Param limit query int false "page size"
// @Success 200
// @Router /profiles [get]
func (h *AdminHandler) GetProfiles(c *gin.Context) {
	actor, ok := auth.ValuesFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "auth.no_context", "authentication required", nil)
		return
	}

	page := parseQueryInt(c, "page", 1)
	limit := parseQueryInt(c, "limit", 20)

	result, err := h.service.ListProfiles(c.Request.Context(), actor, admin.ListProfilesInput{
		Query:     c.Query("q"),
		Role:      c.Query("role"),
		CompanyID: c.Query("company_id"),
		LoungeID:  c.Query("lounge_id"),
		Status:    c.Query("status"),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		httpx.RespondAppError(c, err)
		return
	}

	items := make([]gin.H, 0, len(result.Profiles))
	for _, p := range result.Profiles {
		items = append(items, profileResponse(p))
	}

	data := gin.H{
		"items": items,
		"pagination": gin.H{
			"total": result.Total,
			"page":  result.Page,
			"limit": result.Limit,
		},
	}
	httpx.RespondSuccess(c, http.StatusOK, data, "profiles fetched")
}

type createUserRequest struct {
	Email             string   `json:"email" binding:"required,email"`
	Name              string   `json:"name" binding:"required"`
	Roles             []string `json:"roles"`
	CompanyID         *string  `json:"company_id"`
	LoungeID          *string  `json:"lounge_id"`
	TemporaryPassword *string  `json:"temporary_password"`
}

// CreateUser godoc
// @Summary Create admin user
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body createUserRequest true "create user payload"
// @Success 201
// @Router /admin/users [post]
func (h *AdminHandler) CreateUser(c *gin.Context) {
	actor, ok := auth.ValuesFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "auth.no_context", "authentication required", nil)
		return
	}
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "validation_failed", "invalid request payload", err.Error())
		return
	}

	profile, err := h.service.CreateUser(c.Request.Context(), actor, admin.CreateUserInput{
		Email:             strings.TrimSpace(req.Email),
		Name:              strings.TrimSpace(req.Name),
		Roles:             req.Roles,
		CompanyID:         req.CompanyID,
		LoungeID:          req.LoungeID,
		TemporaryPassword: req.TemporaryPassword,
	})
	if err != nil {
		httpx.RespondAppError(c, err)
		return
	}

	httpx.RespondSuccess(c, http.StatusCreated, profileResponse(*profile), "user created")
}

type updateUserRequest struct {
	Name        *string `json:"name"`
	Email       *string `json:"email"`
	CompanyID   *string `json:"company_id"`
	LoungeID    *string `json:"lounge_id"`
	PrimaryRole *string `json:"primary_role"`
}

// UpdateUser godoc
// @Summary Update user profile
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param body body updateUserRequest true "update payload"
// @Success 200
// @Router /admin/users/{id} [patch]
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	actor, ok := auth.ValuesFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "auth.no_context", "authentication required", nil)
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "validation_failed", "invalid request payload", err.Error())
		return
	}

	profile, err := h.service.UpdateUser(c.Request.Context(), actor, c.Param("id"), admin.UpdateUserInput{
		Name:        trimPointer(req.Name),
		Email:       trimPointer(req.Email),
		CompanyID:   trimPointer(req.CompanyID),
		LoungeID:    trimPointer(req.LoungeID),
		PrimaryRole: trimPointer(req.PrimaryRole),
	})
	if err != nil {
		httpx.RespondAppError(c, err)
		return
	}

	httpx.RespondSuccess(c, http.StatusOK, profileResponse(*profile), "user updated")
}

type replaceRolesRequest struct {
	Roles []string `json:"roles" binding:"required"`
}

// ReplaceRoles godoc
// @Summary Replace user roles
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param body body replaceRolesRequest true "roles payload"
// @Success 200
// @Router /admin/users/{id}/roles [put]
func (h *AdminHandler) ReplaceRoles(c *gin.Context) {
	actor, ok := auth.ValuesFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "auth.no_context", "authentication required", nil)
		return
	}
	var req replaceRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "validation_failed", "invalid request payload", err.Error())
		return
	}

	profile, err := h.service.ReplaceRoles(c.Request.Context(), actor, c.Param("id"), req.Roles)
	if err != nil {
		httpx.RespondAppError(c, err)
		return
	}

	httpx.RespondSuccess(c, http.StatusOK, profileResponse(*profile), "roles updated")
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus godoc
// @Summary Update user status
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param body body updateStatusRequest true "status payload"
// @Success 200
// @Router /admin/users/{id}/status [post]
func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	actor, ok := auth.ValuesFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "auth.no_context", "authentication required", nil)
		return
	}
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "validation_failed", "invalid request payload", err.Error())
		return
	}

	profile, err := h.service.UpdateStatus(c.Request.Context(), actor, c.Param("id"), req.Status)
	if err != nil {
		httpx.RespondAppError(c, err)
		return
	}

	httpx.RespondSuccess(c, http.StatusOK, profileResponse(*profile), "status updated")
}

func parseQueryInt(c *gin.Context, key string, fallback int) int {
	raw := c.Query(key)
	if raw == "" {
		return fallback
	}
	if v, err := strconv.Atoi(raw); err == nil {
		return v
	}
	return fallback
}

func trimPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func profileResponse(p models.UserProfile) gin.H {
	resp := gin.H{
		"id":         p.ID,
		"sub":        p.Sub,
		"email":      p.Email,
		"name":       p.Name,
		"roles":      []string(p.Roles),
		"status":     p.Status,
		"created_at": p.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at": p.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if p.CompanyID != nil {
		resp["company_id"] = *p.CompanyID
	} else {
		resp["company_id"] = nil
	}
	if p.LoungeID != nil {
		resp["lounge_id"] = *p.LoungeID
	} else {
		resp["lounge_id"] = nil
	}
	return resp
}
