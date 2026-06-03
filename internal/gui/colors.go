package gui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
)

type Theme struct {
	CashWarriorTitle                string
	PTermInfoColor                  string
	PTermWarningColor               string
	PTermErrorColor                 string
	PTermSuccessColor               string
	PTermDebugColor                 string
	TableTitleText                  lipgloss.Color
	TableSubtitleText               lipgloss.Color
	TableHeaderText                 lipgloss.Color
	TableRowText                    lipgloss.Color
	TableAmountIncomeText           lipgloss.Color
	TableAmountExpenseText          lipgloss.Color
	TableAmountTransferText         lipgloss.Color
	SummaryTitleText                lipgloss.Color
	SummaryHeaderText               lipgloss.Color
	SummaryRowText                  lipgloss.Color
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
		CashWarriorTitle:                "#F8F8F2",
		TableTitleText:                  lipgloss.Color("#282A36"),
		TableSubtitleText:               lipgloss.Color("244"),
		TableHeaderText:                 lipgloss.Color("#282A36"),
		TableRowText:                    lipgloss.Color("#F8F8F2"),
		TableAmountIncomeText:           lipgloss.Color("#50FA7B"),
		TableAmountExpenseText:          lipgloss.Color("#FF5555"),
		TableAmountTransferText:         lipgloss.Color("#8BE9FD"),
		SummaryTitleText:                lipgloss.Color("#6272A4"),
		SummaryHeaderText:               lipgloss.Color("#8BE9FD"),
		SummaryRowText:                  lipgloss.Color("#F8F8F2"),
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
		CashWarriorTitle:                "#ECEFF4",
		TableTitleText:                  lipgloss.Color("#2E3440"),
		TableSubtitleText:               lipgloss.Color("#81A1C1"),
		TableHeaderText:                 lipgloss.Color("#2E3440"),
		TableRowText:                    lipgloss.Color("#E5E9F0"),
		TableAmountIncomeText:           lipgloss.Color("#A3BE8C"),
		TableAmountExpenseText:          lipgloss.Color("#BF616A"),
		TableAmountTransferText:         lipgloss.Color("#88C0D0"),
		SummaryTitleText:                lipgloss.Color("#ECEFF4"),
		SummaryHeaderText:               lipgloss.Color("#88C0D0"),
		SummaryRowText:                  lipgloss.Color("#E5E9F0"),
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
		CashWarriorTitle:                "#fdf6e3",
		TableTitleText:                  lipgloss.Color("#fdf6e3"),
		TableSubtitleText:               lipgloss.Color("#93a1a1"),
		TableHeaderText:                 lipgloss.Color("#fdf6e3"),
		TableRowText:                    lipgloss.Color("#eee8d5"),
		TableAmountIncomeText:           lipgloss.Color("#859900"),
		TableAmountExpenseText:          lipgloss.Color("#dc322f"),
		TableAmountTransferText:         lipgloss.Color("#268bd2"),
		SummaryTitleText:                lipgloss.Color("#fdf6e3"),
		SummaryHeaderText:               lipgloss.Color("#2aa198"),
		SummaryRowText:                  lipgloss.Color("#eee8d5"),
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
	"tokyonight": {
		CashWarriorTitle:                "#73DACA",
		TableTitleText:                  lipgloss.Color("#1a1b26"),
		TableSubtitleText:               lipgloss.Color("#a9b1d6"),
		TableHeaderText:                 lipgloss.Color("#1a1b26"),
		TableRowText:                    lipgloss.Color("#c0caf5"),
		TableAmountIncomeText:           lipgloss.Color("#9ece6a"),
		TableAmountExpenseText:          lipgloss.Color("#f7768e"),
		TableAmountTransferText:         lipgloss.Color("#7aa2f7"),
		SummaryTitleText:                lipgloss.Color("#e0af68"),
		SummaryHeaderText:               lipgloss.Color("#7dcfff"),
		SummaryRowText:                  lipgloss.Color("#c0caf5"),
		TableRowEvenBg:                  lipgloss.Color("#1a1b26"),
		TableRowOddBg:                   lipgloss.Color("#24283b"),
		TableBorder:                     lipgloss.Color("#414868"),
		AccountsTitleBackground:         lipgloss.Color("#bb9af7"),
		AccountsHeaderBackground:        lipgloss.Color("#7dcfff"),
		TransactionListTitleBackground:  lipgloss.Color("#9ece6a"),
		TransactionListHeaderBackground: lipgloss.Color("#e0af68"),
		CategoriesTitleBackground:       lipgloss.Color("#f7768e"),
		CategoriesHeaderBackground:      lipgloss.Color("#ff9e64"),
	},
	"gruvbox-dark": {
		CashWarriorTitle:                "#FABD2F",
		TableTitleText:                  lipgloss.Color("#1D2021"),
		TableSubtitleText:               lipgloss.Color("#A89984"),
		TableHeaderText:                 lipgloss.Color("#1D2021"),
		TableRowText:                    lipgloss.Color("#EBDBB2"),
		TableAmountIncomeText:           lipgloss.Color("#B8BB26"),
		TableAmountExpenseText:          lipgloss.Color("#FB4934"),
		TableAmountTransferText:         lipgloss.Color("#83A598"),
		SummaryTitleText:                lipgloss.Color("#D5C4A1"),
		SummaryHeaderText:               lipgloss.Color("#FE8019"),
		SummaryRowText:                  lipgloss.Color("#EBDBB2"),
		TableRowEvenBg:                  lipgloss.Color("#1D2021"),
		TableRowOddBg:                   lipgloss.Color("#282828"),
		TableBorder:                     lipgloss.Color("#665C54"),
		AccountsTitleBackground:         lipgloss.Color("#D3869B"),
		AccountsHeaderBackground:        lipgloss.Color("#8EC07C"),
		TransactionListTitleBackground:  lipgloss.Color("#B8BB26"),
		TransactionListHeaderBackground: lipgloss.Color("#FABD2F"),
		CategoriesTitleBackground:       lipgloss.Color("#FE8019"),
		CategoriesHeaderBackground:      lipgloss.Color("#FB4934"),
	},
	"catppuccin-mocha": {
		CashWarriorTitle:                "#89DCEB",
		TableTitleText:                  lipgloss.Color("#1E1E2E"),
		TableSubtitleText:               lipgloss.Color("#A6ADC8"),
		TableHeaderText:                 lipgloss.Color("#1E1E2E"),
		TableRowText:                    lipgloss.Color("#CDD6F4"),
		TableAmountIncomeText:           lipgloss.Color("#A6E3A1"),
		TableAmountExpenseText:          lipgloss.Color("#F38BA8"),
		TableAmountTransferText:         lipgloss.Color("#89B4FA"),
		SummaryTitleText:                lipgloss.Color("#F9E2AF"),
		SummaryHeaderText:               lipgloss.Color("#89DCEB"),
		SummaryRowText:                  lipgloss.Color("#CDD6F4"),
		TableRowEvenBg:                  lipgloss.Color("#1E1E2E"),
		TableRowOddBg:                   lipgloss.Color("#313244"),
		TableBorder:                     lipgloss.Color("#585B70"),
		AccountsTitleBackground:         lipgloss.Color("#CBA6F7"),
		AccountsHeaderBackground:        lipgloss.Color("#94E2D5"),
		TransactionListTitleBackground:  lipgloss.Color("#A6E3A1"),
		TransactionListHeaderBackground: lipgloss.Color("#F9E2AF"),
		CategoriesTitleBackground:       lipgloss.Color("#F38BA8"),
		CategoriesHeaderBackground:      lipgloss.Color("#FAB387"),
	},
	"kanagawa": {
		CashWarriorTitle:                "#7FB4CA",
		TableTitleText:                  lipgloss.Color("#1F1F28"),
		TableSubtitleText:               lipgloss.Color("#A6A69C"),
		TableHeaderText:                 lipgloss.Color("#1F1F28"),
		TableRowText:                    lipgloss.Color("#DCD7BA"),
		TableAmountIncomeText:           lipgloss.Color("#98BB6C"),
		TableAmountExpenseText:          lipgloss.Color("#E46876"),
		TableAmountTransferText:         lipgloss.Color("#7FB4CA"),
		SummaryTitleText:                lipgloss.Color("#E6C384"),
		SummaryHeaderText:               lipgloss.Color("#7AA89F"),
		SummaryRowText:                  lipgloss.Color("#DCD7BA"),
		TableRowEvenBg:                  lipgloss.Color("#1F1F28"),
		TableRowOddBg:                   lipgloss.Color("#2A2A37"),
		TableBorder:                     lipgloss.Color("#54546D"),
		AccountsTitleBackground:         lipgloss.Color("#957FB8"),
		AccountsHeaderBackground:        lipgloss.Color("#7AA89F"),
		TransactionListTitleBackground:  lipgloss.Color("#98BB6C"),
		TransactionListHeaderBackground: lipgloss.Color("#E6C384"),
		CategoriesTitleBackground:       lipgloss.Color("#E46876"),
		CategoriesHeaderBackground:      lipgloss.Color("#FFA066"),
	},
	"onedark": {
		CashWarriorTitle:                "#61AFEF",
		TableTitleText:                  lipgloss.Color("#282C34"),
		TableSubtitleText:               lipgloss.Color("#ABB2BF"),
		TableHeaderText:                 lipgloss.Color("#282C34"),
		TableRowText:                    lipgloss.Color("#DCDFE4"),
		TableAmountIncomeText:           lipgloss.Color("#98C379"),
		TableAmountExpenseText:          lipgloss.Color("#E06C75"),
		TableAmountTransferText:         lipgloss.Color("#61AFEF"),
		SummaryTitleText:                lipgloss.Color("#E5C07B"),
		SummaryHeaderText:               lipgloss.Color("#56B6C2"),
		SummaryRowText:                  lipgloss.Color("#DCDFE4"),
		TableRowEvenBg:                  lipgloss.Color("#282C34"),
		TableRowOddBg:                   lipgloss.Color("#313640"),
		TableBorder:                     lipgloss.Color("#4B5263"),
		AccountsTitleBackground:         lipgloss.Color("#C678DD"),
		AccountsHeaderBackground:        lipgloss.Color("#56B6C2"),
		TransactionListTitleBackground:  lipgloss.Color("#98C379"),
		TransactionListHeaderBackground: lipgloss.Color("#E5C07B"),
		CategoriesTitleBackground:       lipgloss.Color("#E06C75"),
		CategoriesHeaderBackground:      lipgloss.Color("#D19A66"),
	},
	"rose-pine": {
		CashWarriorTitle:                "#9CCFD8",
		TableTitleText:                  lipgloss.Color("#191724"),
		TableSubtitleText:               lipgloss.Color("#908CAA"),
		TableHeaderText:                 lipgloss.Color("#191724"),
		TableRowText:                    lipgloss.Color("#E0DEF4"),
		TableAmountIncomeText:           lipgloss.Color("#31748F"),
		TableAmountExpenseText:          lipgloss.Color("#EB6F92"),
		TableAmountTransferText:         lipgloss.Color("#9CCFD8"),
		SummaryTitleText:                lipgloss.Color("#F6C177"),
		SummaryHeaderText:               lipgloss.Color("#9CCFD8"),
		SummaryRowText:                  lipgloss.Color("#E0DEF4"),
		TableRowEvenBg:                  lipgloss.Color("#191724"),
		TableRowOddBg:                   lipgloss.Color("#232136"),
		TableBorder:                     lipgloss.Color("#524F67"),
		AccountsTitleBackground:         lipgloss.Color("#C4A7E7"),
		AccountsHeaderBackground:        lipgloss.Color("#9CCFD8"),
		TransactionListTitleBackground:  lipgloss.Color("#31748F"),
		TransactionListHeaderBackground: lipgloss.Color("#F6C177"),
		CategoriesTitleBackground:       lipgloss.Color("#EB6F92"),
		CategoriesHeaderBackground:      lipgloss.Color("#EA9A97"),
	},
	"everforest-dark": {
		CashWarriorTitle:                "#83C092",
		TableTitleText:                  lipgloss.Color("#2D353B"),
		TableSubtitleText:               lipgloss.Color("#A7C080"),
		TableHeaderText:                 lipgloss.Color("#2D353B"),
		TableRowText:                    lipgloss.Color("#D3C6AA"),
		TableAmountIncomeText:           lipgloss.Color("#A7C080"),
		TableAmountExpenseText:          lipgloss.Color("#E67E80"),
		TableAmountTransferText:         lipgloss.Color("#7FBBB3"),
		SummaryTitleText:                lipgloss.Color("#DBBC7F"),
		SummaryHeaderText:               lipgloss.Color("#83C092"),
		SummaryRowText:                  lipgloss.Color("#D3C6AA"),
		TableRowEvenBg:                  lipgloss.Color("#2D353B"),
		TableRowOddBg:                   lipgloss.Color("#343F44"),
		TableBorder:                     lipgloss.Color("#5C6A72"),
		AccountsTitleBackground:         lipgloss.Color("#D699B6"),
		AccountsHeaderBackground:        lipgloss.Color("#7FBBB3"),
		TransactionListTitleBackground:  lipgloss.Color("#A7C080"),
		TransactionListHeaderBackground: lipgloss.Color("#DBBC7F"),
		CategoriesTitleBackground:       lipgloss.Color("#E67E80"),
		CategoriesHeaderBackground:      lipgloss.Color("#E69875"),
	},
	"paper-ink": {
		CashWarriorTitle:                "#1F2937",
		TableTitleText:                  lipgloss.Color("#111827"),
		TableSubtitleText:               lipgloss.Color("#6B7280"),
		TableHeaderText:                 lipgloss.Color("#111827"),
		TableRowText:                    lipgloss.Color("#1F2937"),
		TableAmountIncomeText:           lipgloss.Color("#047857"),
		TableAmountExpenseText:          lipgloss.Color("#B91C1C"),
		TableAmountTransferText:         lipgloss.Color("#1D4ED8"),
		SummaryTitleText:                lipgloss.Color("#374151"),
		SummaryHeaderText:               lipgloss.Color("#2563EB"),
		SummaryRowText:                  lipgloss.Color("#1F2937"),
		TableRowEvenBg:                  lipgloss.Color("#FFFFFF"),
		TableRowOddBg:                   lipgloss.Color("#F3F4F6"),
		TableBorder:                     lipgloss.Color("#D1D5DB"),
		AccountsTitleBackground:         lipgloss.Color("#93C5FD"),
		AccountsHeaderBackground:        lipgloss.Color("#BFDBFE"),
		TransactionListTitleBackground:  lipgloss.Color("#A7F3D0"),
		TransactionListHeaderBackground: lipgloss.Color("#C7D2FE"),
		CategoriesTitleBackground:       lipgloss.Color("#FBCFE8"),
		CategoriesHeaderBackground:      lipgloss.Color("#FDE68A"),
	},
	"aurora": {
		CashWarriorTitle:                "#7DF9FF",
		TableTitleText:                  lipgloss.Color("#081229"),
		TableSubtitleText:               lipgloss.Color("#7AA2C6"),
		TableHeaderText:                 lipgloss.Color("#081229"),
		TableRowText:                    lipgloss.Color("#D6EDFF"),
		TableAmountIncomeText:           lipgloss.Color("#7CFFCB"),
		TableAmountExpenseText:          lipgloss.Color("#FF7DAF"),
		TableAmountTransferText:         lipgloss.Color("#7DF9FF"),
		SummaryTitleText:                lipgloss.Color("#9BE7FF"),
		SummaryHeaderText:               lipgloss.Color("#77FFD9"),
		SummaryRowText:                  lipgloss.Color("#D6EDFF"),
		TableRowEvenBg:                  lipgloss.Color("#081229"),
		TableRowOddBg:                   lipgloss.Color("#0F1E3A"),
		TableBorder:                     lipgloss.Color("#2A4B73"),
		AccountsTitleBackground:         lipgloss.Color("#7C83FF"),
		AccountsHeaderBackground:        lipgloss.Color("#64D2FF"),
		TransactionListTitleBackground:  lipgloss.Color("#4CFFB2"),
		TransactionListHeaderBackground: lipgloss.Color("#8CEBFF"),
		CategoriesTitleBackground:       lipgloss.Color("#FF7DAF"),
		CategoriesHeaderBackground:      lipgloss.Color("#FFC56E"),
	},
	"acid-pop": {
		CashWarriorTitle:                "#39FF14",
		TableTitleText:                  lipgloss.Color("#11001C"),
		TableSubtitleText:               lipgloss.Color("#FF9CF7"),
		TableHeaderText:                 lipgloss.Color("#11001C"),
		TableRowText:                    lipgloss.Color("#F8F4FF"),
		TableAmountIncomeText:           lipgloss.Color("#39FF14"),
		TableAmountExpenseText:          lipgloss.Color("#FF2E88"),
		TableAmountTransferText:         lipgloss.Color("#00E5FF"),
		SummaryTitleText:                lipgloss.Color("#FFE66D"),
		SummaryHeaderText:               lipgloss.Color("#00E5FF"),
		SummaryRowText:                  lipgloss.Color("#F8F4FF"),
		TableRowEvenBg:                  lipgloss.Color("#11001C"),
		TableRowOddBg:                   lipgloss.Color("#1A0030"),
		TableBorder:                     lipgloss.Color("#7A1CAC"),
		AccountsTitleBackground:         lipgloss.Color("#FF2E88"),
		AccountsHeaderBackground:        lipgloss.Color("#00E5FF"),
		TransactionListTitleBackground:  lipgloss.Color("#39FF14"),
		TransactionListHeaderBackground: lipgloss.Color("#FFE66D"),
		CategoriesTitleBackground:       lipgloss.Color("#B026FF"),
		CategoriesHeaderBackground:      lipgloss.Color("#FF7A00"),
	},
	"terminal-mono": {
		CashWarriorTitle:                "#C0C0C0",
		TableTitleText:                  lipgloss.Color("#000000"),
		TableSubtitleText:               lipgloss.Color("#808080"),
		TableHeaderText:                 lipgloss.Color("#000000"),
		TableRowText:                    lipgloss.Color("#C0C0C0"),
		TableAmountIncomeText:           lipgloss.Color("#00FF00"),
		TableAmountExpenseText:          lipgloss.Color("#FF0000"),
		TableAmountTransferText:         lipgloss.Color("#00B7FF"),
		SummaryTitleText:                lipgloss.Color("#C0C0C0"),
		SummaryHeaderText:               lipgloss.Color("#A0A0A0"),
		SummaryRowText:                  lipgloss.Color("#C0C0C0"),
		TableRowEvenBg:                  lipgloss.Color("#000000"),
		TableRowOddBg:                   lipgloss.Color("#111111"),
		TableBorder:                     lipgloss.Color("#666666"),
		AccountsTitleBackground:         lipgloss.Color("#333333"),
		AccountsHeaderBackground:        lipgloss.Color("#4D4D4D"),
		TransactionListTitleBackground:  lipgloss.Color("#2F2F2F"),
		TransactionListHeaderBackground: lipgloss.Color("#5A5A5A"),
		CategoriesTitleBackground:       lipgloss.Color("#3A3A3A"),
		CategoriesHeaderBackground:      lipgloss.Color("#5F5F5F"),
	},
	"neon-noir": {
		CashWarriorTitle:                "#00F5FF",
		TableTitleText:                  lipgloss.Color("#07070D"),
		TableSubtitleText:               lipgloss.Color("#A88BFF"),
		TableHeaderText:                 lipgloss.Color("#07070D"),
		TableRowText:                    lipgloss.Color("#E8F7FF"),
		TableAmountIncomeText:           lipgloss.Color("#39FF88"),
		TableAmountExpenseText:          lipgloss.Color("#FF4DA6"),
		TableAmountTransferText:         lipgloss.Color("#00F5FF"),
		SummaryTitleText:                lipgloss.Color("#FFEA00"),
		SummaryHeaderText:               lipgloss.Color("#00F5FF"),
		SummaryRowText:                  lipgloss.Color("#E8F7FF"),
		TableRowEvenBg:                  lipgloss.Color("#07070D"),
		TableRowOddBg:                   lipgloss.Color("#101021"),
		TableBorder:                     lipgloss.Color("#532B8C"),
		AccountsTitleBackground:         lipgloss.Color("#FF00CC"),
		AccountsHeaderBackground:        lipgloss.Color("#00F5FF"),
		TransactionListTitleBackground:  lipgloss.Color("#39FF88"),
		TransactionListHeaderBackground: lipgloss.Color("#FFEA00"),
		CategoriesTitleBackground:       lipgloss.Color("#FF4DA6"),
		CategoriesHeaderBackground:      lipgloss.Color("#FF8A00"),
	},
	"retro-arcade": {
		CashWarriorTitle:                "#6DF2FF",
		TableTitleText:                  lipgloss.Color("#120458"),
		TableSubtitleText:               lipgloss.Color("#B6A5FF"),
		TableHeaderText:                 lipgloss.Color("#120458"),
		TableRowText:                    lipgloss.Color("#E9E6FF"),
		TableAmountIncomeText:           lipgloss.Color("#5CFF5C"),
		TableAmountExpenseText:          lipgloss.Color("#FF5C8A"),
		TableAmountTransferText:         lipgloss.Color("#6DF2FF"),
		SummaryTitleText:                lipgloss.Color("#FFD166"),
		SummaryHeaderText:               lipgloss.Color("#6DF2FF"),
		SummaryRowText:                  lipgloss.Color("#E9E6FF"),
		TableRowEvenBg:                  lipgloss.Color("#120458"),
		TableRowOddBg:                   lipgloss.Color("#1F0A72"),
		TableBorder:                     lipgloss.Color("#7A5CFF"),
		AccountsTitleBackground:         lipgloss.Color("#FF4ECD"),
		AccountsHeaderBackground:        lipgloss.Color("#6DF2FF"),
		TransactionListTitleBackground:  lipgloss.Color("#5CFF5C"),
		TransactionListHeaderBackground: lipgloss.Color("#FFD166"),
		CategoriesTitleBackground:       lipgloss.Color("#FF5C8A"),
		CategoriesHeaderBackground:      lipgloss.Color("#FF9F43"),
	},
	"fog": {
		CashWarriorTitle:                "#C7D0D9",
		TableTitleText:                  lipgloss.Color("#20242B"),
		TableSubtitleText:               lipgloss.Color("#8C98A5"),
		TableHeaderText:                 lipgloss.Color("#20242B"),
		TableRowText:                    lipgloss.Color("#D8E0E8"),
		TableAmountIncomeText:           lipgloss.Color("#8ED1B2"),
		TableAmountExpenseText:          lipgloss.Color("#E7A7A7"),
		TableAmountTransferText:         lipgloss.Color("#9EC3E6"),
		SummaryTitleText:                lipgloss.Color("#C7D0D9"),
		SummaryHeaderText:               lipgloss.Color("#AFC2D6"),
		SummaryRowText:                  lipgloss.Color("#D8E0E8"),
		TableRowEvenBg:                  lipgloss.Color("#20242B"),
		TableRowOddBg:                   lipgloss.Color("#2A3038"),
		TableBorder:                     lipgloss.Color("#5A6673"),
		AccountsTitleBackground:         lipgloss.Color("#92A8C0"),
		AccountsHeaderBackground:        lipgloss.Color("#AFC2D6"),
		TransactionListTitleBackground:  lipgloss.Color("#8EB8A8"),
		TransactionListHeaderBackground: lipgloss.Color("#C9BDAE"),
		CategoriesTitleBackground:       lipgloss.Color("#C3A7B8"),
		CategoriesHeaderBackground:      lipgloss.Color("#B9B3A9"),
	},
	"sunset-drive": {
		CashWarriorTitle:                "#FFB86B",
		TableTitleText:                  lipgloss.Color("#2A103A"),
		TableSubtitleText:               lipgloss.Color("#D2A9C8"),
		TableHeaderText:                 lipgloss.Color("#2A103A"),
		TableRowText:                    lipgloss.Color("#FFE7D1"),
		TableAmountIncomeText:           lipgloss.Color("#FFD166"),
		TableAmountExpenseText:          lipgloss.Color("#FF6B6B"),
		TableAmountTransferText:         lipgloss.Color("#70D6FF"),
		SummaryTitleText:                lipgloss.Color("#FFB86B"),
		SummaryHeaderText:               lipgloss.Color("#70D6FF"),
		SummaryRowText:                  lipgloss.Color("#FFE7D1"),
		TableRowEvenBg:                  lipgloss.Color("#2A103A"),
		TableRowOddBg:                   lipgloss.Color("#3A1A4D"),
		TableBorder:                     lipgloss.Color("#7D4E88"),
		AccountsTitleBackground:         lipgloss.Color("#FF8FA3"),
		AccountsHeaderBackground:        lipgloss.Color("#70D6FF"),
		TransactionListTitleBackground:  lipgloss.Color("#FFD166"),
		TransactionListHeaderBackground: lipgloss.Color("#FFB86B"),
		CategoriesTitleBackground:       lipgloss.Color("#FF6B6B"),
		CategoriesHeaderBackground:      lipgloss.Color("#F4A261"),
	},
	"toxic-waste": {
		CashWarriorTitle:                "#C6FF00",
		TableTitleText:                  lipgloss.Color("#0A0A0A"),
		TableSubtitleText:               lipgloss.Color("#9ACD32"),
		TableHeaderText:                 lipgloss.Color("#0A0A0A"),
		TableRowText:                    lipgloss.Color("#E8FFB0"),
		TableAmountIncomeText:           lipgloss.Color("#7CFC00"),
		TableAmountExpenseText:          lipgloss.Color("#FF3B30"),
		TableAmountTransferText:         lipgloss.Color("#E5FF00"),
		SummaryTitleText:                lipgloss.Color("#C6FF00"),
		SummaryHeaderText:               lipgloss.Color("#E5FF00"),
		SummaryRowText:                  lipgloss.Color("#E8FFB0"),
		TableRowEvenBg:                  lipgloss.Color("#0A0A0A"),
		TableRowOddBg:                   lipgloss.Color("#151A00"),
		TableBorder:                     lipgloss.Color("#5A6B00"),
		AccountsTitleBackground:         lipgloss.Color("#9ACD32"),
		AccountsHeaderBackground:        lipgloss.Color("#C6FF00"),
		TransactionListTitleBackground:  lipgloss.Color("#7CFC00"),
		TransactionListHeaderBackground: lipgloss.Color("#E5FF00"),
		CategoriesTitleBackground:       lipgloss.Color("#B9FF00"),
		CategoriesHeaderBackground:      lipgloss.Color("#E0FF4F"),
	},
	"contrast-max": {
		CashWarriorTitle:                "#FFFFFF",
		TableTitleText:                  lipgloss.Color("#000000"),
		TableSubtitleText:               lipgloss.Color("#FFFFFF"),
		TableHeaderText:                 lipgloss.Color("#000000"),
		TableRowText:                    lipgloss.Color("#FFFFFF"),
		TableAmountIncomeText:           lipgloss.Color("#FFFFFF"),
		TableAmountExpenseText:          lipgloss.Color("#FFFFFF"),
		TableAmountTransferText:         lipgloss.Color("#FFFFFF"),
		SummaryTitleText:                lipgloss.Color("#FFFFFF"),
		SummaryHeaderText:               lipgloss.Color("#FFFFFF"),
		SummaryRowText:                  lipgloss.Color("#FFFFFF"),
		TableRowEvenBg:                  lipgloss.Color("#000000"),
		TableRowOddBg:                   lipgloss.Color("#000000"),
		TableBorder:                     lipgloss.Color("#FFFFFF"),
		AccountsTitleBackground:         lipgloss.Color("#FFFFFF"),
		AccountsHeaderBackground:        lipgloss.Color("#FFFFFF"),
		TransactionListTitleBackground:  lipgloss.Color("#FFFFFF"),
		TransactionListHeaderBackground: lipgloss.Color("#FFFFFF"),
		CategoriesTitleBackground:       lipgloss.Color("#FFFFFF"),
		CategoriesHeaderBackground:      lipgloss.Color("#FFFFFF"),
	},
	"kenu": {
		CashWarriorTitle:                "#9BE99B",
		TableTitleText:                  lipgloss.Color("#0A1A0A"),
		TableSubtitleText:               lipgloss.Color("#6FAF6F"),
		TableHeaderText:                 lipgloss.Color("#0A1A0A"),
		TableRowText:                    lipgloss.Color("#B7DDB7"),
		TableAmountIncomeText:           lipgloss.Color("#9BE99B"),
		TableAmountExpenseText:          lipgloss.Color("#7FBF7F"),
		TableAmountTransferText:         lipgloss.Color("#A8D8A8"),
		SummaryTitleText:                lipgloss.Color("#9BE99B"),
		SummaryHeaderText:               lipgloss.Color("#7FD67F"),
		SummaryRowText:                  lipgloss.Color("#B7DDB7"),
		TableRowEvenBg:                  lipgloss.Color("#0A1A0A"),
		TableRowOddBg:                   lipgloss.Color("#102510"),
		TableBorder:                     lipgloss.Color("#4F7F4F"),
		AccountsTitleBackground:         lipgloss.Color("#5A8F5A"),
		AccountsHeaderBackground:        lipgloss.Color("#6FAF6F"),
		TransactionListTitleBackground:  lipgloss.Color("#7FD67F"),
		TransactionListHeaderBackground: lipgloss.Color("#9BE99B"),
		CategoriesTitleBackground:       lipgloss.Color("#5F9F5F"),
		CategoriesHeaderBackground:      lipgloss.Color("#76B676"),
	},
	"blue-kenu": {
		CashWarriorTitle:                "#A7D8FF",
		TableTitleText:                  lipgloss.Color("#07121F"),
		TableSubtitleText:               lipgloss.Color("#6FA7D9"),
		TableHeaderText:                 lipgloss.Color("#07121F"),
		TableRowText:                    lipgloss.Color("#CFE8FF"),
		TableAmountIncomeText:           lipgloss.Color("#BFE1FF"),
		TableAmountExpenseText:          lipgloss.Color("#7FB8E6"),
		TableAmountTransferText:         lipgloss.Color("#A7D8FF"),
		SummaryTitleText:                lipgloss.Color("#A7D8FF"),
		SummaryHeaderText:               lipgloss.Color("#8BC8FF"),
		SummaryRowText:                  lipgloss.Color("#CFE8FF"),
		TableRowEvenBg:                  lipgloss.Color("#07121F"),
		TableRowOddBg:                   lipgloss.Color("#0C1D31"),
		TableBorder:                     lipgloss.Color("#4E6F8E"),
		AccountsTitleBackground:         lipgloss.Color("#5A86B3"),
		AccountsHeaderBackground:        lipgloss.Color("#6FA7D9"),
		TransactionListTitleBackground:  lipgloss.Color("#8BC8FF"),
		TransactionListHeaderBackground: lipgloss.Color("#A7D8FF"),
		CategoriesTitleBackground:       lipgloss.Color("#5F91C1"),
		CategoriesHeaderBackground:      lipgloss.Color("#79ADDC"),
	},
}

