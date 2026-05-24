package db

import (
	"database/sql"
	"time"
)

type Category struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CategoryListFilter struct {
	NameLike string
	Limit    int
	Offset   int
}

type CreateCategoryInput struct {
	Name string
}

func GetCategoryByID(db *sql.DB, id int64) (Category, error) {
	var category Category
	err := db.QueryRow(`
SELECT id, name, created_at, updated_at
FROM categories
WHERE id = ?
`, id).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	return category, err
}

func ListCategories(db *sql.DB, filter CategoryListFilter) ([]Category, error) {
	query := `
SELECT id, name, created_at, updated_at
FROM categories
`
	args := make([]interface{}, 0)

	if filter.NameLike != "" {
		query += "WHERE name LIKE ?\n"
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

	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		err = rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func CategoryExists(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM categories WHERE name = ?
)
`, name).Scan(&exists)
	return exists, err
}

func InsertCategory(db *sql.DB, input CreateCategoryInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO categories (name)
VALUES (?)
`, input.Name)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
