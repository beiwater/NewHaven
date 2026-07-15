package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
)

func TestWriteAppErrTyped(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"validation", apperr.Validation("bad input"), http.StatusBadRequest, ErrorValidation},
		{"bad request", apperr.BadRequest("bad"), http.StatusBadRequest, ErrorBadRequest},
		{"unauthorized", apperr.Unauthorized("login"), http.StatusUnauthorized, ErrorUnauthorized},
		{"forbidden", apperr.Forbidden("no access"), http.StatusForbidden, ErrorForbidden},
		{"not found", apperr.NotFound("missing"), http.StatusNotFound, ErrorNotFound},
		{"conflict", apperr.Conflict("dupe"), http.StatusBadRequest, ErrorConflict},
		{"insufficient funds", apperr.InsufficientFunds("no money"), http.StatusBadRequest, ErrorInsufficientFunds},
		{"insufficient inventory", apperr.InsufficientInventory("no items"), http.StatusBadRequest, ErrorInsufficientInv},
		{"rate limited", apperr.RateLimited("slow down"), http.StatusTooManyRequests, ErrorRateLimited},
		{"internal", apperr.Internal("boom"), http.StatusInternalServerError, ErrorInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeAppErr(w, tt.err)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			body := w.Body.String()
			if !strings.Contains(body, tt.wantCode) {
				t.Errorf("body missing code %q: %s", tt.wantCode, body)
			}
		})
	}
}

func TestWriteAppErrUnknown(t *testing.T) {
	w := httptest.NewRecorder()
	writeAppErr(w, errors.New("some random error"))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, ErrorInternal) {
		t.Errorf("body missing internal error code: %s", body)
	}
}

func TestWriteAppErrNil(t *testing.T) {
	w := httptest.NewRecorder()
	writeAppErr(w, nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestWriteAppErrConflictReturns400(t *testing.T) {
	w := httptest.NewRecorder()
	writeAppErr(w, apperr.Conflict("already done"))
	if w.Code != http.StatusBadRequest {
		t.Errorf("conflict status = %d, want 400 (Phase 18 compat)", w.Code)
	}
	// But error code must still be CONFLICT.
	body := w.Body.String()
	if !strings.Contains(body, ErrorConflict) {
		t.Errorf("body missing CONFLICT code: %s", body)
	}
}

func TestWriteAppErrKindInternalSanitizesMessage(t *testing.T) {
	// Create an internal error with sensitive info in message.
	rawCause := errors.New("sensitive db error: connection refused")
	e := apperr.Internalf("crash: %v", rawCause)
	w := httptest.NewRecorder()
	writeAppErr(w, e)
	body := w.Body.String()
	if strings.Contains(body, "sensitive") {
		t.Errorf("internal error leaked sensitive info: %s", body)
	}
	if strings.Contains(body, rawCause.Error()) {
		t.Errorf("internal error leaked cause: %s", body)
	}
	if !strings.Contains(body, "an unexpected error occurred") {
		t.Errorf("body missing generic message: %s", body)
	}
}

func TestHttpStatusForKind(t *testing.T) {
	tests := []struct {
		kind apperr.Kind
		want int
	}{
		{apperr.KindValidation, http.StatusBadRequest},
		{apperr.KindBadRequest, http.StatusBadRequest},
		{apperr.KindUnauthorized, http.StatusUnauthorized},
		{apperr.KindForbidden, http.StatusForbidden},
		{apperr.KindNotFound, http.StatusNotFound},
		{apperr.KindConflict, http.StatusBadRequest},
		{apperr.KindInsufficientFunds, http.StatusBadRequest},
		{apperr.KindInsufficientInventory, http.StatusBadRequest},
		{apperr.KindRateLimited, http.StatusTooManyRequests},
		{apperr.KindInternal, http.StatusInternalServerError},
		{"UNKNOWN_KIND", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			got := httpStatusForKind(tt.kind)
			if got != tt.want {
				t.Errorf("httpStatusForKind(%q) = %d, want %d", tt.kind, got, tt.want)
			}
		})
	}
}

func TestErrorCodeForKind(t *testing.T) {
	tests := []struct {
		kind apperr.Kind
		want string
	}{
		{apperr.KindValidation, ErrorValidation},
		{apperr.KindNotFound, ErrorNotFound},
		{apperr.KindInternal, ErrorInternal},
		{"UNKNOWN", ErrorInternal},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			got := errorCodeForKind(tt.kind)
			if got != tt.want {
				t.Errorf("errorCodeForKind(%q) = %q, want %q", tt.kind, got, tt.want)
			}
		})
	}
}
