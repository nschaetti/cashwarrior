package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Budget(parsed parser.ParsedCmdLine, _ config.Config) error {
	fmt.Println("budget")
	return nil
}
