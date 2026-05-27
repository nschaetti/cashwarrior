package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

const (
	defaultFakeitCount    = 25
	minimumFakeAccounts   = 4
	minimumFakeCategories = 8
	minimumFakePlaces     = 8
	minimumFakeTags       = 8
	minimumFakeGroups     = 4
)

var fakeitMonthNames = map[string]time.Month{
	"january":   time.January,
	"jan":       time.January,
	"february":  time.February,
	"feb":       time.February,
	"march":     time.March,
	"mar":       time.March,
	"april":     time.April,
	"apr":       time.April,
	"may":       time.May,
	"june":      time.June,
	"jun":       time.June,
	"july":      time.July,
	"jul":       time.July,
	"august":    time.August,
	"aug":       time.August,
	"september": time.September,
	"sep":       time.September,
	"october":   time.October,
	"oct":       time.October,
	"november":  time.November,
	"nov":       time.November,
	"december":  time.December,
	"dec":       time.December,
}

type fakeitOptions struct {
	AccountName  string
	CategoryName string
	Type         string
	Year         int
	Month        time.Month
	Count        int
}

type fakeitContext struct {
	Options          fakeitOptions
	Accounts         []db.Account
	Categories       []db.Category
	Places           []db.Place
	Tags             []db.Tag
	Groups           []db.TransactionGroup
	SpecificAccount  *db.Account
	SpecificCategory *db.Category
	DateFrom         time.Time
	DateTo           time.Time
	SequenceByMonth  map[string]int
}

func Fakeit(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	if len(parsed.Filters) != 0 {
		return fmt.Errorf("no filters allowed")
	}

	gofakeit.Seed(time.Now().UnixNano())

	options, err := parseFakeitOptions(parsed)
	if err != nil {
		return err
	}

	ctx, err := prepareFakeitContext(cashDb, cfg, options)
	if err != nil {
		return err
	}

	createdTransactions := 0
	createdTransfers := 0
	for range options.Count {
		if options.Type == "transfer" {
			if err := insertFakeTransfer(cashDb, ctx); err != nil {
				return err
			}
			createdTransfers++
			createdTransactions += 2
			continue
		}

		if err := insertFakeTransaction(cashDb, ctx, options.Type); err != nil {
			return err
		}
		createdTransactions++
	}

	pterm.Success.Printf(
		"Created %d transactions and %d transfers\n",
		createdTransactions,
		createdTransfers,
	)
	return nil
}

func parseFakeitOptions(parsed parser.ParsedCmdLine) (fakeitOptions, error) {
	options := fakeitOptions{Count: defaultFakeitCount}
	seenAttributes := make(map[string]bool)

	for _, arg := range parsed.Args {
		switch arg.Kind {
		case parser.TokenAttribute:
			if seenAttributes[arg.Key] {
				return options, fmt.Errorf("attribute %s specified multiple times", arg.Key)
			}
			seenAttributes[arg.Key] = true

			switch arg.Key {
			case "account":
				options.AccountName = arg.Value
			case "category":
				options.CategoryName = arg.Value
			case "type":
				transactionType := strings.ToLower(arg.Value)
				if transactionType != "expense" && transactionType != "income" && transactionType != "transfer" {
					return options, fmt.Errorf("unsupported type %s", arg.Value)
				}
				options.Type = transactionType
			case "year":
				year, err := strconv.Atoi(arg.Value)
				if err != nil {
					return options, fmt.Errorf("invalid year %s", arg.Value)
				}
				options.Year = year
			case "month":
				month, ok := fakeitMonthNames[strings.ToLower(arg.Value)]
				if !ok {
					return options, fmt.Errorf("invalid month %s", arg.Value)
				}
				options.Month = month
			default:
				return options, fmt.Errorf("unsupported attribute %s", arg.Key)
			}
		case parser.TokenText:
			count, err := strconv.Atoi(arg.Raw)
			if err != nil || count <= 0 {
				return options, fmt.Errorf("unsupported argument %s", arg.Raw)
			}
			if options.Count != defaultFakeitCount {
				return options, fmt.Errorf("count specified multiple times")
			}
			options.Count = count
		default:
			return options, fmt.Errorf("unsupported token %s", arg.Raw)
		}
	}

	if options.Month != 0 && options.Year == 0 {
		options.Year = time.Now().Year()
	}

	if options.Type == "" {
		options.Type = "mixed"
	}

	return options, nil
}

