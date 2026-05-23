package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Tags(parsed parser.ParsedCmdLine, _ config.Config) error {
	fmt.Println("tags")
	return nil
}
