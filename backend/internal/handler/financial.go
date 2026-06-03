package handler

import (
	"math"
	"net/http"
	"time"
)

func (h *Handler) RegisterFinancial(mux *http.ServeMux) {
	mux.HandleFunc("/api/v2/companies/me/income-statement/", h.withAuth(h.handleIncomeStatement))
	mux.HandleFunc("/api/v2/companies/me/balance-sheet/", h.withAuth(h.handleBalanceSheet))
	mux.HandleFunc("/api/v2/companies/me/cashflow-statement/", h.withAuth(h.handleCashflowStatement))
	mux.HandleFunc("/api/v2/companies/me/cashflow/recent/", h.withAuth(h.handleRecentCashflow))
	mux.HandleFunc("/api/v2/companies/me/past-finances-overview/", h.withAuth(h.handlePastFinancesOverview))
	mux.HandleFunc("/api/v3/companies/me/past-finances/", h.withAuth(h.handlePastFinances))
}

func (h *Handler) handleIncomeStatement(w http.ResponseWriter, r *http.Request) {
	fs := h.svc.FinancialStatements(h.companyID(r))
	writeJSON(w, 200, fs["incomeStatement"])
}

func (h *Handler) handleBalanceSheet(w http.ResponseWriter, r *http.Request) {
	fs := h.svc.FinancialStatements(h.companyID(r))
	writeJSON(w, 200, fs["balanceSheet"])
}

func (h *Handler) handleCashflowStatement(w http.ResponseWriter, r *http.Request) {
	fs := h.svc.FinancialStatements(h.companyID(r))
	writeJSON(w, 200, fs["cashflowStatement"])
}

func (h *Handler) handleRecentCashflow(w http.ResponseWriter, r *http.Request) {
	st := h.svc.Snapshot()
	data := []map[string]any{}
	for i, e := range st.Ledger {
		if i >= 100 {
			break
		}
		delta := math.Abs(e.Amount)
		if e.Direction == "out" {
			delta = -delta
		}
		data = append(data, map[string]any{"kind": e.Kind, "moneyDelta": delta, "at": e.At})
	}
	oldest := time.Now().UTC().Format(time.RFC3339)
	if len(st.Ledger) > 0 {
		oldest = st.Ledger[len(st.Ledger)-1].At
	}
	writeJSON(w, 200, map[string]any{"data": data, "oldestPulled": oldest, "money": st.GetCompany(h.companyID(r)).Money})
}

func (h *Handler) handlePastFinancesOverview(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"weeklyNet": []float64{2100, 3800, 2750, 4200}})
}

func (h *Handler) handlePastFinances(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"series": []map[string]any{
		{"date": "2026-05-28", "net": 890.2},
		{"date": "2026-05-29", "net": 1022.4},
	}})
}
