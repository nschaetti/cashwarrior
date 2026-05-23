package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Init(parsed parser.ParsedCmdLine, _ config.Config) error {
	fmt.Println("init")
	return nil
}
