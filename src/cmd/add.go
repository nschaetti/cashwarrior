package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
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
		for _, amount := range parsed.GetAmounts() {
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
	amount := parsed.GetAmounts()[0]

	// No zero amount allowed
	if amount.Amount == 0 {
		return parsed, fmt.Errorf("amount cannot be zero")
	}

	return parsed, nil
}

func validateAttributes(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	// Get count of attributes
	attrsCount := parsed.GetAttributesCount(false)

	for attr, count := range attrsCount {
		if count > 1 {
			return parsed, fmt.Errorf("attribute %s specified multiple times", attr)
		}
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

func getTransactionDescription(parsed parser.ParsedCmdLine, counts map[parser.TokenKind]int) string {
	if counts[parser.TokenText] == 1 {
		return parsed.Args[0].Raw
	}
	var desc string
	for _, arg := range parsed.Args {
		if arg.Kind != parser.TokenText {
			continue
		}
		desc += arg.Raw + " "
	}
	return desc[0 : len(desc)-1]
}

func getNextIdentifier(cashDb *sql.DB) (domain.TransactionID, error) {
	// Get next transaction identifier
	lastTransaction, err := db.GetLastTransaction(cashDb)
	if err != nil {
		return domain.TransactionID{}, err
	}

	// Parse identifier
	id, err := domain.ParseTransactionID(lastTransaction.Identifier)
	if err != nil {
		return domain.TransactionID{}, err
	}

	// Same year-month -> increment sequence number
	if id.Month == int(time.Now().Month()) && id.Year == int(time.Now().Year()) {
		return domain.TransactionID{
			Year:  id.Year,
			Month: id.Month,
			Num:   id.Num + 1,
		}, nil
	}

	return domain.TransactionID{
		Year:  time.Now().Year(),
		Month: int(time.Now().Month()),
		Num:   0,
	}, nil
}

func getTransactionAmount(parsed parser.ParsedCmdLine, counts map[parser.TokenKind]int) float32 {
	return parsed.GetAmounts()[0].Amount
}

func getAttributeValue(parsed parser.ParsedCmdLine, attr string) string {
	for _, arg := range parsed.Args {
		if arg.Kind != parser.TokenAttribute {
			continue
		}
		if arg.Key == attr {
			return arg.Value
		}
	}
	return ""
}

func getTransactionDatetime(
	attributes map[string]string,
	config config.Config,
) (time.Time, error) {
	// Check attributes exists
	var dateTimeExists bool
	var dateExists bool
	var timeExists bool
	for key, _ := range attributes {
		switch key {
		case "datetime":
			dateTimeExists = true
		case "date":
			dateExists = true
		case "time":
			timeExists = true
		}
	}
	fmt.Println(dateTimeExists, dateExists, timeExists)
	// Can be specified by datetime, or date alone (time 00:00:00), or date+time
	if dateTimeExists && dateExists {
		return time.Time{}, fmt.Errorf("datetime and date specified, only one allowed")
	}
	if dateTimeExists && timeExists {
		return time.Time{}, fmt.Errorf("datetime and time specified, only one allowed")

	}

	// Datetime given
	if dateTimeExists {
		return time.Parse(config.Display.DateFormat, attributes["datetime"])
	}

	// Date and time given
	if dateExists && timeExists {
		return time.Parse(
			config.Display.DateFormat,
			attributes["date"]+" "+attributes["time"],
		)
	} else if dateExists {
		return time.Parse(strings.Split(config.Display.DateFormat, " ")[0], attributes["date"])
	} else if timeExists {
		return time.Parse(strings.Split(config.Display.DateFormat, " ")[1], attributes["time"])
	}

	return time.Now(), nil
}

func getAttributes(parsed parser.ParsedCmdLine) map[string]string {
	var attributes map[string]string = make(map[string]string)
	for _, arg := range parsed.Args {
		if arg.Kind != parser.TokenAttribute {
			continue
		}
		attributes[arg.Key] = arg.Value
	}
	return attributes
}

func Add(parsed parser.ParsedCmdLine, config config.Config, cashDb *sql.DB) error {
	// Get count by token kind for validation
	tokenKindsCount := parsed.GetTokenKindCount(false)

	// We check the commnand line first
	parsed, err := runCommandLineValidation(parsed, config, cashDb, tokenKindsCount)
	if err != nil {
		return err
	}

	// Get attributes
	attributes := getAttributes(parsed)

	fmt.Println("add")
	fmt.Println(parsed)
	fmt.Println(attributes)

	// Get transaction description (merge it if necessary) & amount
	desc := getTransactionDescription(parsed, tokenKindsCount)
	amount := getTransactionAmount(parsed, tokenKindsCount)

	// Get next transaction identifier
	nextIdentifier, err := getNextIdentifier(cashDb)

	// Transaction type
	var transactionType string
	if amount < 0.0 {
		transactionType = "expense"
	} else {
		transactionType = "income"
	}

	// Get transaction datetime
	transactionTime, err := getTransactionDatetime(attributes, config)
	if err != nil {
		return err
	}

	// Get store
	//transactionStore := getTransactionStore(parsed)

	fmt.Println(desc)
	fmt.Println(nextIdentifier)
	fmt.Println(amount)
	fmt.Println(transactionType)
	fmt.Println(transactionTime)

	return nil
}