var activeThemeName = "dracula"

func init() {
	for name, theme := range themes {
		themes[name] = normalizeTheme(theme)
	}
}

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
		return normalizeTheme(theme)
	}
	return normalizeTheme(themes["dracula"])
}

func normalizeTheme(theme Theme) Theme {
	if theme.PTermInfoColor == "" {
		theme.PTermInfoColor = string(theme.SummaryHeaderText)
	}
	if theme.PTermWarningColor == "" {
		theme.PTermWarningColor = string(theme.TransactionListHeaderBackground)
	}
	if theme.PTermErrorColor == "" {
		theme.PTermErrorColor = string(theme.TableAmountExpenseText)
	}
	if theme.PTermSuccessColor == "" {
		theme.PTermSuccessColor = string(theme.TableAmountIncomeText)
	}
	if theme.PTermDebugColor == "" {
		theme.PTermDebugColor = string(theme.TableBorder)
	}
	return theme
}

func ApplyPTermTheme(theme Theme) {
	applyMessageStyle := func(color string, setter func(style *pterm.Style)) {
		c, err := HexToANSIColor(color)
		if err != nil {
			return
		}
		setter(pterm.NewStyle(c))
	}

	applyPrefix := func(color string, label string, setter func(text string)) {
		rgb, err := HexToRGB(color)
		if err != nil {
			return
		}
		setter(rgb.Sprint(label))
	}

	applyPrefix(theme.PTermInfoColor, "INFO", func(text string) {
		pterm.Info.Prefix.Text = text
		pterm.Info.Prefix.Style = pterm.NewStyle()
	})
	applyMessageStyle(theme.PTermInfoColor, func(style *pterm.Style) {
		pterm.Info.MessageStyle = style
	})
	applyPrefix(theme.PTermWarningColor, "WARN", func(text string) {
		pterm.Warning.Prefix.Text = text
		pterm.Warning.Prefix.Style = pterm.NewStyle()
	})
	applyMessageStyle(theme.PTermWarningColor, func(style *pterm.Style) {
		pterm.Warning.MessageStyle = style
	})
	applyPrefix(theme.PTermErrorColor, "ERROR", func(text string) {
		pterm.Error.Prefix.Text = text
		pterm.Error.Prefix.Style = pterm.NewStyle()
	})
	applyMessageStyle(theme.PTermErrorColor, func(style *pterm.Style) {
		pterm.Error.MessageStyle = style
	})
	applyPrefix(theme.PTermSuccessColor, "SUCCESS", func(text string) {
		pterm.Success.Prefix.Text = text
		pterm.Success.Prefix.Style = pterm.NewStyle()
	})
	applyMessageStyle(theme.PTermSuccessColor, func(style *pterm.Style) {
		pterm.Success.MessageStyle = style
	})
	applyPrefix(theme.PTermDebugColor, "DEBUG", func(text string) {
		pterm.Debug.Prefix.Text = text
		pterm.Debug.Prefix.Style = pterm.NewStyle()
	})
	applyMessageStyle(theme.PTermDebugColor, func(style *pterm.Style) {
		pterm.Debug.MessageStyle = style
	})
}

