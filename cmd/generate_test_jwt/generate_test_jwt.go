package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/sebastiw/sidan-backend/src/auth"
	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/logger"
)

func main() {
	// Initialize logger and config
	logger.SetupLogging()
	config.Init()

	// Get member ID from command line or use default
	memberID := int64(1)
	email := "test@example.com"

	if len(os.Args) > 1 {
		id, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			log.Fatal("Invalid member ID. Usage: go run generate_test_jwt.go [memberID] [email]")
		}
		memberID = id
	}

	if len(os.Args) > 2 {
		email = os.Args[2]
	}

	// Grant all scopes for testing
	scopes := []string{
		auth.ReadMemberScope,
		auth.ModifyEntryScope,
		auth.WriteImageScope,
		auth.WriteEmailScope,
		auth.WriteMemberScope,
		auth.WriteArrScope,
		auth.ReadArticleScope,
		auth.WriteArticleScope,
		auth.UseAdvancedFilterScope,
	}

	// Generate JWT
	token, err := auth.GenerateJWT(memberID, email, scopes, "test", config.GetJWTSecret())
	if err != nil {
		log.Fatal("Failed to generate JWT:", err)
	}

	fmt.Println("========================================")
	fmt.Println("Generated Test JWT Token")
	fmt.Println("========================================")
	fmt.Printf("Member ID: %d\n", memberID)
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Scopes: %v\n", scopes)
	fmt.Println("\nJWT Token:")
	fmt.Println(token)
	fmt.Println("\nTo use this token, run:")
	fmt.Printf("export SIDAN_JWT='%s'\n", token)
	fmt.Println("\nThen run the test script:")
	fmt.Println("./test_jwt_auth.sh")
}
