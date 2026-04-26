// Package kittycycle – HTTP handler (transport) layer.
package kittycycle

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/response"
)

// Handler processes HTTP requests for the kitty cycle domain.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all kitty-cycle endpoints onto the provided router group.
// All routes require a valid JWT (middleware.AuthMiddleware).
//
//	POST  /api/v1/kitty-cycles/schedule           – schedule a new kitty cycle
//	GET   /api/v1/kitty-cycles/:id                – get cycle + schedule
//	GET   /api/v1/kitty-cycles/groups/:groupID    – list cycles for a group
//	PATCH /api/v1/kitty-cycles/:id/start          – start a PENDING cycle
//	PATCH /api/v1/kitty-cycles/:id/cancel         – cancel a cycle
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	cycles := rg.Group("/kitty-cycles", middleware.AuthMiddleware())
	{
		cycles.POST("/schedule", h.scheduleCycle)
		cycles.GET("/groups/:groupID", h.getCyclesByGroup)
		cycles.GET("/:id", h.getCycle)
		cycles.PATCH("/:id/start", h.startCycle)
		cycles.PATCH("/:id/cancel", h.cancelCycle)
		cycles.POST("/schedule-entries/:scheduleID/roll-host", h.rollHost)
	}
}

func (h *Handler) scheduleCycle(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	var req ScheduleCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	cycle, err := h.svc.ScheduleCycle(callerID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, cycle)
}

func (h *Handler) getCycle(c *gin.Context) {
	id := c.Param("id")
	cycle, err := h.svc.GetCycle(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, cycle)
}

func (h *Handler) getCyclesByGroup(c *gin.Context) {
	groupID := c.Param("groupID")
	cycles, err := h.svc.GetCyclesByGroup(groupID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, cycles)
}

func (h *Handler) startCycle(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	id := c.Param("id")
	result, err := h.svc.StartCycle(callerID, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) rollHost(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	scheduleID := c.Param("scheduleID")
	result, err := h.svc.RollHost(callerID, scheduleID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) cancelCycle(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	id := c.Param("id")

	var req CancelCycleRequest
	// Body is optional — ignore binding error for an empty body.
	_ = c.ShouldBindJSON(&req)

	if err := h.svc.CancelCycle(callerID, id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// callerID extracts the authenticated user ID or writes a 401.
func (h *Handler) callerID(c *gin.Context) string {
	id := c.GetString("user_id")
	if id == "" {
		response.Error(c, apperrors.New(http.StatusUnauthorized, "unauthorized"))
	}
	return id
}
