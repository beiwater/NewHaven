package httpapi
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/newhaven/backend-next/internal/domain/chat"
	"github.com/newhaven/backend-next/internal/storage"
)

// ChatHandler handles private room-based chat endpoints.
type ChatHandler struct {
	chat      storage.ChatStorage
	companies storage.CompanyStorage
	maxMsgLen int
}

func NewChatHandler(chat storage.ChatStorage, companies storage.CompanyStorage, maxMsgLen int) *ChatHandler {
	return &ChatHandler{chat: chat, companies: companies, maxMsgLen: maxMsgLen}
}

// lookupCompanyName resolves the company's display name from storage.
func (h *ChatHandler) lookupCompanyName(ctx context.Context, companyID int) string {
	c, err := h.companies.GetCompany(ctx, companyID)
	if err != nil {
		return fmt.Sprintf("Company-%d", companyID)
	}
	return c.Name
}

// handleCreateRoom creates or gets a room with another company.
func (h *ChatHandler) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "not authenticated", nil)
		return
	}

	var req struct {
		CompanyId int `json:"companyId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "invalid request", nil)
		return
	}
	if req.CompanyId <= 0 || req.CompanyId == companyID {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "invalid company", nil)
		return
	}

	room, err := h.chat.GetOrCreateRoom(r.Context(), companyID, req.CompanyId)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to create room", nil)
		return
	}

	msgs, _ := h.chat.GetRoomMessages(r.Context(), room.ID, 50)

	writeJSON(w, http.StatusOK, map[string]any{
		"room":     room,
		"messages": msgs,
	})
}

// handleListRooms lists the user's chat rooms.
func (h *ChatHandler) handleListRooms(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "not authenticated", nil)
		return
	}

	rooms, err := h.chat.GetUserRooms(r.Context(), companyID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to list rooms", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rooms": rooms,
	})
}

// handleGetRoomMessages returns messages for a room (with participant check).
func (h *ChatHandler) handleGetRoomMessages(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "not authenticated", nil)
		return
	}

	roomID := extractRoomID(r.URL.Path)
	if !h.isParticipant(roomID, companyID) {
		writeErr(w, http.StatusForbidden, ErrorForbidden, "not a participant", nil)
		return
	}

	msgs, err := h.chat.GetRoomMessages(r.Context(), roomID, 50)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to get messages", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"messages": msgs,
	})
}

// handleSendMessage sends a message to a room (with participant check).
func (h *ChatHandler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "not authenticated", nil)
		return
	}

	roomID := extractRoomID(r.URL.Path)
	if !h.isParticipant(roomID, companyID) {
		writeErr(w, http.StatusForbidden, ErrorForbidden, "not a participant", nil)
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "invalid request", nil)
		return
	}
	if len([]rune(req.Body)) > h.maxMsgLen {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "message too long", nil)
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "message cannot be empty", nil)
		return
	}

	companyName := h.lookupCompanyName(r.Context(), companyID)

	msg := &chat.Message{
		RoomID:     roomID,
		SenderID:   companyID,
		SenderName: companyName,
		Content:    req.Body,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	if err := h.chat.SaveRoomMessage(r.Context(), msg); err != nil {
		writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to save message", nil)
		return
	}

	writeJSON(w, http.StatusOK, msg)
}

// isParticipant checks if the given company ID is a participant in the room.
func (h *ChatHandler) isParticipant(roomID string, companyID int) bool {
	parts := strings.SplitN(strings.TrimPrefix(roomID, "p:"), "-", 2)
	if len(parts) != 2 {
		return false
	}
	a, _ := strconv.Atoi(parts[0])
	b, _ := strconv.Atoi(parts[1])
	return companyID == a || companyID == b
}

// extractRoomID extracts the room ID from a URL path like
// /api/v2/chat/room/p:1000001-1000002/messages/.
func extractRoomID(path string) string {
	parts := strings.Split(path, "/")
	for _, p := range parts {
		if strings.HasPrefix(p, "p:") {
			return p
		}
	}
	return ""
}
