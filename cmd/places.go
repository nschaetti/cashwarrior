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
	"github.com/nschaetti/cashwarrior/internal/utils"
)

const transferPlaceName = "transfer"

func Places(parsed parser.ParsedCmdLine, _ config.Config, cashDb db.DBTX) error {
	switch parsed.Subcommand {
	case "list", "ls":
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
	case "add":
		return addPlace(parsed, cashDb)
	case "rename", "rn":
		return renamePlace(parsed, cashDb)
	case "delete", "rm":
		return deletePlace(parsed, cashDb)
	default:
		return fmt.Errorf("unknown places subcommand %s", parsed.Subcommand)
	}
}

func addPlace(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	name := strings.TrimSpace(parsed.Args[0].RawString())
	if name == "" {
		return fmt.Errorf("place name cannot be empty")
	}
	exists, err := db.PlaceExists(cashDb, name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("place %s already exists", name)
	}
	if _, err := db.InsertStore(cashDb, db.CreatePlaceInput{Name: name}); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("place", map[string]any{"action": "created", "name": name}, 1)
	}
	fmt.Printf("Place %s created\n", name)
	return nil
}

func deletePlace(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	name := strings.TrimSpace(parsed.Args[0].RawString())
	if name == transferPlaceName {
		return fmt.Errorf("place transfer is required by transfers and cannot be deleted")
	}
	place, err := db.GetStoreByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("place %s does not exist", name)
	}
	if err != nil {
		return err
	}
	count, err := db.CountTransactionsByPlaceID(cashDb, place.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("place %s has linked transactions", name)
	}
	if !parsed.HasFlag("yes") && !utils.AskYesNo(fmt.Sprintf("Delete place %s?", name)) {
		return nil
	}
	if err := db.DeleteStoreByID(cashDb, place.ID); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("place", map[string]any{"action": "deleted", "name": name}, 1)
	}
	fmt.Printf("Place %s deleted\n", name)
	return nil
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
