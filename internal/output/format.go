package output

import "fmt"

// Format identifies a representation used for command output.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// ParseFormat validates a requested output format.
func ParseFormat(value string) (Format, error) {
	format := Format(value)
	switch format {
	case FormatTable, FormatJSON:
		return format, nil
	default:
		return "", fmt.Errorf("unknown output format %q", value)
	}
}
