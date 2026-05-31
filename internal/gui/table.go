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
	rowMetadata     []map[string]string
	subtitleColor   lipgloss.Color
	foregroundColor lipgloss.Color
	backgroundColor lipgloss.Color
	currentRowColor lipgloss.Color
	commentColor    lipgloss.Color
	titleTextColor  lipgloss.Color
	headerTextColor lipgloss.Color
	marginLeft      int
	marginBottom    int
	typeName        string
}

const (
	TableTypeDefault = "default"
	TableTypeSummary = "summary"
)

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
		rowMetadata:     make([]map[string]string, 0),
		marginLeft:      1,
		marginBottom:    0,
		typeName:        TableTypeDefault,
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
	t.rowMetadata = append(t.rowMetadata, nil)
	return t
}

func (t *Table) AddRowWithMetadata(row []string, metadata map[string]string) *Table {
	t.rows = append(t.rows, row)
	t.rowMetadata = append(t.rowMetadata, metadata)
	return t
}

func (t *Table) AddRows(rows [][]string) *Table {
	for _, row := range rows {
		t.AddRow(row...)
	}
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

func (t *Table) WithType(typeName string) *Table {
	t.typeName = typeName
	return t
}

func (t *Table) Render() string {
	theme := CurrentTheme()
	titleTextColor := t.titleTextColor
	headerTextColor := t.headerTextColor
	rowTextColor := t.foregroundColor
	if t.typeName == TableTypeSummary {
		titleTextColor = theme.SummaryTitleText
		headerTextColor = theme.SummaryHeaderText
		rowTextColor = theme.SummaryRowText
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleTextColor).
		Padding(0, 0)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(t.subtitleColor).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(headerTextColor).
		Padding(0, 1)

	evenRowStyle := lipgloss.NewStyle().
		Foreground(rowTextColor).
		Padding(0, 1)

	oddRowStyle := lipgloss.NewStyle().
		Foreground(rowTextColor).
		Padding(0, 1)

	if t.typeName != TableTypeSummary {
		titleStyle = titleStyle.Background(t.titleBgColor).Padding(0, 3).MarginTop(2)
		headerStyle = headerStyle.Background(t.headerBgColor)
		evenRowStyle = evenRowStyle.Background(t.backgroundColor)
		oddRowStyle = oddRowStyle.Background(t.currentRowColor)
	}

	borderStyle := lipgloss.NewStyle().
		Foreground(t.commentColor)

	amountCol := -1
	for i, header := range t.headers {
		if header == "Amount" {
			amountCol = i
			break
		}
	}

	styleForCell := func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return headerStyle
		}

		baseStyle := oddRowStyle
		if row%2 == 0 {
			baseStyle = evenRowStyle
		}

		if amountCol == -1 || col != amountCol {
			return baseStyle
		}

		dataRow := row
		if dataRow < 0 || dataRow >= len(t.rowMetadata) {
			dataRow = row - 1
		}
		if dataRow < 0 || dataRow >= len(t.rowMetadata) {
			return baseStyle
		}

		metadata := t.rowMetadata[dataRow]
		if metadata == nil {
			return baseStyle
		}

		switch metadata["type"] {
		case "income":
			return baseStyle.Foreground(theme.TableAmountIncomeText)
		case "expense":
			return baseStyle.Foreground(theme.TableAmountExpenseText)
		case "transfer_in", "transfer_out", "transfer":
			return baseStyle.Foreground(theme.TableAmountTransferText)
		default:
			return baseStyle
		}
	}

	renderedTable := table.New().
		Border(lipgloss.Border{}).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false).
		BorderHeader(false).
		BorderRow(false).
		BorderStyle(borderStyle).
		Headers(t.headers...).
		Rows(t.rows...).
		StyleFunc(styleForCell).
		Render()

	if t.typeName != TableTypeSummary {
		renderedTable = table.New().
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
			StyleFunc(styleForCell).
			Render()
	}

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
