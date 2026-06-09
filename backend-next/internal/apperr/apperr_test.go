package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorImplementsError(t *testing.T) {
	e := NotFound("order not found")
	if e.Error() != "order not found" {
		t.Errorf("Error() = %q, want %q", e.Error(), "order not found")
	}
}

func TestErrorWrapping(t *testing.T) {
	cause := errors.New("db failure")
	e := Wrap(KindInternal, cause)
	if !errors.Is(e, cause) {
		t.Error("errors.Is should find the cause")
	}
	if e.Kind != KindInternal {
		t.Errorf("Kind = %q, want %q", e.Kind, KindInternal)
	}
}

func TestErrorWrapMsg(t *testing.T) {
	cause := errors.New("underlying error")
	e := WrapMsg(KindNotFound, "resource not found", cause)
	if e.Message != "resource not found" {
		t.Errorf("Message = %q, want %q", e.Message, "resource not found")
	}
	if !errors.Is(e, cause) {
		t.Error("errors.Is should find the cause")
	}
}

func TestHasKind(t *testing.T) {
	e := Validation("invalid input")
	if !HasKind(e, KindValidation) {
		t.Error("HasKind should match KindValidation")
	}
	if HasKind(e, KindNotFound) {
		t.Error("HasKind should not match KindNotFound")
	}
	// Non-typed errors
	if HasKind(errors.New("raw"), KindInternal) {
		t.Error("HasKind on raw error should return false")
	}
}

func TestKindOf(t *testing.T) {
	if KindOf(Validation("x")) != KindValidation {
		t.Error("KindOf(Validation) should be KindValidation")
	}
	if KindOf(errors.New("raw")) != KindInternal {
		t.Error("KindOf(raw error) should be KindInternal")
	}
	if KindOf(nil) != KindInternal {
		t.Error("KindOf(nil) should be KindInternal")
	}
}

func TestIsUserError(t *testing.T) {
	if !IsUserError(NotFound("x")) {
		t.Error("NotFound error should be user error")
	}
	if IsUserError(errors.New("raw")) {
		t.Error("raw error should not be user error")
	}
}

func TestConstructors(t *testing.T) {
	tests := []struct {
		name string
		e    *Error
		kind Kind
	}{
		{"Validation", Validation("a"), KindValidation},
		{"Validationf", Validationf("a %d", 1), KindValidation},
		{"BadRequest", BadRequest("a"), KindBadRequest},
		{"BadRequestf", BadRequestf("a %d", 1), KindBadRequest},
		{"Unauthorized", Unauthorized("a"), KindUnauthorized},
		{"Forbidden", Forbidden("a"), KindForbidden},
		{"NotFound", NotFound("a"), KindNotFound},
		{"NotFoundf", NotFoundf("a %d", 1), KindNotFound},
		{"Conflict", Conflict("a"), KindConflict},
		{"InsufficientFunds", InsufficientFunds("a"), KindInsufficientFunds},
		{"InsufficientInventory", InsufficientInventory("a"), KindInsufficientInventory},
		{"RateLimited", RateLimited("a"), KindRateLimited},
		{"Internal", Internal("a"), KindInternal},
		{"Internalf", Internalf("a %d", 1), KindInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.e.Kind != tt.kind {
				t.Errorf("Kind = %q, want %q", tt.e.Kind, tt.kind)
			}
			if tt.e.Error() == "" {
				t.Error("Error() should not be empty")
			}
		})
	}
}

func TestWrapNilCause(t *testing.T) {
	e := Wrap(KindNotFound, nil)
	if e.Kind != KindNotFound {
		t.Errorf("Kind = %q", e.Kind)
	}
	if e.Message != string(KindNotFound) {
		t.Errorf("Message = %q, want %q", e.Message, string(KindNotFound))
	}
	if e.Unwrap() != nil {
		t.Error("Unwrap of nil cause should be nil")
	}
}

func TestErrorsIsChain(t *testing.T) {
	/* Callers can use errors.Is on an *Error to reach its cause. */
	leaf := errors.New("leaf")
	cause := fmt.Errorf("wrapped: %w", leaf)
	e := Wrap(KindInternal, cause)
	if !errors.Is(e, cause) {
		t.Error("errors.Is should traverse the chain to cause")
	}
	if !errors.Is(e, leaf) {
		t.Error("errors.Is should traverse through fmt.Errorf to leaf")
	}
}
