// Package group – membership HTTP handler (transport) layer.
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
	svc  MembershipService
	repo MembershipRepository // used as GroupRoleProvider for role-check middleware
}

// NewMembershipHandler constructs a MembershipHandler.
func NewMembershipHandler(svc MembershipService, repo MembershipRepository) *MembershipHandler {
	return &MembershipHandler{svc: svc, repo: repo}
}

// RegisterRoutes wires membership endpoints onto the provided router group.
//
//	GET    /groups/:id/members                   – list all members in a group
//	POST   /groups/:id/members                   – add a member (Admin/SuperAdmin)
//	GET    /groups/:id/members/count             – count members
//	PATCH  /groups/:id/members/:memberId/role    – change a member's role (Admin/SuperAdmin)
//	GET    /groups/:id/my-role                   – caller's role in the group
//	GET    /groups-member-counts                 – member counts for all groups
func (h *MembershipHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/groups-member-counts", h.countAllMembers)
	rg.GET("/groups/:id/my-role", middleware.AuthMiddleware(), h.getMyRole)

	members := rg.Group("/groups/:id/members")
	{
		members.GET("", h.listMembers)
		members.POST("", middleware.AuthMiddleware(), h.addMember)
		members.GET("/count", h.countMembers)
		members.PATCH("/:memberId/role",
			middleware.AuthMiddleware(),
			middleware.RequireGroupRole(h.repo, "id", "ADMIN"),
			h.updateMemberRole,
		)
	}
}

// listMembers returns every membership record for the given group.
func (h *MembershipHandler) listMembers(c *gin.Context) {
	groupID := c.Param("id")
	memberships, err := h.svc.ListMembers(groupID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, memberships)
}

// addMember adds a new member to an existing group (Admin or SuperAdmin only).
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

	executingUserGlobalRole := c.GetString("global_role")
	membership, err := h.svc.AddMember(groupID, executingUserID, executingUserGlobalRole, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, membership)
}

// countMembers returns the number of members in a group.
func (h *MembershipHandler) countMembers(c *gin.Context) {
	groupID := c.Param("id")
	count, err := h.svc.CountMembers(groupID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"count": count})
}

// countAllMembers returns the number of members for every group.
func (h *MembershipHandler) countAllMembers(c *gin.Context) {
	counts, err := h.svc.CountAllMembers()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, counts)
}

// getMyRole returns the calling user's role in a group.
//
//	GET /api/v1/groups/:id/my-role
func (h *MembershipHandler) getMyRole(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.GetString("user_id")

	role, err := h.svc.GetMemberRole(groupID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"role": role, "group_id": groupID})
}

// updateMemberRole changes a member's role in a group (Admin/SuperAdmin only).
//
//	PATCH /api/v1/groups/:id/members/:memberId/role
//	Body: { "role": "ADMIN|TREASURER|MEMBER|VIEWER" }
func (h *MembershipHandler) updateMemberRole(c *gin.Context) {
	groupID := c.Param("id")
	memberID := c.Param("memberId")
	executingUserID := c.GetString("user_id")
	executingUserGlobalRole := c.GetString("global_role")

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	if err := h.svc.UpdateMemberRole(groupID, executingUserID, executingUserGlobalRole, memberID, req.Role); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "role updated", "group_id": groupID, "member_id": memberID, "role": req.Role})
}
