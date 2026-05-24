package cmd

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

var CommandValidations = []func(
	parser.ParsedCmdLine,
	config.Config,
	*sql.DB,
	map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error){
	validateAtLeastTwoArgs,
	validateNoFilters,
	validateAmount,
	validateAttributes,
}

func validateAtLeastTwoArgs(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	if len(parsed.Args) < 2 {
		return parsed, fmt.Errorf("we need at least an amount and a description")
	}
	return parsed, nil
}

func validateNoFilters(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	if len(parsed.Filters) != 0 {
		return parsed, fmt.Errorf("no filters allowed")
	}
	return parsed, nil
}

func validateAmount(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	// Zero or multiple amounts => problem
	if counts[parser.TokenAmount] == 0 {
		return parsed, fmt.Errorf("no amount specified")
	} else if counts[parser.TokenAmount] > 1 {
		name, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("y").
			Show("You specified multiple amounts. Would you like to sum them up?:")
		if err != nil {
			return parsed, err
		}
		if name != "y" {
			return parsed, fmt.Errorf("multiple amounts specified, only one allowed if not summed up")
		}

		// Sum up amounts
		var sumAmount float32
		for _, amount := range parsed.GetAmount() {
			sumAmount += amount.Amount
		}

		// No zero amount allowed
		if sumAmount == 0 {
			return parsed, fmt.Errorf("summed up amount is zero")
		}

		// Remove amounts from parsed command line
		parsed.RemoveByKind(parser.TokenAmount)

		// Add summed up amount
		parsed.Append(parser.Token{
			Raw:    strconv.FormatFloat(float64(sumAmount), 'f', 2, 32),
			Amount: sumAmount,
			Kind:   parser.TokenAmount,
		}, false)

		return parsed, nil
	}

	// We have an amount, check if it's valid
	amount := parsed.GetAmount()[0]

	// No zero amount allowed
	if amount.Amount == 0 {
		return parsed, fmt.Errorf("amount cannot be zero")
	}

	return parsed, nil
}

func checkAttribute(counts map[string]int, attr string, threshold int) error {
	if counts[attr] > threshold {
		return fmt.Errorf("only %d attributes %s allowed", threshold, attr)
	}
	return nil
}

func validateAttributes(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	// Get count of attributes
	attrsCount := parsed.GetAttributesCount(false)

	// Date
	if err := checkAttribute(attrsCount, "date", 1); err != nil {
		return parsed, err
	}

	// Time
	if err := checkAttribute(attrsCount, "time", 1); err != nil {
		return parsed, err
	}

	// Account
	if err := checkAttribute(attrsCount, "account", 1); err != nil {
		return parsed, err
	}

	// Category
	if err := checkAttribute(attrsCount, "category", 1); err != nil {
		return parsed, err
	}

	// Currency
	if err := checkAttribute(attrsCount, "currency", 1); err != nil {
		return parsed, err
	}

	return parsed, nil
}

func runCommandLineValidation(parsed parser.ParsedCmdLine, config config.Config, db *sql.DB, counts map[parser.TokenKind]int) (parser.ParsedCmdLine, error) {
	for _, check := range CommandValidations {
		var err error
		parsed, err = check(parsed, config, db, counts)
		if err != nil {
			return parsed, err
		}
	}
	return parsed, nil
}

func Add(parsed parser.ParsedCmdLine, config config.Config, db *sql.DB) error {
	// Get count by token kind for validation
	tokenKindsCount := parsed.GetTokenKindCount(false)

	// We check the commnand line first
	parsed, err := runCommandLineValidation(parsed, config, db, tokenKindsCount)
	if err != nil {
		return err
	}

	fmt.Println("add")
	fmt.Println(parsed)
	return nil
}
