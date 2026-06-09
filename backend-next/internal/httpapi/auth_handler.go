package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/newhaven/backend-next/internal/app/auth"
	"github.com/newhaven/backend-next/internal/apperr"
	domain "github.com/newhaven/backend-next/internal/domain/auth"
)

type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/register", h.handleRegister)
	mux.HandleFunc("/api/login", h.handleLogin)
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "METHOD_NOT_ALLOWED", "only POST is accepted", nil)
		return
	}

	var req domain.RegisterRequest
	if err := decodeAuthRequest(w, r, &req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	resp, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrUsernameTaken) {
			writeErr(w, 400, ErrorConflict, "username already taken", nil)
			return
		}
		if apperr.HasKind(err, apperr.KindValidation) {
			writeAppErr(w, err)
			return
		}
		writeErr(w, 500, ErrorInternal, "registration failed", nil)
		return
	}

	writeSuccess(w, 200, resp)
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "METHOD_NOT_ALLOWED", "only POST is accepted", nil)
		return
	}

	var req domain.LoginRequest
	if err := decodeAuthRequest(w, r, &req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}
	if req.Username == "" || req.Password == "" {
		writeErr(w, 400, ErrorValidation, "username and password are required", nil)
		return
	}

	resp, err := h.svc.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeErr(w, 401, ErrorUnauthorized, "invalid username or password", nil)
			return
		}
		writeErr(w, 500, ErrorInternal, "login failed", nil)
		return
	}

	writeSuccess(w, 200, resp)
}

func decodeAuthRequest(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}
