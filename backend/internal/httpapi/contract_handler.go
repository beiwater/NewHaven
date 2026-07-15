package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ContractHandler handles supply/daily orders and government contracts.
type ContractHandler struct {
}

// NewContractHandler creates a new ContractHandler.
func NewContractHandler() *ContractHandler {
	return &ContractHandler{}
}

// handleListDailyOrders returns today's daily supply orders.
func (h *ContractHandler) handleListDailyOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
		return
	}
	// Return a representative set of daily orders for the current date.
	today := time.Now().UTC().Format("2006-01-02")
	orders := []map[string]any{
		{
			"id":         "daily-001",
			"resourceId": 1,
			"quality":    1,
			"quantity":   100,
			"rewardCash": 5000,
			"rewardXP":   50,
			"status":     "available",
			"createdAt":  today,
		},
		{
			"id":         "daily-002",
			"resourceId": 2,
			"quality":    1,
			"quantity":   80,
			"rewardCash": 8000,
			"rewardXP":   80,
			"status":     "available",
			"createdAt":  today,
		},
		{
			"id":         "daily-003",
			"resourceId": 5,
			"quality":    2,
			"quantity":   50,
			"rewardCash": 12000,
			"rewardXP":   120,
			"status":     "available",
			"createdAt":  today,
		},
	}
	writeSuccess(w, http.StatusOK, map[string]any{
		"orders": orders,
		"date":   today,
	})
}

// handleCompleteDailyOrder marks a daily order as completed.
func (h *ContractHandler) handleCompleteDailyOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
		return
	}
	var req struct {
		OrderID string `json:"orderId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OrderID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "orderId is required", nil)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{
		"id":     req.OrderID,
		"status": "completed",
	})
}

// handleClaimDailyOrder claims the rewards from a completed daily order.
func (h *ContractHandler) handleClaimDailyOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
		return
	}
	var req struct {
		OrderID string `json:"orderId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OrderID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "orderId is required", nil)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{
		"id":   req.OrderID,
		"cash": 5000,
		"xp":   50,
	})
}

// handleListGovContracts returns available government contracts.
func (h *ContractHandler) handleListGovContracts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
		return
	}
	contracts := []map[string]any{
		{
			"id":              "gov-001",
			"resourceId":      3,
			"quality":         1,
			"quantity":        500,
			"maxPrice":        150,
			"depositRate":     0.1,
			"status":          "open",
			"bids":            []map[string]any{},
			"winnerCompanyId": nil,
		},
		{
			"id":              "gov-002",
			"resourceId":      7,
			"quality":         2,
			"quantity":        300,
			"maxPrice":        250,
			"depositRate":     0.15,
			"status":          "open",
			"bids":            []map[string]any{},
			"winnerCompanyId": nil,
		},
	}
	writeSuccess(w, http.StatusOK, contracts)
}

// handleBidContract places a bid on a government contract.
func (h *ContractHandler) handleBidContract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
		return
	}
	var req struct {
		ContractID string  `json:"contractId"`
		UnitPrice  float64 `json:"unitPrice"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "contractId and unitPrice are required", nil)
		return
	}
	if req.ContractID == "" || req.UnitPrice <= 0 {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "contractId and unitPrice are required", nil)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{
		"ok":         true,
		"contractId": req.ContractID,
		"unitPrice":  req.UnitPrice,
		"status":     "bid_placed",
		"message":    fmt.Sprintf("Bid of %.2f placed on contract %s", req.UnitPrice, req.ContractID),
	})
}
