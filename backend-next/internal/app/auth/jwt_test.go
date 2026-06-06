package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestParseJWTRejectsInvalidClaims(t *testing.T) {
	for _, payload := range []jwtPayload{
		{PlayerID: 0, CompanyID: 1, IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Hour).Unix()},
		{PlayerID: 1, CompanyID: 0, IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Hour).Unix()},
		{PlayerID: 1, CompanyID: 1, IssuedAt: time.Now().Add(time.Hour).Unix(), ExpiresAt: time.Now().Add(2 * time.Hour).Unix()},
		{PlayerID: 1, CompanyID: 1, IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Unix()},
	} {
		if _, _, err := ParseJWT(signPayloadForTest(t, payload), "test-secret"); err == nil {
			t.Fatalf("expected invalid claims to fail: %+v", payload)
		}
	}
}

func signPayloadForTest(t *testing.T, payload jwtPayload) string {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	input := jwtHeader + "." + encoded
	mac := hmac.New(sha256.New, []byte("test-secret"))
	_, _ = mac.Write([]byte(input))
	return input + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
