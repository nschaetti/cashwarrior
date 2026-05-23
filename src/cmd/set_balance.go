package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func SetBalance(parsed parser.ParsedCmdLine, _ config.Config) error {
	fmt.Println("set-balance")
	return nil
}