func prepareFakeitContext(cashDb db.DBTX, cfg config.Config, options fakeitOptions) (*fakeitContext, error) {
	ctx := &fakeitContext{Options: options}

	if options.AccountName != "" {
		account, err := db.GetAccountByName(cashDb, options.AccountName)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account %s does not exist", options.AccountName)
		}
		if err != nil {
			return nil, err
		}
		ctx.SpecificAccount = &account
	}

	if options.CategoryName != "" {
		category, err := db.GetCategoryByName(cashDb, options.CategoryName)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category %s does not exist", options.CategoryName)
		}
		if err != nil {
			return nil, err
		}
		ctx.SpecificCategory = &category
	}

	if err := ensureFakeData(cashDb, cfg); err != nil {
		return nil, err
	}

	accounts, err := db.ListAccounts(cashDb, db.AccountListFilter{})
	if err != nil {
		return nil, err
	}
	ctx.Accounts = accounts

	categories, err := db.ListCategories(cashDb, db.CategoryListFilter{})
	if err != nil {
		return nil, err
	}
	ctx.Categories = make([]db.Category, 0, len(categories))
	for _, category := range categories {
		if category.Name == "root" {
			continue
		}
		ctx.Categories = append(ctx.Categories, category)
	}

	places, err := db.ListPlaces(cashDb, db.PlaceListFilter{})
	if err != nil {
		return nil, err
	}
	ctx.Places = places

	tags, err := db.ListTags(cashDb, db.TagListFilter{})
	if err != nil {
		return nil, err
	}
	ctx.Tags = tags

	groups, err := db.ListTransactionGroups(cashDb, db.TransactionGroupListFilter{})
	if err != nil {
		return nil, err
	}
	ctx.Groups = groups

	ctx.DateFrom, ctx.DateTo = resolveFakeitDateRange(options)

	ctx.SequenceByMonth, err = loadFakeitSequenceByMonth(cashDb)
	if err != nil {
		return nil, err
	}

	if len(ctx.Accounts) == 0 {
		return nil, fmt.Errorf("no accounts available")
	}
	if len(ctx.Categories) == 0 {
		return nil, fmt.Errorf("no categories available")
	}
	if len(ctx.Places) == 0 {
		return nil, fmt.Errorf("no places available")
	}

	if options.Type == "transfer" {
		compatibleAccounts := ctx.compatibleTransferAccounts()
		if len(compatibleAccounts) < 2 {
			return nil, fmt.Errorf("not enough compatible accounts to create transfers")
		}
	}

	return ctx, nil
}

func ensureFakeData(cashDb db.DBTX, cfg config.Config) error {
	if err := ensureFakeAccounts(cashDb, cfg); err != nil {
		return err
	}
	if err := ensureFakeCategories(cashDb); err != nil {
		return err
	}
	if err := ensureFakePlaces(cashDb); err != nil {
		return err
	}
	if err := ensureFakeTags(cashDb); err != nil {
		return err
	}
	if err := ensureFakeGroups(cashDb); err != nil {
		return err
	}
	return nil
}

