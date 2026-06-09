package httpapi

import (
	"net/http"

	"github.com/newhaven/backend-next/internal/app/finance"
)

// FinanceHandler handles finance-related HTTP endpoints.
type FinanceHandler struct {
	svc *finance.Service
}

// NewFinanceHandler creates a new FinanceHandler.
func NewFinanceHandler(svc *finance.Service) *FinanceHandler {
	return &FinanceHandler{svc: svc}
}

// handleRecentCashflow returns recent cashflow entries.
func (h *FinanceHandler) handleRecentCashflow(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.GetRecentCashflow(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleIncomeStatement returns the income statement.
func (h *FinanceHandler) handleIncomeStatement(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.GetIncomeStatement(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleBalanceSheet returns the balance sheet.
func (h *FinanceHandler) handleBalanceSheet(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.GetBalanceSheet(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleCashflowStatement returns the cashflow statement.
func (h *FinanceHandler) handleCashflowStatement(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.GetCashflowStatement(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handlePastFinances returns the past finances series.
func (h *FinanceHandler) handlePastFinances(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.GetPastFinances(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}
