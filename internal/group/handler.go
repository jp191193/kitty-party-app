// Package group – HTTP handler (transport) layer.
//
// Handlers translate HTTP ↔ domain objects and delegate all logic to the
// Service. They never touch the repository directly.
package group

import (
	"fmt"
	"net/http"

	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler holds a reference to the group Service.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler with the provided Service.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all group endpoints onto the provided router group.
//
// Routes:
//
//	GET    /groups                        – list all groups
//	GET    /groups/:id                    – get group by ID
//	GET    /groups?created_by=<memberID>  – filter by creator
//	POST   /groups                        – create a group
//	PUT    /groups/:id                    – update a group
//	DELETE /groups/:id                    – delete a group
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	groups := rg.Group("/groups")
	{
		groups.GET("", h.listGroups)
		groups.GET("/:id", h.getGroup)
		groups.POST("", h.createGroup)
		groups.PUT("/:id", h.updateGroup)
		groups.DELETE("/:id", h.deleteGroup)
	}
}

// listGroups returns all groups, or filters by ?created_by=<memberID>.
func (h *Handler) listGroups(c *gin.Context) {
	createdBy := c.Query("created_by")

	if createdBy != "" {
		groups, err := h.svc.GetByCreator(createdBy)
		if err != nil {
			response.Error(c, err)
			return
		}
		response.OK(c, groups)
		return
	}

	groups, err := h.svc.GetAll()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, groups)
}

// getGroup returns a single group by its UUID.
func (h *Handler) getGroup(c *gin.Context) {
	id := c.Param("id")
	g, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, g)
}

// createGroup creates a new kitty party group.
func (h *Handler) createGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	g, err := h.svc.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, g)
}

// updateGroup partially updates an existing group.
func (h *Handler) updateGroup(c *gin.Context) {
	id := c.Param("id")

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	g, err := h.svc.Update(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, g)
}

// deleteGroup permanently removes a group.
func (h *Handler) deleteGroup(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
