package output

import "time"

type ShowTransaction struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	Amount      float64       `json:"amount"`
	Currency    string        `json:"currency"`
	Account     string        `json:"account"`
	Place       string        `json:"place"`
	Description string        `json:"description"`
	Date        time.Time     `json:"date"`
	Category    string        `json:"category"`
	Group       string        `json:"group"`
	Tags        []string      `json:"tags"`
	Deleted     bool          `json:"deleted"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Transfer    *ShowTransfer `json:"transfer,omitempty"`
}

type ShowTransfer struct {
	Amount      float64 `json:"amount"`
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	PairID      string  `json:"pair_id"`
}

type SummaryDay struct {
	Date         string               `json:"date"`
	Transactions int                  `json:"transactions"`
	Currencies   []SummaryDayCurrency `json:"currencies"`
}

type SummaryDayCurrency struct {
	Currency string  `json:"currency"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Net      float64 `json:"net"`
}

type SummaryDaysData struct {
	Days []SummaryDay `json:"days"`
}

type AccountListItem struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Currency       string  `json:"currency"`
	InitialBalance float64 `json:"initial_balance"`
	Balance        float64 `json:"balance"`
	Operations     int     `json:"operations"`
	MonthIncome    float64 `json:"month_income"`
	MonthExpenses  float64 `json:"month_expenses"`
	MonthNet       float64 `json:"month_net"`
	MonthTransfers float64 `json:"month_transfers"`
}

type AccountBalanceItem struct {
	ID          int64     `json:"id"`
	Identifier  string    `json:"identifier"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Vendor      string    `json:"vendor"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Category    string    `json:"category"`
	Balance     float64   `json:"balance"`
}

type AccountBalanceData struct {
	Account        string               `json:"account"`
	Currency       string               `json:"currency"`
	InitialBalance float64              `json:"initial_balance"`
	Transactions   []AccountBalanceItem `json:"transactions"`
}

type CategoriesData struct {
	Categories []CategoryListItem `json:"categories"`
}
type CategoryListItem struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	ParentID     *int64  `json:"parent_id"`
	Depth        int     `json:"depth"`
	Transactions int     `json:"transactions"`
	Expenses     float64 `json:"expenses"`
	Incomes      float64 `json:"incomes"`
	Net          float64 `json:"net"`
}

type GroupsData struct {
	Groups []GroupListItem `json:"groups"`
}
type GroupListItem struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Transactions int        `json:"transactions"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Sum          float64    `json:"sum"`
}

type PlacesData struct {
	Places []PlaceListItem `json:"places"`
}
type PlaceListItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TagsData struct {
	Tags []TagListItem `json:"tags"`
}
type TagListItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Transactions int    `json:"transactions"`
}
