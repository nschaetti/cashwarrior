package cmd

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestAddTag(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	err := addTag(parser.ParsedCmdLine{Command: "tags", Subcommand: "add", Args: []parser.Token{{Kind: parser.TokenText, Raw: "travel"}}}, cashDB)
	if err != nil {
		t.Fatalf("addTag returned error: %v", err)
	}

	tag, err := db.GetTagByName(cashDB, "travel")
	if err != nil {
		t.Fatalf("GetTagByName returned error: %v", err)
	}
	if tag.Name != "travel" {
		t.Fatalf("name = %q, want travel", tag.Name)
	}
}

func TestModifyTag(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertTag(cashDB, db.CreateTagInput{Name: "travel"})
	if err != nil {
		t.Fatalf("InsertTag returned error: %v", err)
	}

	err = modifyTag(parser.ParsedCmdLine{Command: "tags", Subcommand: "modify", Args: []parser.Token{{Kind: parser.TokenText, Raw: "travel"}, {Kind: parser.TokenAttribute, Key: "tag", Value: "vacation", Raw: "tag:vacation"}}}, cashDB)
	if err != nil {
		t.Fatalf("modifyTag returned error: %v", err)
	}

	tag, err := db.GetTagByName(cashDB, "vacation")
	if err != nil {
		t.Fatalf("GetTagByName returned error: %v", err)
	}
	if tag.Name != "vacation" {
		t.Fatalf("name = %q, want vacation", tag.Name)
	}
}

func TestDeleteTagDeletesUnusedTagAfterConfirmation(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertTag(cashDB, db.CreateTagInput{Name: "travel"})
	if err != nil {
		t.Fatalf("InsertTag returned error: %v", err)
	}

	withInput(t, "y\n", func() {
		err = deleteTag(parser.ParsedCmdLine{Command: "tags", Subcommand: "delete", Args: []parser.Token{{Kind: parser.TokenText, Raw: "travel"}}}, cashDB)
	})
	if err != nil {
		t.Fatalf("deleteTag returned error: %v", err)
	}
}
