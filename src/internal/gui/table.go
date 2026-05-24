package gui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Table struct {
	title           string
	subtitle        string
	titleBgColor    lipgloss.Color
	headerBgColor   lipgloss.Color
	headers         []string
	rows            [][]string
	subtitleColor   lipgloss.Color
	foregroundColor lipgloss.Color
	backgroundColor lipgloss.Color
	currentRowColor lipgloss.Color
	commentColor    lipgloss.Color
	titleTextColor  lipgloss.Color
	headerTextColor lipgloss.Color
	marginLeft      int
	marginBottom    int
}

func NewTable() *Table {
	theme := CurrentTheme()
	return &Table{
		titleBgColor:    theme.AccountsTitleBackground,
		headerBgColor:   theme.AccountsHeaderBackground,
		subtitleColor:   theme.TableSubtitleText,
		foregroundColor: theme.TableRowText,
		backgroundColor: theme.TableRowEvenBg,
		currentRowColor: theme.TableRowOddBg,
		commentColor:    theme.TableBorder,
		titleTextColor:  theme.TableTitleText,
		headerTextColor: theme.TableHeaderText,
		headers:         make([]string, 0),
		rows:            make([][]string, 0),
		marginLeft:      4,
		marginBottom:    2,
	}
}

func (t *Table) WithTitle(title string, bgColor lipgloss.Color) *Table {
	t.title = title
	t.titleBgColor = bgColor
	return t
}

func (t *Table) WithSubtitle(subtitle string) *Table {
	t.subtitle = subtitle
	return t
}

func (t *Table) WithHeaders(headers ...string) *Table {
	t.headers = headers
	return t
}

func (t *Table) WithHeaderBackground(color lipgloss.Color) *Table {
	t.headerBgColor = color
	return t
}

func (t *Table) AddRow(row ...string) *Table {
	t.rows = append(t.rows, row)
	return t
}

func (t *Table) AddRows(rows [][]string) *Table {
	t.rows = append(t.rows, rows...)
	return t
}

func (t *Table) WithMarginLeft(margin int) *Table {
	t.marginLeft = margin
	return t
}

func (t *Table) WithMarginBottom(margin int) *Table {
	t.marginBottom = margin
	return t
}

func (t *Table) Render() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(t.titleTextColor).
		Background(t.titleBgColor).
		Padding(0, 3).
		MarginTop(2)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(t.subtitleColor).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(t.headerTextColor).
		Background(t.headerBgColor).
		Padding(0, 1)

	evenRowStyle := lipgloss.NewStyle().
		Foreground(t.foregroundColor).
		Background(t.backgroundColor).
		Padding(0, 1)

	oddRowStyle := lipgloss.NewStyle().
		Foreground(t.foregroundColor).
		Background(t.currentRowColor).
		Padding(0, 1)

	borderStyle := lipgloss.NewStyle().
		Foreground(t.commentColor)

	renderedTable := table.New().
		Border(lipgloss.Border{
			Top:    "─",
			Bottom: "─",
		}).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false).
		BorderHeader(false).
		BorderRow(false).
		BorderStyle(borderStyle).
		Headers(t.headers...).
		Rows(t.rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			if row%2 == 0 {
				return evenRowStyle
			}

			return oddRowStyle
		}).
		Render()

	output := ""
	if t.title != "" {
		output += titleStyle.Render(t.title) + "\n"
	}
	if t.subtitle != "" {
		output += subtitleStyle.Render(t.subtitle) + "\n"
	}
	output += renderedTable

	return lipgloss.NewStyle().
		MarginLeft(t.marginLeft).
		MarginBottom(t.marginBottom).
		Render(output)
}

func (t *Table) Print() {
	fmt.Println(t.Render())
}
