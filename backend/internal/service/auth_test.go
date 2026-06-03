package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
	"go-sim-api/internal/middleware"
	"go-sim-api/internal/model"
)

func TestRegisterPlayerWithPassword(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "test-key-123"
	svc := New(testData(), cfg, nil)

	result, err := svc.RegisterPlayer("alice", "secret123")
	if err != nil {
		t.Fatalf("RegisterPlayer failed: %v", err)
	}

	p := result["player"].(model.Player)
	if p.Username != "alice" {
		t.Errorf("expected username alice, got %s", p.Username)
	}
	if p.Token == "" {
		t.Errorf("expected non-empty token")
	}
	if result["companyId"] == 0 {
		t.Errorf("expected non-zero companyId")
	}

	pid, cid, err := middleware.ParseJWT(p.Token, "test-key-123")
	if err != nil {
		t.Fatalf("ParseJWT failed on returned token: %v", err)
	}
	if pid <= 0 {
		t.Errorf("expected positive playerID, got %d", pid)
	}
	if cid <= 0 {
		t.Errorf("expected positive companyID, got %d", cid)
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "test-key"
	svc := New(testData(), cfg, nil)

	_, err := svc.RegisterPlayer("bob", "pass1")
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	_, err = svc.RegisterPlayer("bob", "pass2")
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}
}

func TestLoginWithCorrectPassword(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "login-test-key"
	svc := New(testData(), cfg, nil)

	_, err := svc.RegisterPlayer("carol", "my-password")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	result, err := svc.LoginPlayer("carol", "my-password")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	tokenStr := result["token"].(string)
	pid, cid, err := middleware.ParseJWT(tokenStr, "login-test-key")
	if err != nil {
		t.Fatalf("ParseJWT failed on login token: %v", err)
	}
	if pid <= 0 || cid <= 0 {
		t.Errorf("invalid pid=%d or cid=%d from parsed JWT", pid, cid)
	}
}

