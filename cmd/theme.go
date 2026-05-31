package cmd

import (
	"fmt"
	"sort"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/nschaetti/cashwarrior/internal/utils"
	"github.com/pterm/pterm"
)

func Theme(parsed parser.ParsedCmdLine, _ config.Config, query db.DBTX) error {
	_ = query

	themes := gui.ThemeNames()
	sort.Strings(themes)

	if len(parsed.Args) == 0 {
		fmt.Println()
		fmt.Println("Available themes:")
		for _, themeName := range themes {
			fmt.Printf("  %s\n", themeName)
		}
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  cash theme <theme-name>")
		fmt.Println()
		return nil
	}

	themeName := parsed.Args[0].Raw
	if !gui.ThemeExists(themeName) {
		return fmt.Errorf("unknown theme %q (available: %s)", themeName, joinThemes(themes))
	}

	configPath := utils.ExpandPath(config.DefaultConfigFile)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	cfg.Display.Theme = themeName
	if err = config.SaveConfig(configPath, cfg); err != nil {
		return err
	}

	pterm.Success.Println("Theme updated: " + themeName)
	return nil
}

func joinThemes(themes []string) string {
	if len(themes) == 0 {
		return ""
	}
	result := themes[0]
	for i := 1; i < len(themes); i++ {
		result += ", " + themes[i]
	}
	return result
}
