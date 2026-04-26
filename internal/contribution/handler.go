package contribution

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/response"
)

// Handler manages the HTTP endpoints for dues and payments.
type Handler struct {
	svc Service
}

// NewHandler constructs a new Contribution Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires the endpoints onto the provided router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	contributions := rg.Group("/contributions")
	{
		// /api/v1/contributions/dues/generate
		contributions.POST("/dues/generate", h.generateDues)

		// /api/v1/contributions/groups/:groupID/dues
		contributions.GET("/groups/:groupID/dues", h.listGroupDues)

		// /api/v1/contributions/payments
		contributions.POST("/payments", h.recordPayment)
	}
}

func (h *Handler) generateDues(c *gin.Context) {
	var req GenerateDuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	dues, err := h.svc.GenerateMonthlyDues(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dues)
}

func (h *Handler) listGroupDues(c *gin.Context) {
	groupID := c.Param("groupID")
	
	var cycleMonth *int
	monthStr := c.Query("cycle_month")
	if monthStr != "" {
		m, err := strconv.Atoi(monthStr)
		if err == nil {
			cycleMonth = &m
		}
	}

	dues, err := h.svc.GetGroupDues(c.Request.Context(), groupID, cycleMonth)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dues)
}

func (h *Handler) recordPayment(c *gin.Context) {
	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	payment, err := h.svc.RecordPayment(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, payment)
}
