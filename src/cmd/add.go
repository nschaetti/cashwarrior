package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Add(parsed parser.ParsedCmdLine, _ config.Config) error {
	fmt.Println("add")
	return nil
}
