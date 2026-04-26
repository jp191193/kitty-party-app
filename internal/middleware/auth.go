package middleware

import (
	"errors"
	"net/http"
	"strings"

	"kitty-party-app/internal/apperrors"
	"kitty-party-app/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTSecret is set at startup from config (JWT_SECRET env var).
var JWTSecret = []byte("super_secret_key")

// GroupRoleProvider is satisfied by group.MembershipRepository.
// Defined here to avoid an import cycle (group imports middleware).
type GroupRoleProvider interface {
	GetRole(groupID, memberID string) (string, error)
}

// AuthMiddleware validates the JWT token and sets user_id + global_role in context.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, apperrors.New(http.StatusUnauthorized, "Missing Authorization header"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, apperrors.New(http.StatusUnauthorized, "Invalid Authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return JWTSecret, nil
		})

		if err != nil || !token.Valid {
			response.Error(c, apperrors.New(http.StatusUnauthorized, "Invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, apperrors.New(http.StatusUnauthorized, "Invalid token claims"))
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			response.Error(c, apperrors.New(http.StatusUnauthorized, "user_id not found in token"))
			c.Abort()
			return
		}

		globalRole, _ := claims["global_role"].(string)
		if globalRole == "" {
			globalRole = "User"
		}

		c.Set("user_id", userID)
		c.Set("global_role", globalRole)
		c.Next()
	}
}

// RequireSuperAdmin rejects requests from non-SuperAdmin users.
// Must be chained after AuthMiddleware.
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("global_role") != "SuperAdmin" {
			response.Error(c, apperrors.New(http.StatusForbidden, "SuperAdmin access required"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireGroupRole checks that the calling user has one of the specified roles
// in the group identified by the URL parameter named groupIDParam.
// SuperAdmins bypass this check automatically.
// Must be chained after AuthMiddleware.
func RequireGroupRole(provider GroupRoleProvider, groupIDParam string, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// SuperAdmin bypasses all group-role checks
		if c.GetString("global_role") == "SuperAdmin" {
			c.Next()
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			response.Error(c, apperrors.New(http.StatusUnauthorized, "user not authenticated"))
			c.Abort()
			return
		}

		groupID := c.Param(groupIDParam)
		if groupID == "" {
			response.Error(c, apperrors.New(http.StatusBadRequest, "group ID missing from URL"))
			c.Abort()
			return
		}

		role, err := provider.GetRole(groupID, userID)
		if err != nil || role == "" {
			response.Error(c, apperrors.New(http.StatusForbidden, "not a member of this group"))
			c.Abort()
			return
		}

		for _, r := range allowedRoles {
			if role == r {
				c.Set("group_role", role)
				c.Next()
				return
			}
		}

		response.Error(c, apperrors.New(http.StatusForbidden, "insufficient permissions in this group"))
		c.Abort()
	}
}
