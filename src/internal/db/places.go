package db

import (
	"database/sql"
	"time"
)

type PlaceDBEntry struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PlaceListFilter struct {
	ID       *int64
	NameLike string
	Limit    int
	Offset   int
}

type CreatePlaceDBInput struct {
	Name string
	Kind string
}

func GetPlaceByID(db *sql.DB, id int64) (PlaceDBEntry, error) {
	var place PlaceDBEntry
	err := db.QueryRow(`
SELECT id, name, created_at, updated_at
FROM places
WHERE id = ?
`, id).Scan(
		&place.ID,
		&place.Name,
		&place.CreatedAt,
		&place.UpdatedAt,
	)
	if err != nil {
		return place, err
	}

	return place, nil
}

func GetPlaceByName(db *sql.DB, name string) (PlaceDBEntry, error) {
	var place PlaceDBEntry
	err := db.QueryRow(`
SELECT id, name, created_at, updated_at
FROM places
WHERE name = ?
`, name).Scan(
		&place.Name,
	)
	if err != nil {
		return place, err
	}
	return place, nil
}

func ListPlaces(db *sql.DB, filter PlaceListFilter) ([]PlaceDBEntry, error) {
	query := `
SELECT id, name, created_at, updated_at
FROM places
`
	args := make([]interface{}, 0)

	if filter.ID != nil {
		query += "WHERE id = ?\n"
		args = append(args, *filter.ID)
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

	places := make([]PlaceDBEntry, 0)
	for rows.Next() {
		var place PlaceDBEntry
		err = rows.Scan(
			&place.ID,
			&place.Name,
		)
		if err != nil {
			return nil, err
		}
		places = append(places, place)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return places, nil
}

func PlaceExists(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM places 
	WHERE name = ?
)
`, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func InsertPlace(db *sql.DB, input CreatePlaceDBInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO places (name)
VALUES (?)
`, input.Name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
