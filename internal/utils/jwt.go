package utils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateCoinbaseJWT generates a JWT token for Coinbase API authentication
func GenerateCoinbaseJWT(requestMethod, requestPath string) (string, error) {
	// Get configuration from environment
	keyID := os.Getenv("COINBASE_API_KEY_ID")
	keySecret := os.Getenv("COINBASE_API_SECRET")
	requestHost := os.Getenv("COINBASE_API_URL")

	if keyID == "" || keySecret == "" || requestHost == "" {
		return "", errors.New("missing required environment variables for JWT generation")
	}

	// Decode the base64-encoded secret
	decodedSecret, err := base64.StdEncoding.DecodeString(keySecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode API secret: %v", err)
	}

	// Create the private key
	privateKey := ed25519.NewKeyFromSeed(decodedSecret[:32])

	// Create the JWT claims
	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"sub": keyID,
		"iss": "cdp",
		"aud": []string{"cdp_service"},
		"nbf": now,
		"exp": now + 120, // 2 minutes expiration
		"uri": fmt.Sprintf("%s %s%s", requestMethod, requestHost, requestPath),
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	// Set the key ID in the header
	token.Header["kid"] = keyID

	// Generate the JWT string
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %v", err)
	}

	return tokenString, nil
}

// GenerateAuthHeaders generates the necessary headers for Coinbase API authentication
func GenerateAuthHeaders(method, path string) (map[string]string, error) {
	token, err := GenerateCoinbaseJWT(method, path)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}, nil
}
