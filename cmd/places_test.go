package cmd

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestPlacesRename(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	if _, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Coop"}); err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	err := Places(parser.ParsedCmdLine{
		Command:    "places",
		Subcommand: "rename",
		Args:       []parser.Token{{Kind: parser.TokenText, Raw: "Coop"}, {Kind: parser.TokenText, Raw: "Migros"}},
	}, config.GetDefaultConfig(), cashDB)
	if err != nil {
		t.Fatalf("Places(rename) returned error: %v", err)
	}

	if _, err := db.GetPlaceByName(cashDB, "Migros"); err != nil {
		t.Fatalf("GetPlaceByName(Migros) returned error: %v", err)
	}
}
