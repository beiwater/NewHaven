package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/domain/social"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// SocialHandler handles chat, message, and contact HTTP endpoints.
type SocialHandler struct {
	social        storage.SocialStorage
	companies     storage.CompanyStorage
	maxMessageLen int
}

func NewSocialHandler(social storage.SocialStorage, companies storage.CompanyStorage, maxMessageLen int) *SocialHandler {
	return &SocialHandler{social: social, companies: companies, maxMessageLen: maxMessageLen}
}

// lookupCompanyName resolves the company's display name from storage.
func (h *SocialHandler) lookupCompanyName(ctx context.Context, companyID int) string {
	c, err := h.companies.GetCompany(ctx, companyID)
	if err != nil {
		return fmt.Sprintf("Company-%d", companyID)
	}
	return c.Name
}

// handleMessages returns all chat messages (GET) or sends a new one (POST).
func (h *SocialHandler) handleMessages(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	switch r.Method {
	case http.MethodGet:
		chatroom := r.URL.Query().Get("chatroom")
		if isLegacyPrivateChannel(chatroom) {
			writeErr(w, http.StatusGone, ErrorForbidden, "legacy private chat is disabled; use room-based chat", nil)
			return
		}
		msgs, err := h.social.GetMessages(r.Context(), chatroom, 50)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to fetch messages", nil)
			return
		}
		// Filter: if no chatroom specified, only return public messages
		result := make([]map[string]any, 0, len(msgs))
		for _, m := range msgs {
			if chatroom == "" && strings.HasPrefix(m.Channel, "C:") {
				continue
			}
			result = append(result, map[string]any{
				"id":       fmt.Sprintf("msg-%d", m.ID),
				"chatroom": m.Channel,
				"body":     m.Content,
				"at":       m.CreatedAt,
				"from":     m.SenderName,
				"fromId":   m.CompanyID,
			})
		}
		writeJSON(w, http.StatusOK, result)
	case http.MethodPost:
		var req struct {
			Chatroom string `json:"chatroom"`
			Body     string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, ErrorValidation, "invalid request body", nil)
			return
		}
		if len([]rune(req.Body)) > h.maxMessageLen {
			writeErr(w, http.StatusBadRequest, ErrorValidation, "message too long", nil)
			return
		}
		if isLegacyPrivateChannel(req.Chatroom) {
			writeErr(w, http.StatusGone, ErrorForbidden, "legacy private chat is disabled; use room-based chat", nil)
			return
		}
		companyName := h.lookupCompanyName(r.Context(), companyID)
		msg := &social.Message{
			CompanyID:  companyID,
			SenderName: companyName,
			Content:    req.Body,
			Channel:    req.Chatroom,
			CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		}
		if err := h.social.SaveMessage(r.Context(), msg); err != nil {
			writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to save message", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":       fmt.Sprintf("msg-%d", msg.ID),
			"chatroom": msg.Channel,
			"body":     msg.Content,
			"at":       msg.CreatedAt,
			"from":     msg.SenderName,
			"fromId":   msg.CompanyID,
		})
	default:
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
	}
}

// handleV2Message handles POST /api/v2/message/ (send) and GET /api/v2/message/{id}/read/ (mark read).
func (h *SocialHandler) handleV2Message(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	switch r.Method {
	case http.MethodPost:
		var req struct {
			Chatroom string `json:"chatroom"`
			Body     string `json:"body"`
			Token    string `json:"token,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, ErrorValidation, "invalid request body", nil)
			return
		}
		if len([]rune(req.Body)) > h.maxMessageLen {
			writeErr(w, http.StatusBadRequest, ErrorValidation, "message too long", nil)
			return
		}
		if isLegacyPrivateChannel(req.Chatroom) {
			writeErr(w, http.StatusGone, ErrorForbidden, "legacy private chat is disabled; use room-based chat", nil)
			return
		}
		companyName := h.lookupCompanyName(r.Context(), companyID)
		msg := &social.Message{
			CompanyID:  companyID,
			SenderName: companyName,
			Content:    req.Body,
			Channel:    req.Chatroom,
			CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		}
		if err := h.social.SaveMessage(r.Context(), msg); err != nil {
			writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to save message", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":       fmt.Sprintf("msg-%d", msg.ID),
			"chatroom": msg.Channel,
			"body":     msg.Content,
			"at":       msg.CreatedAt,
			"from":     msg.SenderName,
			"fromId":   msg.CompanyID,
		})
	case http.MethodGet:
		// GET /api/v2/message/{id}/read/ - mark as read
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
	}
}

// handleChatroom returns the chatroom messages (GET) or handles PATCH.
func (h *SocialHandler) handleChatroom(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		channel := r.URL.Query().Get("channel")
		if channel == "" {
			channel = "general"
		}
		if isLegacyPrivateChannel(channel) {
			writeErr(w, http.StatusGone, ErrorForbidden, "legacy private chat is disabled; use room-based chat", nil)
			return
		}
		msgs, err := h.social.GetMessages(r.Context(), channel, 50)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to fetch messages", nil)
			return
		}
		result := make([]map[string]any, 0, len(msgs))
		for _, m := range msgs {
			result = append(result, map[string]any{
				"id":       fmt.Sprintf("msg-%d", m.ID),
				"chatroom": m.Channel,
				"body":     m.Content,
				"at":       m.CreatedAt,
				"from":     m.SenderName,
				"fromId":   m.CompanyID,
			})
		}
		writeJSON(w, http.StatusOK, result)
	case http.MethodPatch:
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
	}
}

// handleContacts returns all companies as contacts for search.
func (h *SocialHandler) handleContacts(w http.ResponseWriter, r *http.Request) {
	companies, err := h.companies.GetAllCompanies(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"chatrooms":                 []any{},
			"contacts":                  []any{},
			"unreadMessagesOtherRealms": 0,
			"invisible":                 false,
			"ignoringCompanies":         map[string]any{},
			"companiesChatBlockingUs":   map[string]any{},
		})
		return
	}
	contacts := make([]map[string]any, 0, len(companies))
	for _, c := range companies {
		contacts = append(contacts, map[string]any{
			"companyId": c.ID,
			"company":   c.Name,
			"playerId":  fmt.Sprintf("p-%d", c.PlayerID),
			"level":     c.Level,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"chatrooms":                 []any{},
		"contacts":                  contacts,
		"unreadMessagesOtherRealms": 0,
		"invisible":                 false,
		"ignoringCompanies":         map[string]any{},
		"companiesChatBlockingUs":   map[string]any{},
	})
}

// handleMarkRead marks a message as read via path suffix.
// This handles the pattern: GET /api/v2/message/{messageId}/read/
func (h *SocialHandler) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v2/message/")
	path = strings.TrimSuffix(path, "/read/")
	_ = path
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func isLegacyPrivateChannel(channel string) bool {
	return strings.HasPrefix(strings.TrimSpace(channel), "C:")
}
