package auth

import (
	"fmt"
	"net/http"
	"time"

	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/member"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string         `json:"token"`
	User  *member.Member `json:"user"`
}

type Handler struct {
	memberRepo member.Repository
}

func NewHandler(memberRepo member.Repository) *Handler {
	return &Handler{memberRepo: memberRepo}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.login)
		auth.POST("/logout", h.logout)
	}
}

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error())))
		return
	}

	creds, err := h.memberRepo.FindByEmail(req.Email)
	if err != nil {
		// Generic message to prevent email enumeration
		response.Error(c, apperrors.New(http.StatusUnauthorized, "invalid email or password"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, apperrors.New(http.StatusUnauthorized, "invalid email or password"))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     creds.Member.ID,
		"global_role": creds.Member.GlobalRole,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(middleware.JWTSecret)
	if err != nil {
		response.Error(c, apperrors.New(http.StatusInternalServerError, "failed to sign token"))
		return
	}

	response.OK(c, LoginResponse{Token: tokenString, User: creds.Member})
}

// logout is a client-side concern for stateless JWT; the endpoint signals success
// so the frontend can handle the flow uniformly.
func (h *Handler) logout(c *gin.Context) {
	response.NoContent(c)
}
