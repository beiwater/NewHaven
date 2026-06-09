package auth

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/newhaven/backend-next/internal/apperr"
	domain "github.com/newhaven/backend-next/internal/domain/auth"
)

const (
	minUsernameRunes = 3
	maxUsernameRunes = 32
	minPasswordBytes = 6
	maxPasswordBytes = 72
)

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func validateRegistration(req *domain.RegisterRequest) error {
	usernameRunes := utf8.RuneCountInString(req.Username)
	if usernameRunes < minUsernameRunes || usernameRunes > maxUsernameRunes {
		return apperr.Validation("username must be between 3 and 32 characters")
	}
	for _, r := range req.Username {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' {
			return apperr.Validation("username may only contain letters, numbers, underscores, and hyphens")
		}
	}
	if len(req.Password) < minPasswordBytes || len(req.Password) > maxPasswordBytes {
		return apperr.Validation("password must be between 6 and 72 bytes")
	}
	if utf8.RuneCountInString(req.Name) > 80 {
		return apperr.Validation("display name must be at most 80 characters")
	}
	if utf8.RuneCountInString(req.Gender) > 64 {
		return apperr.Validation("gender must be at most 64 characters")
	}
	if len(req.Email) > 254 {
		return apperr.Validation("email must be at most 254 characters")
	}
	if req.Email != "" {
		address, err := mail.ParseAddress(req.Email)
		if err != nil || address.Address != req.Email {
			return apperr.Validation("email must be valid")
		}
	}
	return nil
}
