package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	userID := flag.String("userid", "c1e663a4-50df-435b-b8c4-b5eea7369ebf", "User ID to encode in token")
	email := flag.String("email", "jay@example.com", "Email to encode in token")
	flag.Parse()

	// This must match the secret in middleware/auth.go
	secret := []byte("super_secret_key")

	claims := jwt.MapClaims{
		"user_id": *userID,
		"email":   *email,
		"exp":     time.Now().Add(time.Hour * 24 * 365).Unix(), // Valid for 1 year for dev purposes
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n--- JWT for user_id: %s ---\n", *userID)
	fmt.Println("Bearer " + tokenString)
	fmt.Println("\n(Copy the entire line starting with 'Bearer' into your Authorization header)")
}
