package dateproposal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/response"
)

// Handler processes HTTP requests for the date-proposal domain.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all date-proposal endpoints onto the provided router group.
//
//	POST  /date-proposals                              – create a proposal
//	GET   /date-proposals/:id                          – get proposal + votes
//	GET   /date-proposals/schedule/:scheduleID         – proposals for a schedule entry
//	POST  /date-proposals/:id/votes                    – cast a vote
//	PATCH /date-proposals/:id/accept                   – host accepts proposal
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := middleware.AuthMiddleware()

	dp := rg.Group("/date-proposals", auth)
	{
		dp.POST("", h.createProposal)
		dp.GET("/schedule/:scheduleID", h.listBySchedule)
		dp.GET("/:id", h.getProposal)
		dp.POST("/:id/votes", h.castVote)
		dp.PATCH("/:id/accept", h.acceptProposal)
	}
}

func (h *Handler) createProposal(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	var req CreateProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest,
			fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	proposal, err := h.svc.CreateProposal(callerID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, proposal)
}

func (h *Handler) getProposal(c *gin.Context) {
	id := c.Param("id")
	result, err := h.svc.GetProposal(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) listBySchedule(c *gin.Context) {
	scheduleID := c.Param("scheduleID")
	list, err := h.svc.ListProposalsBySchedule(scheduleID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) castVote(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	id := c.Param("id")

	var req CastVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest,
			fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	result, err := h.svc.CastVote(callerID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) acceptProposal(c *gin.Context) {
	callerID := h.callerID(c)
	if callerID == "" {
		return
	}

	id := c.Param("id")
	result, err := h.svc.AcceptProposal(callerID, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) callerID(c *gin.Context) string {
	id := c.GetString("user_id")
	if id == "" {
		response.Error(c, apperrors.New(http.StatusUnauthorized, "unauthorized"))
	}
	return id
}
