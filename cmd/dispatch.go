package cmd

import (
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)
import "github.com/nschaetti/cashwarrior/internal/config"

type CommandFunc func(parsed parser.ParsedCmdLine, config config.Config, db db.DBTX) error

var Handlers = map[string]CommandFunc{
	"init":        Init,
	"add":         Add,
	"show":        Show,
	"categories":  Categories,
	"stats":       Stats,
	"group":       Group,
	"tags":        Tags,
	"modify":      Modify,
	"report":      Report,
	"list":        List,
	"delete":      Delete,
	"purge":       Purge,
	"restore":     Restore,
	"undo":        Undo,
	"by":          By,
	"accounts":    Accounts,
	"balance":     Balance,
	"fakeit":      Fakeit,
	"transfer":    Transfer,
	"set-balance": SetBalance,
	"budget":      Budget,
	"config":      Config,
	"backup":      Backup,
	"import":      Import,
	"theme":       Theme,
	"sum":         Sum,
	"summary":     Summary,
	"groups":      Groups,
	"places":      Places,
	"help":        Help,
}

func Dispatch(parsed parser.ParsedCmdLine, cfg config.Config, tx db.DBTX) error {
	if parsed.HasFlag("help") {
		printHelp(parsed)
		return nil
	}

	if !parser.IsKnownCommand(parsed.Command) {
		return &parser.ParseError{Code: parser.ParseErrorNoCommand, Message: "unknown command"}
	}

	fn, ok := Handlers[parsed.Command]
	if !ok {
		return &parser.ParseError{Code: parser.ParseErrorNoCommand, Message: "command has no handler"}
	}
	return fn(parsed, cfg, tx)
}
