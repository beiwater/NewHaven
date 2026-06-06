// Package apperr provides the typed application error model.
// Services return *Error to convey a stable kind and public-safe message.
// Handlers use errors.As to map kinds to HTTP status/code without string matching.
//
// This package MUST NOT import any other backend-next internal package.
package apperr

import (
	"errors"
	"fmt"
)

// Kind identifies the category of an application error.
type Kind string

const (
	KindValidation            Kind = "VALIDATION"
	KindBadRequest            Kind = "BAD_REQUEST"
	KindUnauthorized          Kind = "UNAUTHORIZED"
	KindForbidden             Kind = "FORBIDDEN"
	KindNotFound              Kind = "NOT_FOUND"
	KindConflict              Kind = "CONFLICT"
	KindInsufficientFunds     Kind = "INSUFFICIENT_FUNDS"
	KindInsufficientInventory Kind = "INSUFFICIENT_INVENTORY"
	KindRateLimited           Kind = "RATE_LIMITED"
	KindInternal              Kind = "INTERNAL"
)

// Error is a typed application error with a stable kind and public-safe message.
type Error struct {
	Kind    Kind   // stable category for handler mapping
	Message string // public-safe description (safe to include in API response)
	Cause   error  // optional wrapped error (for logging / root cause only)
}

func (e *Error) Error() string { return e.Message }

// Unwrap returns the optional cause. Implements the errors.Unwrap interface.
func (e *Error) Unwrap() error { return e.Cause }

// --- constructors ---

func Validation(msg string) *Error {
	return &Error{Kind: KindValidation, Message: msg}
}

func Validationf(format string, args ...any) *Error {
	return &Error{Kind: KindValidation, Message: fmt.Sprintf(format, args...)}
}

func BadRequest(msg string) *Error {
	return &Error{Kind: KindBadRequest, Message: msg}
}

func BadRequestf(format string, args ...any) *Error {
	return &Error{Kind: KindBadRequest, Message: fmt.Sprintf(format, args...)}
}

func Unauthorized(msg string) *Error {
	return &Error{Kind: KindUnauthorized, Message: msg}
}

func Forbidden(msg string) *Error {
	return &Error{Kind: KindForbidden, Message: msg}
}

func NotFound(msg string) *Error {
	return &Error{Kind: KindNotFound, Message: msg}
}

func NotFoundf(format string, args ...any) *Error {
	return &Error{Kind: KindNotFound, Message: fmt.Sprintf(format, args...)}
}

func Conflict(msg string) *Error {
	return &Error{Kind: KindConflict, Message: msg}
}

func InsufficientFunds(msg string) *Error {
	return &Error{Kind: KindInsufficientFunds, Message: msg}
}

func InsufficientInventory(msg string) *Error {
	return &Error{Kind: KindInsufficientInventory, Message: msg}
}

func RateLimited(msg string) *Error {
	return &Error{Kind: KindRateLimited, Message: msg}
}

func Internal(msg string) *Error {
	return &Error{Kind: KindInternal, Message: msg}
}

func Internalf(format string, args ...any) *Error {
	return &Error{Kind: KindInternal, Message: fmt.Sprintf(format, args...)}
}

// Wrap creates an Error with the given kind and a wrapped error suitable for
// logging. The public message is derived from the kind; the cause is retained
// via Cause for inspection and logging.
func Wrap(kind Kind, cause error) *Error {
	if cause == nil {
		return &Error{Kind: kind, Message: string(kind)}
	}
	return &Error{Kind: kind, Message: cause.Error(), Cause: cause}
}

// WrapMsg creates an Error with an explicit public-safe message and a wrapped cause.
func WrapMsg(kind Kind, message string, cause error) *Error {
	return &Error{Kind: kind, Message: message, Cause: cause}
}

// --- matching ---

// HasKind reports whether err is an *Error with the given kind.
func HasKind(err error, kind Kind) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Kind == kind
	}
	return false
}

// KindOf returns the Kind of err if it is an *Error, or KindInternal for
// unrecognized errors.
func KindOf(err error) Kind {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Kind
	}
	return KindInternal
}

// IsUserError reports whether err is a typed apperr.Error as opposed to
// an unexpected system error.
func IsUserError(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr)
}
