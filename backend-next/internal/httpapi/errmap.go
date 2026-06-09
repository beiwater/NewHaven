package httpapi

import (
	"errors"
	"net/http"

	"github.com/newhaven/backend-next/internal/apperr"
)

// httpStatusForKind returns the HTTP status code for an apperr Kind.
func httpStatusForKind(k apperr.Kind) int {
	switch k {
	case apperr.KindValidation:
		return http.StatusBadRequest
	case apperr.KindBadRequest:
		return http.StatusBadRequest
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindConflict:
		return http.StatusBadRequest
	case apperr.KindInsufficientFunds:
		return http.StatusBadRequest
	case apperr.KindInsufficientInventory:
		return http.StatusBadRequest
	case apperr.KindRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// errorCodeForKind returns the response error code string for an apperr Kind.
func errorCodeForKind(k apperr.Kind) string {
	switch k {
	case apperr.KindValidation:
		return ErrorValidation
	case apperr.KindBadRequest:
		return ErrorBadRequest
	case apperr.KindUnauthorized:
		return ErrorUnauthorized
	case apperr.KindForbidden:
		return ErrorForbidden
	case apperr.KindNotFound:
		return ErrorNotFound
	case apperr.KindConflict:
		return ErrorConflict
	case apperr.KindInsufficientFunds:
		return ErrorInsufficientFunds
	case apperr.KindInsufficientInventory:
		return ErrorInsufficientInv
	case apperr.KindRateLimited:
		return ErrorRateLimited
	default:
		return ErrorInternal
	}
}

// writeAppErr sends a typed application error as an HTTP response.
func writeAppErr(w http.ResponseWriter, err error) {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		status := httpStatusForKind(appErr.Kind)
		code := errorCodeForKind(appErr.Kind)
		msg := appErr.Message
		// Sanitize: never send internal error details to client.
		if appErr.Kind == apperr.KindInternal {
			msg = "an unexpected error occurred"
		}
		writeErr(w, status, code, msg, nil)
		return
	}
	// Unknown error: write internal error.
	writeErr(w, http.StatusInternalServerError, ErrorInternal, "an unexpected error occurred", nil)
}