func TestLoginWithWrongPassword(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "test-key"
	svc := New(testData(), cfg, nil)

	svc.RegisterPlayer("dave", "correct-password")
	_, err := svc.LoginPlayer("dave", "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestLoginWithNonexistentUser(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "test-key"
	svc := New(testData(), cfg, nil)

	_, err := svc.LoginPlayer("nobody", "any-password")
	if err == nil {
		t.Fatal("expected error for nonexistent user, got nil")
	}
}

func TestPasswordHashNotReturned(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "test-key"
	svc := New(testData(), cfg, nil)

	result, err := svc.RegisterPlayer("eve", "safe-password")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// The PasswordHash field is tagged `json:"-"`, verify it doesn't appear in JSON output.
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	// Check that the player field inside the JSON doesn't contain passwordHash
	playerRaw := m["player"]
	var playerMap map[string]any
	json.Unmarshal(playerRaw, &playerMap)
	if _, exists := playerMap["passwordHash"]; exists {
		t.Error("passwordHash leaked in JSON output")
	}
	if _, exists := playerMap["PasswordHash"]; exists {
		t.Error("PasswordHash leaked in JSON output")
	}

	// But internally the hash IS set
	p := result["player"].(model.Player)
	if p.PasswordHash == "" {
		t.Fatal("PasswordHash must be set internally")
	}
	if p.PasswordHash == "safe-password" {
		t.Fatal("PasswordHash must be a bcrypt hash, not the raw password")
	}
}

func TestJWTExpiry(t *testing.T) {
	expired := createExpiredJWT(1, 1001, "test-key")
	_, _, err := middleware.ParseJWT(expired, "test-key")
	if err == nil {
		t.Fatal("expected error for expired JWT, got nil")
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	validToken, err := middleware.SignJWT(1, 1001, "real-key")
	if err != nil {
		t.Fatalf("SignJWT failed: %v", err)
	}
	_, _, err = middleware.ParseJWT(validToken, "wrong-key")
	if err == nil {
		t.Fatal("expected error for wrong signing key, got nil")
	}
}

func TestJWTTamperedPayload(t *testing.T) {
	validToken, err := middleware.SignJWT(1, 1001, "test-key")
	if err != nil {
		t.Fatalf("SignJWT failed: %v", err)
	}

	parts := splitToken(validToken)
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}
	tampered := parts[0] + ".eyJwaWQiOjk5OX0" + "." + parts[2]
	_, _, err = middleware.ParseJWT(tampered, "test-key")
	if err == nil {
		t.Fatal("expected error for tampered JWT, got nil")
	}
}

func TestTokenNotBase64Username(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "not-base64-test"
	svc := New(testData(), cfg, nil)

	result, _ := svc.RegisterPlayer("mallory", "password")
	p := result["player"].(model.Player)
	tokenStr := p.Token

	decoded, err := decodeBase64Quick(tokenStr)
	if err == nil && string(decoded) == "mallory" {
		t.Error("Token appears to be base64(username) - this is insecure!")
	}
	if tokenStr == "mallory" {
		t.Error("Token is literally the username - this is insecure!")
	}
}

func TestMultiPlayerIsolation(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "isolation-key"
	svc := New(testData(), cfg, nil)

	r1, _ := svc.RegisterPlayer("player1", "pass1")
	r2, _ := svc.RegisterPlayer("player2", "pass2")

	p1 := r1["player"].(model.Player)
	p2 := r2["player"].(model.Player)

	if p1.CompanyID == p2.CompanyID {
		t.Fatal("players must have different company IDs")
	}

	snap := svc.Snapshot()
	c1 := snap.GetCompany(p1.CompanyID)
	c2 := snap.GetCompany(p2.CompanyID)
	if c1 == nil || c2 == nil {
		t.Fatal("expected both companies to exist")
	}
	if c1.ID == c2.ID {
		t.Error("companies should have different IDs")
	}
}

func TestDevModeCreatesDevPlayer(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.JWTSigningKey = "dev-key"
	svc := New(testData(), cfg, nil)

	if len(svc.State.Players) != 1 {
		t.Fatalf("expected 1 player in dev mode, got %d", len(svc.State.Players))
	}
	if svc.State.Players[0].Username != "dev" {
		t.Errorf("expected dev username, got %s", svc.State.Players[0].Username)
	}
	if len(svc.State.Companies) < 1 {
		t.Fatal("expected at least 1 company in dev mode")
	}
}

func TestNoDevModeNoPlayers(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "prod-key"
	svc := New(testData(), cfg, nil)

	if len(svc.State.Players) != 0 {
		t.Errorf("expected 0 players in non-dev mode, got %d", len(svc.State.Players))
	}
	for _, c := range svc.State.Companies {
		if c.ID == cfg.Game.CompanyID {
			t.Error("non-dev mode should not contain the dev company from game.json")
		}
	}
}

func TestJWTRoundTrip(t *testing.T) {
	token, err := middleware.SignJWT(42, 1000042, "roundtrip-key")
	if err != nil {
		t.Fatalf("SignJWT failed: %v", err)
	}

	pid, cid, err := middleware.ParseJWT(token, "roundtrip-key")
	if err != nil {
		t.Fatalf("ParseJWT failed: %v", err)
	}
	if pid != 42 {
		t.Errorf("expected playerID 42, got %d", pid)
	}
	if cid != 1000042 {
		t.Errorf("expected companyID 1000042, got %d", cid)
	}
}

func TestValidateTokenFallback(t *testing.T) {
	cfg := config.DefaultTestConfig()
	cfg.DevMode = false
	cfg.JWTSigningKey = "fallback-key"
	svc := New(testData(), cfg, nil)

	result, _ := svc.RegisterPlayer("fallback-user", "pw")
	p := result["player"].(model.Player)

	pid, cid, ok := svc.ValidateToken(p.Token)
	if !ok {
		t.Fatal("ValidateToken should accept a valid JWT")
	}
	if pid != p.ID || cid != p.CompanyID {
		t.Errorf("wrong ids: got pid=%d cid=%d, expected pid=%d cid=%d", pid, cid, p.ID, p.CompanyID)
	}
}

// --- helpers ---
func testData() *data.StaticData {
	return &data.StaticData{
		Resources: []map[string]any{
			{"id": 1, "name": "Wheat", "dbLetter": 1, "producedPerHourRaw": 100.0},
		},
		EconomyModel: map[string]any{"models": map[string]any{}},
	}
}

func createExpiredJWT(playerID, companyID int, key string) string {
	payloadB := fmt.Sprintf(`{"pid":%d,"cid":%d,"iat":1000000000,"exp":900000000}`, playerID, companyID)
	headerB64 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	payloadB64 := base64RawURLEncode([]byte(payloadB))
	sigInput := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(sigInput))
	sig := base64RawURLEncode(mac.Sum(nil))
	return sigInput + "." + sig
}

func base64RawURLEncode(b []byte) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	n := len(b)
	out := make([]byte, 0, (n+2)/3*4)
	for i := 0; i < n; i += 3 {
		var val uint32
		val = uint32(b[i]) << 16
		if i+1 < n {
			val |= uint32(b[i+1]) << 8
		}
		if i+2 < n {
			val |= uint32(b[i+2])
		}
		out = append(out, alphabet[(val>>18)&0x3F])
		out = append(out, alphabet[(val>>12)&0x3F])
		if i+1 < n {
			out = append(out, alphabet[(val>>6)&0x3F])
		}
		if i+2 < n {
			out = append(out, alphabet[val&0x3F])
		}
	}
	return string(out)
}

func splitToken(token string) []string {
	parts := make([]string, 0)
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	if start < len(token) {
		parts = append(parts, token[start:])
	}
	return parts
}

func decodeBase64Quick(s string) ([]byte, error) {
	n := len(s)
	padding := (4 - n%4) % 4
	for i := 0; i < padding; i++ {
		s += "="
	}
	out := make([]byte, 0, n*3/4)
	val := 0
	bits := -8
	for _, c := range s {
		if c == '=' {
			break
		}
		var v byte
		switch {
		case c >= 'A' && c <= 'Z':
			v = byte(c - 'A')
		case c >= 'a' && c <= 'z':
			v = byte(c - 'a' + 26)
		case c >= '0' && c <= '9':
			v = byte(c - '0' + 52)
		case c == '+' || c == '-':
			v = 62
		case c == '/' || c == '_':
			v = 63
		default:
			return nil, fmt.Errorf("invalid base64 char: %c", c)
		}
		val = (val << 6) | int(v)
		bits += 6
		if bits >= 0 {
			out = append(out, byte((val>>bits)&0xFF))
			bits -= 8
		}
	}
	return out, nil
}
