package main

import (
	"fmt"
	"os"

	"github.com/nschaetti/cashwarrior/cmd"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func main() {
	// Check configuration exists
	cfg, err := config.InitConfig()
	if err != nil {
		fmt.Println("Error creating config file:", err)
		os.Exit(1)
	}
	fmt.Println(cfg)

	// Parse the command line and call the appropriate command.
	parsedCmd, parseErr := parser.ParseAndValidateCmdLine(os.Args[1:])
	if parseErr != nil && parseErr.Code != parser.ParseErrorNoCommand {
		// Just call "list" if there is no command.
		parsedCmd = parser.ParsedCmdLine{
			Command: "list",
			Filters: []parser.Token{},
			Args:    []parser.Token{},
		}
	} else if parseErr != nil {
		fmt.Println("Error:", parseErr.Message)
		os.Exit(1)
	}
	dispatchErr := cmd.Dispatch(parsedCmd, cfg)
	if dispatchErr != nil {
		fmt.Println(dispatchErr)
	}
}