func ensureFakeAccounts(cashDb db.DBTX, cfg config.Config) error {
	accounts, err := db.ListAccounts(cashDb, db.AccountListFilter{})
	if err != nil {
		return err
	}

	defaultCurrencyCount := 0
	for _, account := range accounts {
		if account.Currency == cfg.Default.Currency {
			defaultCurrencyCount++
		}
	}

	for len(accounts) < minimumFakeAccounts || defaultCurrencyCount < 2 {
		name := sanitizeFakeName(strings.ToLower(gofakeit.Company()) + "-wallet")
		if exists, err := db.AccountExists(cashDb, name); err != nil {
			return err
		} else if exists {
			continue
		}

		currency := cfg.Default.Currency
		if len(accounts) >= 2 && defaultCurrencyCount >= 2 {
			currency = []string{cfg.Default.Currency, "EUR", "USD", "GBP"}[rand.Intn(4)]
		}

		if _, err := db.InsertAccount(cashDb, db.CreateAccountInput{Name: name, Currency: currency}); err != nil {
			return err
		}
		accounts = append(accounts, db.Account{Name: name, Currency: currency})
		if currency == cfg.Default.Currency {
			defaultCurrencyCount++
		}
	}
	return nil
}

func ensureFakeCategories(cashDb db.DBTX) error {
	categories, err := db.ListCategories(cashDb, db.CategoryListFilter{})
	if err != nil {
		return err
	}

	nonRoot := 0
	for _, category := range categories {
		if category.Name != "root" {
			nonRoot++
		}
	}

	for nonRoot < minimumFakeCategories {
		name := sanitizeFakeName(strings.ToLower(gofakeit.ProductCategory()))
		if name == "root" {
			continue
		}
		if exists, err := db.CategoryExists(cashDb, name); err != nil {
			return err
		} else if exists {
			continue
		}
		if _, err := db.InsertCategory(cashDb, db.CreateCategoryInput{Name: name}); err != nil {
			return err
		}
		nonRoot++
	}
	return nil
}

func ensureFakePlaces(cashDb db.DBTX) error {
	places, err := db.ListPlaces(cashDb, db.PlaceListFilter{})
	if err != nil {
		return err
	}

	for len(places) < minimumFakePlaces {
		name := sanitizeFakeName(strings.ToLower(gofakeit.Company()))
		if exists, err := db.PlaceExists(cashDb, name); err != nil {
			return err
		} else if exists {
			continue
		}
		if _, err := db.InsertPlace(cashDb, db.CreatePlaceInput{Name: name}); err != nil {
			return err
		}
		places = append(places, db.Place{Name: name})
	}
	return nil
}

func ensureFakeTags(cashDb db.DBTX) error {
	tags, err := db.ListTags(cashDb, db.TagListFilter{})
	if err != nil {
		return err
	}

	for len(tags) < minimumFakeTags {
		name := sanitizeFakeName(strings.ToLower(gofakeit.JobTitle()))
		if exists, err := db.TagExists(cashDb, name); err != nil {
			return err
		} else if exists {
			continue
		}
		if _, err := db.InsertTag(cashDb, db.CreateTagInput{Name: name}); err != nil {
			return err
		}
		tags = append(tags, db.Tag{Name: name})
	}
	return nil
}

func ensureFakeGroups(cashDb db.DBTX) error {
	groups, err := db.ListTransactionGroups(cashDb, db.TransactionGroupListFilter{})
	if err != nil {
		return err
	}

	for len(groups) < minimumFakeGroups {
		name := sanitizeFakeName(strings.ToLower(gofakeit.BuzzWord()))
		if exists, err := db.TransactionGroupExists(cashDb, name); err != nil {
			return err
		} else if exists {
			continue
		}
		if _, err := db.InsertTransactionGroup(cashDb, db.CreateTransactionGroupInput{Name: name}); err != nil {
			return err
		}
		groups = append(groups, db.TransactionGroup{Name: name})
	}
	return nil
}

func resolveFakeitDateRange(options fakeitOptions) (time.Time, time.Time) {
	now := time.Now()
	location := now.Location()

	if options.Year != 0 && options.Month != 0 {
		start := time.Date(options.Year, options.Month, 1, 0, 0, 0, 0, location)
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		return start, end
	}

	if options.Year != 0 {
		start := time.Date(options.Year, time.January, 1, 0, 0, 0, 0, location)
		end := start.AddDate(1, 0, 0).Add(-time.Nanosecond)
		return start, end
	}

	start := now.AddDate(0, -3, 0)
	return start, now
}

