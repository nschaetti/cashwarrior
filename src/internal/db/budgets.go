package db

import (
	"database/sql"
	"time"
)

type Budget struct {
	ID          int64
	CategoryID  int64
	Name        string
	Description string
	Amount      float64
	Currency    string
	Period      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BudgetListFilter struct {
	CategoryID *int64
	NameLike   string
	Limit      int
	Offset     int
}

type CreateBudgetInput struct {
	CategoryID  int64
	Name        string
	Description string
	Amount      float64
	Currency    string
	Period      string
}

func GetBudgetByID(db *sql.DB, id int64) (Budget, error) {
	var budget Budget
	err := db.QueryRow(`
SELECT id, category_id, name, description, amount, currency, period, created_at, updated_at
FROM budgets
WHERE id = ?
`, id).Scan(
		&budget.ID,
		&budget.CategoryID,
		&budget.Name,
		&budget.Description,
		&budget.Amount,
		&budget.Currency,
		&budget.Period,
		&budget.CreatedAt,
		&budget.UpdatedAt,
	)
	return budget, err
}

func ListBudgets(db *sql.DB, filter BudgetListFilter) ([]Budget, error) {
	query := `
SELECT id, category_id, name, description, amount, currency, period, created_at, updated_at
FROM budgets
WHERE 1 = 1
`
	args := make([]interface{}, 0)

	if filter.CategoryID != nil {
		query += "AND category_id = ?\n"
		args = append(args, *filter.CategoryID)
	}
	if filter.NameLike != "" {
		query += "AND name LIKE ?\n"
		args = append(args, "%"+filter.NameLike+"%")
	}

	query += "ORDER BY id\n"
	if filter.Limit > 0 {
		query += "LIMIT ?\n"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += "OFFSET ?\n"
		args = append(args, filter.Offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	budgets := make([]Budget, 0)
	for rows.Next() {
		var budget Budget
		err = rows.Scan(
			&budget.ID,
			&budget.CategoryID,
			&budget.Name,
			&budget.Description,
			&budget.Amount,
			&budget.Currency,
			&budget.Period,
			&budget.CreatedAt,
			&budget.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return budgets, nil
}

func BudgetExists(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM budgets WHERE name = ?
)
`, name).Scan(&exists)
	return exists, err
}

func InsertBudget(db *sql.DB, input CreateBudgetInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO budgets (category_id, name, description, amount, currency, period)
VALUES (?, ?, ?, ?, ?, ?)
`, input.CategoryID, input.Name, input.Description, input.Amount, input.Currency, input.Period)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
