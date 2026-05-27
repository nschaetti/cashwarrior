package db

import (
	"database/sql"
	"time"
)

type Category struct {
	ID        int64
	Name      string
	ParentID  *int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CategoryListFilter struct {
	NameLike string
	Limit    int
	Offset   int
}

type CreateCategoryInput struct {
	Name     string
	ParentID *int64
}

const rootCategoryName = "root"

func GetCategoryByID(db DBTX, id int64) (Category, error) {
	var category Category
	var parentID sql.NullInt64
	err := db.QueryRow(`
SELECT id, name, parent_id, created_at, updated_at
FROM categories
WHERE id = ?
`, id).Scan(
		&category.ID,
		&category.Name,
		&parentID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if parentID.Valid {
		v := parentID.Int64
		category.ParentID = &v
	}
	return category, err
}

func GetCategoryByName(db DBTX, name string) (Category, error) {
	var category Category
	var parentID sql.NullInt64
	err := db.QueryRow(`
SELECT id, name, parent_id, created_at, updated_at
FROM categories
WHERE name = ?
`, name).Scan(
		&category.ID,
		&category.Name,
		&parentID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if parentID.Valid {
		v := parentID.Int64
		category.ParentID = &v
	}
	return category, err
}

func ListCategories(db DBTX, filter CategoryListFilter) ([]Category, error) {
	query := `
SELECT id, name, parent_id, created_at, updated_at
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
		var parentID sql.NullInt64
		err = rows.Scan(
			&category.ID,
			&category.Name,
			&parentID,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if parentID.Valid {
			v := parentID.Int64
			category.ParentID = &v
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func CategoryExists(db DBTX, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM categories WHERE name = ?
)
`, name).Scan(&exists)
	return exists, err
}

func InsertCategory(db DBTX, input CreateCategoryInput) (int64, error) {
	parentID := input.ParentID
	if parentID == nil && input.Name != rootCategoryName {
		rootID, err := getCategoryIDByName(db, rootCategoryName)
		if err != nil {
			return 0, err
		}
		parentID = &rootID
	}

	result, err := db.Exec(`
INSERT INTO categories (name, parent_id)
VALUES (?, ?)
`, input.Name, parentID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func getCategoryIDByName(db DBTX, name string) (int64, error) {
	var categoryID int64
	err := db.QueryRow(`
SELECT id
FROM categories
WHERE name = ?
`, name).Scan(&categoryID)
	if err != nil {
		return 0, err
	}

	return categoryID, nil
}

func (c Category) GetParent(db DBTX) (*Category, error) {
	if c.ParentID == nil {
		return nil, nil
	}

	parent, err := GetCategoryByID(db, *c.ParentID)
	if err != nil {
		return nil, err
	}

	return &parent, nil
}