func loadFakeitSequenceByMonth(cashDb db.DBTX) (map[string]int, error) {
	transactions, err := db.ListTransactions(cashDb, []db.SQLFilter{}, []db.Filter[db.Transaction]{})
	if err != nil {
		return nil, err
	}

	sequenceByMonth := make(map[string]int)
	for _, transaction := range transactions {
		identifier, err := domain.ParseTransactionID(transaction.Identifier)
		if err != nil {
			continue
		}
		key := fakeitSequenceKey(identifier.Year, identifier.Month)
		if identifier.Num > sequenceByMonth[key] {
			sequenceByMonth[key] = identifier.Num
		}
	}
	return sequenceByMonth, nil
}

func fakeitSequenceKey(year int, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}

func (ctx *fakeitContext) nextIdentifier(at time.Time) string {
	key := fakeitSequenceKey(at.Year(), int(at.Month()))
	ctx.SequenceByMonth[key]++
	return domain.TransactionID{Year: at.Year(), Month: int(at.Month()), Num: ctx.SequenceByMonth[key]}.String()
}

func (ctx *fakeitContext) randomDate() time.Time {
	if !ctx.DateFrom.Before(ctx.DateTo) {
		return ctx.DateFrom
	}
	return gofakeit.DateRange(ctx.DateFrom, ctx.DateTo)
}

func (ctx *fakeitContext) randomAccount() db.Account {
	if ctx.SpecificAccount != nil {
		return *ctx.SpecificAccount
	}
	return ctx.Accounts[rand.Intn(len(ctx.Accounts))]
}

func (ctx *fakeitContext) randomCategory() db.Category {
	if ctx.SpecificCategory != nil {
		return *ctx.SpecificCategory
	}
	return ctx.Categories[rand.Intn(len(ctx.Categories))]
}

func (ctx *fakeitContext) randomPlace(includeTransfer bool) db.Place {
	places := make([]db.Place, 0, len(ctx.Places))
	for _, place := range ctx.Places {
		if !includeTransfer && place.Name == "transfer" {
			continue
		}
		places = append(places, place)
	}
	if len(places) == 0 {
		return ctx.Places[rand.Intn(len(ctx.Places))]
	}
	return places[rand.Intn(len(places))]
}

func (ctx *fakeitContext) randomGroupID() *int64 {
	if len(ctx.Groups) == 0 || rand.Intn(100) < 55 {
		return nil
	}
	group := ctx.Groups[rand.Intn(len(ctx.Groups))]
	return &group.ID
}

func (ctx *fakeitContext) compatibleTransferAccounts() []db.Account {
	if ctx.SpecificAccount == nil {
		currencies := make(map[string]int)
		for _, account := range ctx.Accounts {
			currencies[account.Currency]++
		}
		compatible := make([]db.Account, 0, len(ctx.Accounts))
		for _, account := range ctx.Accounts {
			if currencies[account.Currency] >= 2 {
				compatible = append(compatible, account)
			}
		}
		return compatible
	}

	compatible := []db.Account{*ctx.SpecificAccount}
	for _, account := range ctx.Accounts {
		if account.ID == ctx.SpecificAccount.ID {
			continue
		}
		if account.Currency == ctx.SpecificAccount.Currency {
			compatible = append(compatible, account)
		}
	}
	return compatible
}

func insertFakeTransaction(cashDb db.DBTX, ctx *fakeitContext, forcedType string) error {
	transactionTime := ctx.randomDate()
	transactionType := forcedType
	if transactionType == "mixed" {
		if rand.Intn(100) < 78 {
			transactionType = "expense"
		} else {
			transactionType = "income"
		}
	}

	amount := randomTransactionAmount(transactionType)
	account := ctx.randomAccount()
	category := ctx.randomCategory()
	place := ctx.randomPlace(false)
	groupID := ctx.randomGroupID()

	transactionID, err := db.InsertTransaction(cashDb, db.CreateTransactionInput{
		Identifier:  ctx.nextIdentifier(transactionTime),
		Type:        transactionType,
		Amount:      amount,
		Description: fakeitDescription(transactionType),
		Datetime:    transactionTime,
		AccountID:   account.ID,
		CategoryID:  &category.ID,
		PlaceID:     &place.ID,
		GroupID:     groupID,
	})
	if err != nil {
		return err
	}

	return attachRandomTags(cashDb, ctx.Tags, transactionID)
}

