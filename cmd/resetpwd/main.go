// resetpwd is an admin CLI tool for resetting a member's password.
//
// Usage:
//
//	go run ./cmd/resetpwd -email user@example.com -password newpassword123
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	email    := flag.String("email",    "", "Email of the member whose password will be reset (required)")
	password := flag.String("password", "", "New plain-text password — min 8 characters (required)")
	flag.Parse()

	if *email == "" || *password == "" {
		flag.Usage()
		os.Exit(1)
	}
	if len(*password) < 8 {
		log.Fatal("password must be at least 8 characters")
	}

	// Load .env (ignore error — env vars may already be set in production)
	_ = godotenv.Load()

	dbURL := buildDatabaseURL()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Verify the member exists first
	var memberID, name string
	err = pool.QueryRow(ctx,
		`SELECT id, name FROM users WHERE email = $1`, *email,
	).Scan(&memberID, &name)
	if err != nil {
		log.Fatalf("member not found for email %q: %v", *email, err)
	}

	// Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	// Update the record
	tag, err := pool.Exec(ctx,
		`UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`,
		string(hash), memberID,
	)
	if err != nil {
		log.Fatalf("update: %v", err)
	}
	if tag.RowsAffected() == 0 {
		log.Fatalf("no rows updated — something went wrong")
	}

	fmt.Printf("✓ Password reset for %s <%s>\n", name, *email)
}

func buildDatabaseURL() string {
	get := func(key, fallback string) string {
		if v, ok := os.LookupEnv(key); ok && v != "" {
			return v
		}
		return fallback
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		get("DB_USER", "postgres"),
		get("DB_PASSWORD", "postgres"),
		get("DB_HOST", "localhost"),
		get("DB_PORT", "5432"),
		get("DB_NAME", "kitty_party"),
	)
}
