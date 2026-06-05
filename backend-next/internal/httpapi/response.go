package httpapi

import (
	"encoding/json"
	"net/http"
)

// --- Standard response envelope ---

// APIResponse is the standard success response.
type APIResponse struct {
	Data  any          `json:"data"`
	Error *APIError    `json:"error"`
	Meta  *APIMeta     `json:"meta,omitempty"`
}

// APIError is the standard error payload.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// APIMeta holds response metadata.
type APIMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

// --- Error codes ---

const (
	ErrorUnauthorized      = "UNAUTHORIZED"
	ErrorForbidden         = "FORBIDDEN"
	ErrorNotFound          = "NOT_FOUND"
	ErrorBadRequest        = "BAD_REQUEST"
	ErrorValidation        = "VALIDATION_ERROR"
	ErrorInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrorInsufficientInv   = "INSUFFICIENT_INVENTORY"
	ErrorConflict          = "CONFLICT"
	ErrorRateLimited       = "RATE_LIMITED"
	ErrorInternal          = "INTERNAL_ERROR"
)

// --- Response writers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeSuccess sends a standard success response.
func writeSuccess(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, APIResponse{
		Data:  data,
		Error: nil,
	})
}

// writeErr sends a standard error response.
func writeErr(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, APIResponse{
		Data: nil,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// writeValidationErr sends a 400 validation error with field-level details.
func writeValidationErr(w http.ResponseWriter, message string, fields any) {
	writeErr(w, 400, ErrorValidation, message, fields)
}

// --- Health handlers ---

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, 200, map[string]string{"status": "ok"})
}

func handleReadyz(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, 200, map[string]string{"status": "ready"})
}
