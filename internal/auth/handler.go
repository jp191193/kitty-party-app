package auth

import (
	"fmt"
	"net/http"
	"time"

	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type GenerateTokenRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type GenerateTokenResponse struct {
	Token string `json:"token"`
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/token", h.generateToken)
	}
}

func (h *Handler) generateToken(c *gin.Context) {
	var req GenerateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": req.UserID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(middleware.JWTSecret)
	if err != nil {
		response.Error(c, apperrors.New(http.StatusInternalServerError, "failed to sign token"))
		return
	}

	response.OK(c, GenerateTokenResponse{Token: tokenString})
}
