// Package group – membership HTTP handler (transport) layer.
//
// MembershipHandler manages the /groups/:id/members sub-resource.
// It is kept separate from Handler to honour the Single Responsibility Principle
// while still living in the same domain package.
package group

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/response"
)

// MembershipHandler handles HTTP requests for group membership operations.
type MembershipHandler struct {
	svc MembershipService
}

// NewMembershipHandler constructs a MembershipHandler.
func NewMembershipHandler(svc MembershipService) *MembershipHandler {
	return &MembershipHandler{svc: svc}
}

// RegisterRoutes wires membership endpoints onto the provided router group.
//
// Routes:
//
//	GET  /groups/:id/members  – list all members in a group
//	POST /groups/:id/members  – add a member to a group
func (h *MembershipHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/groups-member-counts", h.countAllMembers)

	members := rg.Group("/groups/:id/members")
	{
		members.GET("", h.listMembers)
		members.POST("", middleware.AuthMiddleware(), h.addMember)
		members.GET("/count", h.countMembers)
	}
}

// listMembers returns every membership record for the given group.
//
//	GET /api/v1/groups/:id/members
func (h *MembershipHandler) listMembers(c *gin.Context) {
	groupID := c.Param("id")

	memberships, err := h.svc.ListMembers(groupID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, memberships)
}

// addMember adds a new member to an existing group.
//
//	POST /api/v1/groups/:id/members
//	Body: { "member_id": "<uuid>" }
func (h *MembershipHandler) addMember(c *gin.Context) {
	groupID := c.Param("id")
	
	executingUserID := c.GetString("user_id")
	if executingUserID == "" {
		response.Error(c, apperrors.New(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	membership, err := h.svc.AddMember(groupID, executingUserID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, membership)
}

// countMembers returns the number of members in a group.
//
//	GET /api/v1/groups/:id/members/count
func (h *MembershipHandler) countMembers(c *gin.Context) {
	groupID := c.Param("id")

	count, err := h.svc.CountMembers(groupID)
	if err != nil {
		response.Error(c, err)
		return
	}
	
	// Returning the count wrapped in an object so it can be easily parsed by frontend.
	response.OK(c, gin.H{"count": count})
}

// countAllMembers returns the number of members for every group.
//
//	GET /api/v1/groups-member-counts
func (h *MembershipHandler) countAllMembers(c *gin.Context) {
	counts, err := h.svc.CountAllMembers()
	if err != nil {
		response.Error(c, err)
		return
	}
	
	response.OK(c, counts)
}
