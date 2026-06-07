package httpapi
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"strconv"
	
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
// handleGetRoomMessages returns messages for a room (with participant check and read receipts).
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

	// Get read status of the OTHER participant
	otherID, found := h.findOtherParticipant(roomID, companyID)
	var readUpTo int64
	if found {
		readUpTo = h.chat.GetRoomReadStatus(r.Context(), roomID, otherID)
	}

	result := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, map[string]any{
			"id":          m.ID,
			"room_id":     m.RoomID,
			"sender_id":   m.SenderID,
			"sender_name": m.SenderName,
			"content":     m.Content,
			"created_at":  m.CreatedAt,
			"read":        m.ID <= readUpTo,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"messages": result})
}

// handleSendMessage sends a message to a room (with participant check).
func (h *ChatHandler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "not authenticated", nil)
		return
	}

	var req struct {
		RoomId string `json:"roomId"`
		Body   string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "invalid request", nil)
		return
	}
	if req.RoomId == "" || !h.isParticipant(req.RoomId, companyID) {
		writeErr(w, http.StatusForbidden, ErrorForbidden, "not a participant", nil)
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
		RoomID:     req.RoomId,
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

// handleMarkRead marks messages as read via body (roomId in body to avoid chi colon issues).
func (h *ChatHandler) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "not authenticated", nil)
		return
	}

	var req struct {
		RoomId        string `json:"roomId"`
		LastMessageId int64  `json:"lastMessageId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "invalid request", nil)
		return
	}
	if req.RoomId == "" || !h.isParticipant(req.RoomId, companyID) {
		writeErr(w, http.StatusForbidden, ErrorForbidden, "not a participant", nil)
		return
	}

	if err := h.chat.MarkRoomRead(r.Context(), req.RoomId, int64(companyID), req.LastMessageId); err != nil {
		writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to mark read", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *ChatHandler) getParticipants(roomID string) (int, int, bool) {
	parts := strings.SplitN(strings.TrimPrefix(roomID, "p:"), "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	a, err1 := strconv.Atoi(parts[0])
	b, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return a, b, true
}


func (h *ChatHandler) isParticipant(roomID string, companyID int) bool {
	a, b, ok := h.getParticipants(roomID)
	return ok && (companyID == a || companyID == b)
}

func (h *ChatHandler) findOtherParticipant(roomID string, companyID int) (int, bool) {
	a, b, ok := h.getParticipants(roomID)
	if !ok {
		return 0, false
	}
	if a == companyID {
		return b, true
	}
	if b == companyID {
		return a, true
	}
	return 0, false
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
