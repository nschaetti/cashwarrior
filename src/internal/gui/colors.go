package gui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	TableTitleText                  lipgloss.Color
	TableSubtitleText               lipgloss.Color
	TableHeaderText                 lipgloss.Color
	TableRowText                    lipgloss.Color
	TableRowEvenBg                  lipgloss.Color
	TableRowOddBg                   lipgloss.Color
	TableBorder                     lipgloss.Color
	AccountsTitleBackground         lipgloss.Color
	AccountsHeaderBackground        lipgloss.Color
	TransactionListTitleBackground  lipgloss.Color
	TransactionListHeaderBackground lipgloss.Color
	CategoriesTitleBackground       lipgloss.Color
	CategoriesHeaderBackground      lipgloss.Color
}

var themes = map[string]Theme{
	"dracula": {
		TableTitleText:                  lipgloss.Color("#282A36"),
		TableSubtitleText:               lipgloss.Color("244"),
		TableHeaderText:                 lipgloss.Color("#282A36"),
		TableRowText:                    lipgloss.Color("#F8F8F2"),
		TableRowEvenBg:                  lipgloss.Color("#282A36"),
		TableRowOddBg:                   lipgloss.Color("#44475A"),
		TableBorder:                     lipgloss.Color("#6272A4"),
		AccountsTitleBackground:         lipgloss.Color("#BD93F9"),
		AccountsHeaderBackground:        lipgloss.Color("#8BE9FD"),
		TransactionListTitleBackground:  lipgloss.Color("#50FA7B"),
		TransactionListHeaderBackground: lipgloss.Color("#F1FA8C"),
		CategoriesTitleBackground:       lipgloss.Color("#FF79C6"),
		CategoriesHeaderBackground:      lipgloss.Color("#FF5555"),
	},
	"nord": {
		TableTitleText:                  lipgloss.Color("#ECEFF4"),
		TableSubtitleText:               lipgloss.Color("#81A1C1"),
		TableHeaderText:                 lipgloss.Color("#ECEFF4"),
		TableRowText:                    lipgloss.Color("#E5E9F0"),
		TableRowEvenBg:                  lipgloss.Color("#2E3440"),
		TableRowOddBg:                   lipgloss.Color("#3B4252"),
		TableBorder:                     lipgloss.Color("#4C566A"),
		AccountsTitleBackground:         lipgloss.Color("#5E81AC"),
		AccountsHeaderBackground:        lipgloss.Color("#88C0D0"),
		TransactionListTitleBackground:  lipgloss.Color("#A3BE8C"),
		TransactionListHeaderBackground: lipgloss.Color("#EBCB8B"),
		CategoriesTitleBackground:       lipgloss.Color("#BF616A"),
		CategoriesHeaderBackground:      lipgloss.Color("#D08770"),
	},
	"solarized": {
		TableTitleText:                  lipgloss.Color("#fdf6e3"),
		TableSubtitleText:               lipgloss.Color("#93a1a1"),
		TableHeaderText:                 lipgloss.Color("#fdf6e3"),
		TableRowText:                    lipgloss.Color("#eee8d5"),
		TableRowEvenBg:                  lipgloss.Color("#002b36"),
		TableRowOddBg:                   lipgloss.Color("#073642"),
		TableBorder:                     lipgloss.Color("#586e75"),
		AccountsTitleBackground:         lipgloss.Color("#6c71c4"),
		AccountsHeaderBackground:        lipgloss.Color("#2aa198"),
		TransactionListTitleBackground:  lipgloss.Color("#859900"),
		TransactionListHeaderBackground: lipgloss.Color("#b58900"),
		CategoriesTitleBackground:       lipgloss.Color("#cb4b16"),
		CategoriesHeaderBackground:      lipgloss.Color("#dc322f"),
	},
}

var activeThemeName = "dracula"

func SetTheme(name string) {
	if _, ok := themes[name]; ok {
		activeThemeName = name
		return
	}
	activeThemeName = "dracula"
}

func CurrentTheme() Theme {
	theme, ok := themes[activeThemeName]
	if ok {
		return theme
	}
	return themes["dracula"]
}
