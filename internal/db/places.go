package db

import "time"

type Place struct {
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

type CreatePlaceInput struct {
	Name string
	Kind string
}

func GetStoreByID(db DBTX, id int64) (Place, error) {
	var place Place
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

func GetStoreByName(db DBTX, name string) (Place, error) {
	var place Place
	err := db.QueryRow(`
SELECT id, name, created_at, updated_at
FROM places
WHERE name = ?
`, name).Scan(
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

func ListPlaces(db DBTX, filter PlaceListFilter) ([]Place, error) {
	query := `
SELECT id, name, created_at, updated_at
FROM places
`
	args := make([]interface{}, 0)

	if filter.ID != nil {
		query += "WHERE id = ?\n"
		args = append(args, *filter.ID)
	} else if filter.NameLike != "" {
		query += "WHERE name LIKE ?\n"
		args = append(args, "%"+filter.NameLike+"%")
	}

	query += "ORDER BY name\n"
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

	places := make([]Place, 0)
	for rows.Next() {
		var place Place
		err = rows.Scan(
			&place.ID,
			&place.Name,
			&place.CreatedAt,
			&place.UpdatedAt,
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

func InsertStore(db DBTX, input CreatePlaceInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO places (name)
VALUES (?)
`, input.Name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdatePlaceName(db DBTX, placeID int64, name string) error {
	_, err := db.Exec(`
UPDATE places
SET name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, name, placeID)
	return err
}

func PlaceExists(db DBTX, name string) (bool, error) {
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
