package main

import (
	"fmt"
	"os"

	"github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/logger"
)

func main() {
	logger.SetupLogging()
	config.Init()

	secret := config.GetJWTSecret()
	envSecret := os.Getenv("JWT_SECRET")

	fmt.Printf("Environment JWT_SECRET: %s\n", envSecret)
	fmt.Printf("Config GetJWTSecret(): %s\n", string(secret))
	fmt.Printf("Secrets match: %v\n", string(secret) == envSecret)
	fmt.Println()

	// Generate a token
	token, err := auth.GenerateJWT(1, "test@example.com", []string{"test"}, "test", secret)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return
	}
	fmt.Println("Generated token (first 50 chars):", token[:50])

	// Try to validate it
	claims, err := auth.ValidateJWT(token, secret)
	if err != nil {
		fmt.Println("Error validating token:", err)
		return
	}
	fmt.Println("Token validated successfully!")
	fmt.Printf("Claims: MemberNumber=%d, Email=%s\n", claims.MemberNumber, claims.Email)
}
