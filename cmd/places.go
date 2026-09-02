package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Places(parsed parser.ParsedCmdLine, _ config.Config, cashDb db.DBTX) error {
	switch parsed.Subcommand {
	case "list":
		format, err := commandOutputFormat(parsed)
		if err != nil {
			return err
		}
		if format == output.FormatJSON {
			data, err := getPlacesData(cashDb)
			if err != nil {
				return err
			}
			return renderJSON("places", data, len(data.Places))
		}
		return listPlaces(cashDb)
	case "rename":
		return renamePlace(parsed, cashDb)
	default:
		return fmt.Errorf("unknown places subcommand %s", parsed.Subcommand)
	}
}

func listPlaces(cashDb db.DBTX) error {
	places, err := db.ListPlaces(cashDb, db.PlaceListFilter{})
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(places))
	for _, place := range places {
		rows = append(rows, []string{strconv.FormatInt(place.ID, 10), place.Name})
	}

	theme := gui.CurrentTheme()
	t := gui.NewTable().
		WithTitle("Places", theme.CategoriesTitleBackground).
		WithSubtitle("Configured places").
		WithHeaderBackground(theme.CategoriesHeaderBackground).
		WithHeaders("ID", "Name").
		AddRows(rows)

	fmt.Println(t.Render())
	fmt.Println()
	fmt.Println()
	return nil
}

func renamePlace(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	oldName := strings.TrimSpace(parsed.Args[0].RawString())
	newName := strings.TrimSpace(parsed.Args[1].RawString())

	if oldName == "" || newName == "" {
		return fmt.Errorf("place names cannot be empty")
	}
	if oldName == newName {
		return fmt.Errorf("old and new place names are identical")
	}

	place, err := db.GetStoreByName(cashDb, oldName)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("place %s does not exist", oldName)
	}
	if err != nil {
		return err
	}

	exists, err := db.PlaceExists(cashDb, newName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("place %s already exists", newName)
	}

	if err := db.UpdatePlaceName(cashDb, place.ID, newName); err != nil {
		return err
	}

	if isJSONOutput(parsed) {
		return renderJSON("place", map[string]any{"action": "renamed", "name": newName}, 1)
	}
	fmt.Printf("Place %s renamed to %s\n", oldName, newName)
	return nil
}
