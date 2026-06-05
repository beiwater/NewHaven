package httpapi

import (
	"context"
	"net/http"

	"github.com/newhaven/backend-next/internal/app/auth"
)

type ctxKey string

const (
	PlayerIDKey  ctxKey = "player_id"
	CompanyIDKey ctxKey = "company_id"
)

// PlayerIDFromCtx extracts player ID from the request context.
func PlayerIDFromCtx(ctx context.Context) (int, bool) {
	v, ok := ctx.Value(PlayerIDKey).(int)
	return v, ok
}

// CompanyIDFromCtx extracts company ID from the request context.
func CompanyIDFromCtx(ctx context.Context) (int, bool) {
	v, ok := ctx.Value(CompanyIDKey).(int)
	return v, ok
}

// AuthRequired is middleware that validates JWT Bearer token.
// It extracts player_id and company_id into request context.
func AuthRequired(jwtKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				writeErr(w, 401, ErrorUnauthorized, "missing authorization token", nil)
				return
			}

			playerID, companyID, err := auth.ParseJWT(tokenStr, jwtKey)
			if err != nil {
				writeErr(w, 401, ErrorUnauthorized, "invalid or expired token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), PlayerIDKey, playerID)
			ctx = context.WithValue(ctx, CompanyIDKey, companyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && h[:7] == "Bearer " {
		return h[7:]
	}
	return ""
}
