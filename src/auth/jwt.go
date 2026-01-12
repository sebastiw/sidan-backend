package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// JWTClaims represents the claims stored in JWT
type JWTClaims struct {
	MemberNumber int64    `json:"member_number"`
	Email        string   `json:"email"`
	Scopes       []string `json:"scopes"`
	Provider     string   `json:"provider"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed JWT token
func GenerateJWT(memberNumber int64, email string, scopes []string, provider string, secret []byte) (string, error) {
	now := time.Now()

	claims := &JWTClaims{
		MemberNumber: memberNumber,
		Email:        email,
		Scopes:       scopes,
		Provider:     provider,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "sidan-backend",
			Subject:   email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateJWT validates and parses a JWT token
func ValidateJWT(tokenString string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, ErrInvalidToken
}

// ExtractBearer extracts token from "Bearer <token>" format
func ExtractBearer(authHeader string) string {
	const prefix = "Bearer "
	if len(authHeader) > len(prefix) && authHeader[:len(prefix)] == prefix {
		return authHeader[len(prefix):]
	}
	return ""
}
