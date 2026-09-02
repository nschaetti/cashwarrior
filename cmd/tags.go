package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/nschaetti/cashwarrior/internal/utils"
)

func Tags(parsed parser.ParsedCmdLine, _ config.Config, query db.DBTX) error {
	switch parsed.Subcommand {
	case "list":
		format, err := commandOutputFormat(parsed)
		if err != nil {
			return err
		}
		if format == output.FormatJSON {
			data, err := getTagsData(query)
			if err != nil {
				return err
			}
			return renderJSON("tags", data, len(data.Tags))
		}
		return listTags(query)
	case "add":
		return addTag(parsed, query)
	case "modify":
		return modifyTag(parsed, query)
	case "delete":
		return deleteTag(parsed, query)
	default:
		return fmt.Errorf("unknown tags subcommand %s", parsed.Subcommand)
	}
}

func listTags(query db.DBTX) error {
	tags, err := db.ListTags(query, db.TagListFilter{})
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(tags))
	for _, tag := range tags {
		count, err := db.CountTransactionsByTagID(query, tag.ID)
		if err != nil {
			return err
		}
		rows = append(rows, []string{strconv.FormatInt(tag.ID, 10), tag.Name, strconv.Itoa(count)})
	}

	theme := gui.CurrentTheme()
	t := gui.NewTable().
		WithTitle("Tags", theme.CategoriesTitleBackground).
		WithSubtitle("Configured tags").
		WithHeaderBackground(theme.CategoriesHeaderBackground).
		WithHeaders("ID", "Name", "Transactions").
		AddRows(rows)

	fmt.Println(t.Render())
	fmt.Println()
	fmt.Println()
	return nil
}

func getTagNameArg(arg parser.Arg) (string, error) {
	text, ok := arg.(parser.ArgText)
	if ok {
		if text.Text == "" {
			return "", fmt.Errorf("tag name cannot be empty")
		}
		return text.Text, nil
	}
	attr, ok := arg.(parser.ArgAttribute)
	if ok && attr.Key == "tag" && !attr.Value.IsEmpty() {
		return attr.Value.Raw, nil
	}
	return "", fmt.Errorf("tag name is required")
}

func addTag(parsed parser.ParsedCmdLine, query db.DBTX) error {
	name, err := getTagNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	exists, err := db.TagExists(query, name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("tag %s already exists", name)
	}
	if _, err := db.InsertTag(query, db.CreateTagInput{Name: name}); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("tag", map[string]any{"action": "created", "name": name}, 1)
	}
	fmt.Printf("Tag %s created\n", name)
	return nil
}

func modifyTag(parsed parser.ParsedCmdLine, query db.DBTX) error {
	name, err := getTagNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	tag, err := db.GetTagByName(query, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("tag %s does not exist", name)
	}
	if err != nil {
		return err
	}
	newName := ""
	for _, arg := range parsed.Args[1:] {
		attr, ok := arg.(parser.ArgAttribute)
		if ok && attr.Key == "tag" {
			newName = attr.Value.Raw
		}
	}
	if newName == "" {
		return fmt.Errorf("new tag name cannot be empty")
	}
	if newName != tag.Name {
		exists, err := db.TagExists(query, newName)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("tag %s already exists", newName)
		}
		if err := db.UpdateTagName(query, tag.ID, newName); err != nil {
			return err
		}
	}
	if isJSONOutput(parsed) {
		return renderJSON("tag", map[string]any{"action": "updated", "name": name}, 1)
	}
	fmt.Printf("Tag %s updated\n", name)
	return nil
}

func deleteTag(parsed parser.ParsedCmdLine, query db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	name, err := getTagNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	tag, err := db.GetTagByName(query, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("tag %s does not exist", name)
	}
	if err != nil {
		return err
	}
	count, err := db.CountTransactionsByTagID(query, tag.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("tag %s has linked transactions", name)
	}
	if !parsed.HasFlag("yes") && !utils.AskYesNo(fmt.Sprintf("Delete tag %s?", name)) {
		return nil
	}
	if err := db.DeleteTagByID(query, tag.ID); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("tag", map[string]any{"action": "deleted", "name": name}, 1)
	}
	fmt.Printf("Tag %s deleted\n", name)
	return nil
}
