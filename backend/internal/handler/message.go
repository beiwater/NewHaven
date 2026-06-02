package handler

import (
	"encoding/json"
	"fmt"
	"go-sim-api/internal/model"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) RegisterMessage(mux *http.ServeMux) {
	mux.HandleFunc("/api/messages/", h.withAuth(h.handleMessages))
	mux.HandleFunc("/api/messages_by_company/", h.withAuth(h.handleMessagesByCompany))
	mux.HandleFunc("/api/v2/message/", h.withAuth(h.handleV2Message))
	mux.HandleFunc("/api/v2/chatroom/", h.withAuth(h.handleChatroom))
	mux.HandleFunc("/api/v2/contacts/", h.withAuth(h.handleContacts))
	mux.HandleFunc("/api/v2/newspaper/articles-by-author/", h.withAuth(h.handleNewspaperArticles))
	mux.HandleFunc("/api/v2/newspaper/articles/", h.withAuth(h.handleNewspaperArticleList))
	mux.HandleFunc("/api/v2/newspaper/publishing-costs/", h.withAuth(h.handleNewspaperPublishingCosts))
}

func (h *Handler) handleMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, h.svc.Snapshot().Messages)
	case http.MethodPatch:
		writeJSON(w, 200, map[string]any{"ok": true})
	default:
		writeErr(w, 405, "method not allowed")
	}
}

func (h *Handler) handleMessagesByCompany(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().Notifications)
}

func (h *Handler) handleV2Message(w http.ResponseWriter, r *http.Request) {
	// POST /api/v2/message/ -> send message
	// GET  /api/v2/message/{id}/read/ -> mark as read
	switch r.Method {
	case http.MethodPost:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, "invalid json")
			return
		}
		msg := model.Message{
			ID:       fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Chatroom: fmt.Sprint(body["chatroom"]),
			Body:     fmt.Sprint(body["body"]),
			Token:    fmt.Sprint(body["token"]),
			At:       time.Now().UTC().Format(time.RFC3339),
		}
		writeJSON(w, 200, h.svc.AddMessage(msg))
	case http.MethodGet:
		writeJSON(w, 200, map[string]any{"ok": true})
	default:
		writeErr(w, 405, "method not allowed")
	}
}

func (h *Handler) handleChatroom(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		writeJSON(w, 200, map[string]any{"ok": true})
		return
	}
	writeJSON(w, 200, h.svc.Snapshot().Messages)
}

func (h *Handler) handleContacts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{
		"chatrooms":                 []any{},
		"contacts":                  h.svc.CompaniesByPlayer("network"),
		"unreadMessagesOtherRealms": 0,
		"invisible":                 false,
		"ignoringCompanies":         map[string]any{},
		"companiesChatBlockingUs":   map[string]any{},
	})
}

func (h *Handler) handleNewspaperArticles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []map[string]any{
		{"id": "news-1", "title": "Commodity Spread Narrows", "publishedAt": time.Now().Add(-6 * time.Hour).UTC().Format(time.RFC3339)},
	})
}

func (h *Handler) handleNewspaperArticleList(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v2/newspaper/articles/")
	path = strings.Trim(path, "/")
	if path != "" {
		// GET /api/v2/newspaper/articles/{id}/
		article := h.svc.SampleArticles()
		for _, a := range article {
			if a["id"] == path {
				writeJSON(w, 200, a)
				return
			}
		}
		writeErr(w, 404, "article not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		articles := h.svc.SampleArticles()
		writeJSON(w, 200, map[string]any{"articles": articles, "total": len(articles)})
	case http.MethodPost:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, "invalid json")
			return
		}
		article := map[string]any{
			"id":          fmt.Sprintf("article-%d", time.Now().UnixNano()),
			"title":       body["title"],
			"body":        body["body"],
			"author":      body["author"],
			"publishedAt": time.Now().UTC().Format(time.RFC3339),
			"readCount":   0,
		}
		writeJSON(w, 200, map[string]any{"ok": true, "article": article})
	default:
		writeErr(w, 405, "method not allowed")
	}
}

func (h *Handler) handleNewspaperPublishingCosts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{
		"costs": map[string]any{
			"basic":   map[string]any{"simCash": 10, "money": 1000},
			"premium": map[string]any{"simCash": 50, "money": 5000},
		},
	})
}
