// Package member – HTTP handler layer (transport).
package member

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/response"
)

// Handler holds the member service and handles HTTP requests.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler with the provided Service.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires the handler methods to the provided router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	members := rg.Group("/members")
	{
		members.GET("", h.listMembers)
		members.GET("/:id", h.getMember)
		members.POST("", h.createMember)
		members.PUT("/:id", h.updateMember)
		members.DELETE("/:id", h.deleteMember)
	}
}

// listMembers returns every member in the system.
func (h *Handler) listMembers(c *gin.Context) {
	members, err := h.svc.GetAll()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, members)
}

// getMember returns a single member by UUID.
func (h *Handler) getMember(c *gin.Context) {
	id := c.Param("id")
	m, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, m)
}

// createMember adds a new member to the group.
func (h *Handler) createMember(c *gin.Context) {
	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	m, err := h.svc.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, m)
}

// updateMember partially updates an existing member.
func (h *Handler) updateMember(c *gin.Context) {
	id := c.Param("id")

	var req UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	m, err := h.svc.Update(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, m)
}

// deleteMember removes a member permanently.
func (h *Handler) deleteMember(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
