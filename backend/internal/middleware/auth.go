package middleware

import (
	"crypto/rand"
	"encoding/hex"
)

type ctxKey string

const PlayerIDKey ctxKey = "player_id"
const CompanyIDKey ctxKey = "company_id"

func GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
