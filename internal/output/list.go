package output

import "time"

// ListTransaction is the stable representation of a transaction in list output.
type ListTransaction struct {
	ID          int64     `json:"id"`
	Identifier  string    `json:"identifier"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Account     string    `json:"account"`
	Vendor      string    `json:"vendor"`
	Category    string    `json:"category"`
	Group       string    `json:"group"`
}

type ListCurrencySummary struct {
	Currency string  `json:"currency"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Net      float64 `json:"net"`
}

type ListAccountSummary struct {
	Account  string  `json:"account"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Net      float64 `json:"net"`
}

type ListSummary struct {
	ByCurrency []ListCurrencySummary `json:"by_currency"`
	ByAccount  []ListAccountSummary  `json:"by_account"`
}

// ListData contains all data needed by table and JSON list renderers.
type ListData struct {
	Transactions []ListTransaction `json:"transactions"`
	Summary      ListSummary       `json:"summary"`
}
