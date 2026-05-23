package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Modify(parsed parser.ParsedCmdLine, _ config.Config) error {
	fmt.Println("modify")
	return nil
}
