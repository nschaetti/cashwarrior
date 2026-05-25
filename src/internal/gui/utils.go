package gui

import "strings"

const LeftMargin = 4

func Indent(s string) string {
	return strings.Repeat(" ", LeftMargin) + s
}
