package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// jwtHeader is the fixed JWT header for HS256 (compatible with old backend).
var jwtHeader = mustB64JSON(map[string]string{"alg": "HS256", "typ": "JWT"})

type jwtPayload struct {
	PlayerID  int   `json:"pid"`
	CompanyID int   `json:"cid"`
	IssuedAt  int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
}

func mustB64JSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("jwt header marshal: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// SignJWT creates a signed HS256 JWT compatible with the old backend.
// Token expires after 72 hours.
func SignJWT(playerID, companyID int, signingKey string) (string, error) {
	now := time.Now()
	payload := jwtPayload{
		PlayerID:  playerID,
		CompanyID: companyID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(72 * time.Hour).Unix(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := jwtHeader + "." + encodedPayload
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

// ParseJWT verifies and decodes a HS256 JWT token.
// Returns (playerID, companyID, error).
func ParseJWT(tokenString, signingKey string) (int, int, error) {
	// Decode the token parts
	data, sig, err := splitToken(tokenString)
	if err != nil {
		return 0, 0, err
	}

	// Verify signature
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(data))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return 0, 0, fmt.Errorf("invalid signature")
	}

	// Decode payload
	parts := splitParts(tokenString)
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("malformed token")
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("decode payload: %w", err)
	}

	var payload jwtPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return 0, 0, fmt.Errorf("unmarshal payload: %w", err)
	}

	now := time.Now().Unix()
	if payload.PlayerID <= 0 || payload.CompanyID <= 0 {
		return 0, 0, fmt.Errorf("invalid token subject")
	}
	if payload.IssuedAt <= 0 || payload.ExpiresAt <= 0 || payload.ExpiresAt <= payload.IssuedAt {
		return 0, 0, fmt.Errorf("invalid token timestamps")
	}
	if payload.IssuedAt > now+60 {
		return 0, 0, fmt.Errorf("token issued in the future")
	}
	if now >= payload.ExpiresAt {
		return 0, 0, fmt.Errorf("token expired")
	}

	return payload.PlayerID, payload.CompanyID, nil
}

func splitToken(token string) (data, sig string, err error) {
	parts := splitParts(token)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("malformed token: expected 3 parts, got %d", len(parts))
	}
	return parts[0] + "." + parts[1], parts[2], nil
}

func splitParts(token string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}
