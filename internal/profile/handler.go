// Package profile – HTTP handler layer (transport).
package profile

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/response"
)

// Handler holds the profile service and processes HTTP requests.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler with the provided Service.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires the handler methods to the provided router group.
//
// Routes registered under /members/:id/profile:
//
//	GET  /members/:id/profile         – get member + extended profile
//	PUT  /members/:id/profile         – create / update profile (auth required)
//	GET  /members/:id/profile/summary – full dashboard summary
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	profile := rg.Group("/members/:id/profile")
	{
		profile.GET("", h.getProfile)
		profile.PUT("", middleware.AuthMiddleware(), h.upsertProfile)
		profile.GET("/summary", h.getSummary)
	}
}

// getProfile returns the member's basic info merged with their extended profile.
//
//	GET /api/v1/members/:id/profile
func (h *Handler) getProfile(c *gin.Context) {
	memberID := c.Param("id")

	prof, err := h.svc.GetProfile(memberID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, prof)
}

// upsertProfile creates or updates the extended profile for the authenticated member.
//
//	PUT /api/v1/members/:id/profile
//	Header: Authorization: Bearer <token>
func (h *Handler) upsertProfile(c *gin.Context) {
	memberID := c.Param("id")

	// Ownership check: the JWT user must be updating their own profile.
	callerID := c.GetString("user_id")
	if callerID == "" {
		response.Error(c, apperrors.New(http.StatusUnauthorized, "unauthorized"))
		return
	}
	if callerID != memberID {
		response.Error(c, apperrors.New(http.StatusForbidden, "you can only update your own profile"))
		return
	}

	var req UpsertProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	prof, err := h.svc.UpsertProfile(memberID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, prof)
}

// getSummary returns a full dashboard view for the member.
//
//	GET /api/v1/members/:id/profile/summary
func (h *Handler) getSummary(c *gin.Context) {
	memberID := c.Param("id")

	summary, err := h.svc.GetSummary(memberID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, summary)
}
