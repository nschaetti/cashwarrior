package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Accounts(parsed parser.ParsedCmdLine, _ config.Config) error {
	fmt.Println("accounts")
	return nil
}
