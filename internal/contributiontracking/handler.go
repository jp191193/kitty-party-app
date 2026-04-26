package contributiontracking

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/response"
)

// Handler exposes contribution tracking HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler constructs a tracking Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires endpoints onto the /api/v1 router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	tracking := rg.Group("/contribution-tracking")
	{
		tracking.POST("", h.create)
		tracking.GET("/:id", h.get)
		tracking.PATCH("/:id/payment", h.applyPayment)
		tracking.GET("/dues/:dueID", h.getByDueID)
		tracking.GET("/groups/:groupID", h.listByGroup)
		tracking.GET("/members/:memberID", h.listByMember)
	}
}

func (h *Handler) create(c *gin.Context) {
	var req CreateTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}
	t, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, t)
}

func (h *Handler) get(c *gin.Context) {
	t, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

func (h *Handler) getByDueID(c *gin.Context) {
	t, err := h.svc.GetByDueID(c.Request.Context(), c.Param("dueID"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

func (h *Handler) listByGroup(c *gin.Context) {
	groupID := c.Param("groupID")
	cycleMonth := parseIntQuery(c, "cycle_month")
	cycleYear := parseIntQuery(c, "cycle_year")

	list, err := h.svc.ListByGroup(c.Request.Context(), groupID, cycleMonth, cycleYear)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) listByMember(c *gin.Context) {
	list, err := h.svc.ListByMember(c.Request.Context(), c.Param("memberID"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) applyPayment(c *gin.Context) {
	var req UpdateTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}
	t, err := h.svc.ApplyPayment(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

func parseIntQuery(c *gin.Context, key string) *int {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &v
}
