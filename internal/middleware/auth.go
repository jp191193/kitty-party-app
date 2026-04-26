package middleware

import (
	"errors"
	"net/http"
	"strings"
	"kitty-party-app/internal/response"
	"kitty-party-app/internal/apperrors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte("super_secret_key") // In a real app, this should come from .env

// AuthMiddleware validates the JWT token and extracts the user_id
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
			// Verify the signing method
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

		c.Set("user_id", userID)
		c.Next()
	}
}
