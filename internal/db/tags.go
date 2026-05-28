package db

import "time"

type Tag struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TagListFilter struct {
	NameLike string
	Limit    int
	Offset   int
}

type CreateTagInput struct {
	Name string
}

func GetTagByID(db DBTX, id int64) (Tag, error) {
	var tag Tag
	err := db.QueryRow(`
SELECT id, name, created_at, updated_at
FROM tags
WHERE id = ?
`, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)
	return tag, err
}

func GetTagByName(db DBTX, name string) (Tag, error) {
	var tag Tag
	err := db.QueryRow(`
SELECT id, name, created_at, updated_at
FROM tags
WHERE name = ?
`, name).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)
	return tag, err
}

func ListTags(db DBTX, filter TagListFilter) ([]Tag, error) {
	query := `
SELECT id, name, created_at, updated_at
FROM tags
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

	tags := make([]Tag, 0)
	for rows.Next() {
		var tag Tag
		err = rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func TagExists(db DBTX, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM tags WHERE name = ?
)
`, name).Scan(&exists)
	return exists, err
}

func InsertTag(db DBTX, input CreateTagInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO tags (name)
VALUES (?)
`, input.Name)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func UpdateTagName(db DBTX, tagID int64, name string) error {
	_, err := db.Exec(`
UPDATE tags
SET name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, name, tagID)
	return err
}

func DeleteTagByID(db DBTX, tagID int64) error {
	_, err := db.Exec(`
DELETE FROM tags
WHERE id = ?
`, tagID)
	return err
}

func CountTransactionsByTagID(db DBTX, tagID int64) (int, error) {
	var count int
	err := db.QueryRow(`
SELECT COUNT(*)
FROM transaction_tags
WHERE tag_id = ?
`, tagID).Scan(&count)
	return count, err
}