func HexToANSIColor(hex string) (pterm.Color, error) {
	rgb, err := HexToRGB(hex)
	if err != nil {
		return pterm.FgDefault, err
	}

	palette := []struct {
		color pterm.Color
		r     float64
		g     float64
		b     float64
	}{
		{pterm.FgBlack, 0, 0, 0},
		{pterm.FgRed, 205, 49, 49},
		{pterm.FgGreen, 13, 188, 121},
		{pterm.FgYellow, 229, 229, 16},
		{pterm.FgBlue, 36, 114, 200},
		{pterm.FgMagenta, 188, 63, 188},
		{pterm.FgCyan, 17, 168, 205},
		{pterm.FgLightWhite, 229, 229, 229},
		{pterm.FgDarkGray, 102, 102, 102},
		{pterm.FgLightRed, 241, 76, 76},
		{pterm.FgLightGreen, 35, 209, 139},
		{pterm.FgLightYellow, 245, 245, 67},
		{pterm.FgLightBlue, 59, 142, 234},
		{pterm.FgLightMagenta, 214, 112, 214},
		{pterm.FgLightCyan, 41, 184, 219},
		{pterm.FgWhite, 255, 255, 255},
	}

	best := pterm.FgWhite
	bestDist := math.MaxFloat64
	for _, p := range palette {
		dr := float64(rgb.R) - p.r
		dg := float64(rgb.G) - p.g
		db := float64(rgb.B) - p.b
		d := dr*dr + dg*dg + db*db
		if d < bestDist {
			bestDist = d
			best = p.color
		}
	}

	return best, nil
}

func ThemeExists(name string) bool {
	_, ok := themes[name]
	return ok
}

func ThemeNames() []string {
	names := make([]string, 0, len(themes))
	for name := range themes {
		names = append(names, name)
	}
	return names
}

func HexToRGB(hex string) (pterm.RGB, error) {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return pterm.RGB{}, fmt.Errorf("invalid hex color")
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return pterm.RGB{}, err
	}

	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return pterm.RGB{}, err
	}

	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return pterm.RGB{}, err
	}

	return pterm.NewRGB(uint8(r), uint8(g), uint8(b)), nil
}
