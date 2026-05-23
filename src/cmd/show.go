package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Show(parsed parser.ParsedCmdLine, _ config.Config) error {
	fmt.Println("show")
	return nil
}