func insertFakeTransfer(cashDb db.DBTX, ctx *fakeitContext) error {
	compatible := ctx.compatibleTransferAccounts()
	fromAccount := compatible[rand.Intn(len(compatible))]

	otherAccounts := make([]db.Account, 0, len(compatible)-1)
	for _, account := range compatible {
		if account.ID != fromAccount.ID && account.Currency == fromAccount.Currency {
			otherAccounts = append(otherAccounts, account)
		}
	}
	if len(otherAccounts) == 0 {
		return fmt.Errorf("not enough compatible accounts to create transfers")
	}
	toAccount := otherAccounts[rand.Intn(len(otherAccounts))]

	transactionTime := ctx.randomDate()
	amount := randomTransactionAmount("transfer")
	transferPlace, err := db.GetPlaceByName(cashDb, "transfer")
	if err != nil {
		return err
	}
	description := fakeitDescription("transfer")

	fromTransactionID, err := db.InsertTransaction(cashDb, db.CreateTransactionInput{
		Identifier:  ctx.nextIdentifier(transactionTime),
		Type:        "transfer_out",
		Amount:      -amount,
		Description: description,
		Datetime:    transactionTime,
		AccountID:   fromAccount.ID,
		PlaceID:     &transferPlace.ID,
	})
	if err != nil {
		return err
	}

	toTransactionID, err := db.InsertTransaction(cashDb, db.CreateTransactionInput{
		Identifier:  ctx.nextIdentifier(transactionTime),
		Type:        "transfer_in",
		Amount:      amount,
		Description: description,
		Datetime:    transactionTime,
		AccountID:   toAccount.ID,
		PlaceID:     &transferPlace.ID,
	})
	if err != nil {
		return err
	}

	if _, err := db.InsertTransfer(cashDb, db.CreateTransferInput{
		FromTransactionID: fromTransactionID,
		ToTransactionID:   toTransactionID,
		FromAccountID:     fromAccount.ID,
		ToAccountID:       toAccount.ID,
		Amount:            amount,
	}); err != nil {
		return err
	}

	if err := attachRandomTags(cashDb, ctx.Tags, fromTransactionID); err != nil {
		return err
	}
	return attachRandomTags(cashDb, ctx.Tags, toTransactionID)
}

func attachRandomTags(cashDb db.DBTX, tags []db.Tag, transactionID int64) error {
	if len(tags) == 0 {
		return nil
	}
	maxTags := min(3, len(tags))
	count := rand.Intn(maxTags + 1)
	if count == 0 {
		return nil
	}

	used := make(map[int64]bool)
	for len(used) < count {
		tag := tags[rand.Intn(len(tags))]
		if used[tag.ID] {
			continue
		}
		if err := db.InsertTransactionTag(cashDb, transactionID, tag.ID); err != nil {
			return err
		}
		used[tag.ID] = true
	}
	return nil
}

func randomTransactionAmount(transactionType string) float64 {
	switch transactionType {
	case "income":
		return roundFakeAmount(gofakeit.Float64Range(200, 6500))
	case "transfer":
		return roundFakeAmount(gofakeit.Float64Range(25, 2500))
	default:
		return -roundFakeAmount(gofakeit.Float64Range(5, 450))
	}
}

func roundFakeAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func fakeitDescription(transactionType string) string {
	switch transactionType {
	case "income":
		return fmt.Sprintf("%s payroll", gofakeit.Company())
	case "transfer":
		return fmt.Sprintf("Transfer to %s", gofakeit.Company())
	default:
		return fmt.Sprintf("%s at %s", gofakeit.ProductName(), gofakeit.Company())
	}
}

func sanitizeFakeName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ".", "-", ",", "", "'", "")
	value = replacer.Replace(value)
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	return strings.Trim(value, "-")
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
