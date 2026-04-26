// Package payout – HTTP handler layer (transport).
package payout

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/response"
)

// Handler processes HTTP requests for the payout domain.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all payout endpoints onto the provided router group.
// All routes require a valid JWT (middleware.AuthMiddleware).
//
//	POST  /api/v1/payouts/schedule              – schedule a payout for a cycle
//	POST  /api/v1/payouts/disburse              – record the actual money transfer
//	GET   /api/v1/payouts/groups/:groupID        – list all payouts for a group
//	GET   /api/v1/payouts/recipients/:memberID   – list all payouts a member receives
//	GET   /api/v1/payouts/:id                    – get single payout + disbursement
//	PATCH /api/v1/payouts/:id/cancel             – cancel a scheduled payout
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	payouts := rg.Group("/payouts", middleware.AuthMiddleware())
	{
		payouts.POST("/schedule", h.schedulePayout)
		payouts.POST("/disburse", h.recordDisbursement)
		payouts.GET("/groups/:groupID", h.getGroupPayouts)
		payouts.GET("/recipients/:memberID", h.getRecipientPayouts)
		payouts.GET("/:id", h.getPayout)
		payouts.PATCH("/:id/cancel", h.cancelPayout)
	}
}

// schedulePayout creates a payout schedule for one cycle of a group.
//
//	POST /api/v1/payouts/schedule
func (h *Handler) schedulePayout(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	schedule, err := h.svc.SchedulePayout(callerID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, schedule)
}

// recordDisbursement records the actual money transfer for a scheduled payout.
//
//	POST /api/v1/payouts/disburse
func (h *Handler) recordDisbursement(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	var req RecordDisbursementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	disbursement, err := h.svc.RecordDisbursement(callerID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, disbursement)
}

// getGroupPayouts lists all payout schedules for a group.
//
//	GET /api/v1/payouts/groups/:groupID
func (h *Handler) getGroupPayouts(c *gin.Context) {
	groupID := c.Param("groupID")

	summaries, err := h.svc.GetGroupPayouts(groupID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, summaries)
}

// getRecipientPayouts lists all payouts a member is set to receive.
//
//	GET /api/v1/payouts/recipients/:memberID
func (h *Handler) getRecipientPayouts(c *gin.Context) {
	memberID := c.Param("memberID")

	summaries, err := h.svc.GetRecipientPayouts(memberID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, summaries)
}

// getPayout returns a single payout schedule enriched with its disbursement.
//
//	GET /api/v1/payouts/:id
func (h *Handler) getPayout(c *gin.Context) {
	id := c.Param("id")

	summary, err := h.svc.GetPayout(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, summary)
}

// cancelPayout cancels a payout that has not yet been disbursed.
//
//	PATCH /api/v1/payouts/:id/cancel
func (h *Handler) cancelPayout(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	id := c.Param("id")

	var req CancelScheduleRequest
	// Body is optional — ignore binding error for an empty body.
	_ = c.ShouldBindJSON(&req)

	if err := h.svc.CancelPayout(callerID, id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// callerID extracts the authenticated user's ID from the Gin context.
// It writes a 401 response and returns "" when the claim is missing.
func (h *Handler) callerID(c *gin.Context) string {
	id := c.GetString("user_id")
	if id == "" {
		response.Error(c, apperrors.New(http.StatusUnauthorized, "unauthorized"))
	}
	return id
}
