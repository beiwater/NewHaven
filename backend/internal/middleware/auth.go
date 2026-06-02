package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ctxKey string

const PlayerIDKey ctxKey = "player_id"
const CompanyIDKey ctxKey = "company_id"

// jwtHeader is the fixed JWT header for HS256.
var jwtHeader = mustB64JSON(map[string]string{"alg": "HS256", "typ": "JWT"})

func mustB64JSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("jwt header marshal: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// jwtPayload represents the claims embedded in a JWT.
type jwtPayload struct {
	PlayerID  int   `json:"pid"`
	CompanyID int   `json:"cid"`
	IssuedAt  int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
}

// SignJWT creates a signed HS256 JWT token embedding playerID and companyID.
// The token expires after 72 hours.
func SignJWT(playerID, companyID int, signingKey string) (string, error) {
	if signingKey == "" {
		return "", fmt.Errorf("JWT signing key is empty")
	}
	now := time.Now().Unix()
	payload := jwtPayload{
		PlayerID:  playerID,
		CompanyID: companyID,
		IssuedAt:  now,
		ExpiresAt: now + 72*3600, // 72 hours
	}
	payloadB, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("jwt payload marshal: %w", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadB)

	sigInput := jwtHeader + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return sigInput + "." + sig, nil
}

// ParseJWT verifies and decodes a HS256 JWT token.
// Returns (playerID, companyID, error).
func ParseJWT(tokenString, signingKey string) (int, int, error) {
	if signingKey == "" {
		return 0, 0, fmt.Errorf("JWT signing key is empty")
	}
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Verify signature
	sigInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(sigInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return 0, 0, fmt.Errorf("invalid JWT signature")
	}

	// Decode payload
	payloadB, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid JWT payload encoding: %w", err)
	}
	var payload jwtPayload
	if err := json.Unmarshal(payloadB, &payload); err != nil {
		return 0, 0, fmt.Errorf("invalid JWT payload JSON: %w", err)
	}

	// Check expiry
	if time.Now().Unix() > payload.ExpiresAt {
		return 0, 0, fmt.Errorf("JWT token expired")
	}

	return payload.PlayerID, payload.CompanyID, nil
}

// GenerateCSRFToken generates a cryptographically random hex string for CSRF.
func GenerateCSRFToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
