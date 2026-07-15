package finance

// LedgerEntry records a financial transaction. Append-only.
type LedgerEntry struct {
	ID           int64          `json:"id"`
	CompanyID    int            `json:"company_id"`
	Kind         string         `json:"kind"`
	Amount       float64        `json:"amount"`
	Direction    string         `json:"direction"` // "in" or "out"
	BalanceAfter float64        `json:"balance_after"`
	Metadata     map[string]any `json:"metadata,omitempty"` // content depends on transaction kind; too varied across kinds for a fixed struct, safe to remain map[string]any
	CreatedAt    string         `json:"created_at"`
}

// Bond represents a bond issued by a company.
type Bond struct {
	ID              string  `json:"id"`
	IssuerCompanyID int     `json:"issuer_company_id"`
	FaceValue       float64 `json:"face_value"`
	InterestRate    float64 `json:"interest_rate"`
	TotalQuantity   int     `json:"total_quantity"`
	IssuedQuantity  int     `json:"issued_quantity"`
	Status          string  `json:"status"` // "active", "called", "defaulted"
	LastSettledAt   string  `json:"last_settled_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// BondHolding represents a company's holding of a specific bond.
type BondHolding struct {
	BondID      string `json:"bond_id"`
	CompanyID   int    `json:"company_id"`
	Quantity    int    `json:"quantity"`
	PurchasedAt string `json:"purchased_at"`
}

// FinancialStatement represents income statement or balance sheet.
type FinancialStatement struct {
	PeriodStart string             `json:"period_start"`
	PeriodEnd   string             `json:"period_end"`
	Items       map[string]float64 `json:"items"`
}

// IssueBondRequest is the DTO for issuing a new bond.
type IssueBondRequest struct {
	Amount      int     `json:"amount" validate:"required,min=1"`
	InterestPct float64 `json:"interest_pct" validate:"required,min=0.5,max=2.0"`
}

// BuyBondRequest is the DTO for buying bonds.
type BuyBondRequest struct {
	BondID string `json:"bond_id" validate:"required"`
	Amount int    `json:"amount" validate:"required,min=1"`
}
