package cmd

import (
	"database/sql"

	"github.com/nschaetti/cashwarrior/internal/parser"
)
import "github.com/nschaetti/cashwarrior/internal/config"

type CommandFunc func(parsed parser.ParsedCmdLine, config config.Config, db *sql.DB) error

var Commands = map[string]CommandFunc{
	"init":        Init,
	"add":         Add,
	"show":        Show,
	"categories":  Categories,
	"stats":       Stats,
	"tags":        Tags,
	"modify":      Modify,
	"report":      Report,
	"list":        List,
	"delete":      Delete,
	"undo":        Undo,
	"by":          By,
	"accounts":    Accounts,
	"balance":     Balance,
	"transfer":    Transfer,
	"set-balance": SetBalance,
	"budget":      Budget,
	"config":      Config,
	"theme":       Theme,
	"sum":         Sum,
}

func Dispatch(parsed parser.ParsedCmdLine, cfg config.Config, cashDb *sql.DB) error {
	fn, ok := Commands[parsed.Command]
	if !ok {
		return &parser.ParseError{Code: parser.ParseErrorNoCommand, Message: "unknown command"}
	}
	return fn(parsed, cfg, cashDb)
}
