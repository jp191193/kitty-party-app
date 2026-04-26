package scheduleinvitation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/response"
)

// Handler processes HTTP requests for the schedule-invitation domain.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all invitation endpoints onto the provided router group.
//
//	POST  /kitty-cycles/schedule-entries/:scheduleID/invitations  – send invitations
//	GET   /kitty-cycles/schedule-entries/:scheduleID/invitations  – list invitations
//	GET   /invitations/my                                         – my invitations
//	PATCH /invitations/:id/accept                                 – accept date
//	PATCH /invitations/:id/propose-date                          – mark date-proposed
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := middleware.AuthMiddleware()

	// Schedule-entry scoped endpoints (nested under kitty-cycles prefix)
	entries := rg.Group("/kitty-cycles/schedule-entries", auth)
	{
		entries.POST("/:scheduleID/invitations", h.sendInvitations)
		entries.GET("/:scheduleID/invitations", h.listBySchedule)
	}

	// Member-scoped and per-invitation endpoints
	inv := rg.Group("/invitations", auth)
	{
		inv.GET("/my", h.myInvitations)
		inv.PATCH("/:id/accept", h.acceptDate)
		inv.PATCH("/:id/propose-date", h.markDateProposed)
	}
}

func (h *Handler) sendInvitations(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}
	scheduleID := c.Param("scheduleID")
	result, err := h.svc.SendInvitations(callerID, scheduleID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

func (h *Handler) listBySchedule(c *gin.Context) {
	scheduleID := c.Param("scheduleID")
	list, err := h.svc.GetInvitationsBySchedule(scheduleID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) myInvitations(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}
	list, err := h.svc.GetMyInvitations(callerID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) acceptDate(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}
	id := c.Param("id")
	inv, err := h.svc.AcceptDate(callerID, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inv)
}

func (h *Handler) markDateProposed(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}
	id := c.Param("id")
	inv, err := h.svc.MarkDateProposed(callerID, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inv)
}

func (h *Handler) callerID(c *gin.Context) string {
	id := c.GetString("user_id")
	if id == "" {
		response.Error(c, apperrors.New(http.StatusUnauthorized, "unauthorized"))
	}
	return id
}

