package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/parser"
)

// resolveTargetIdentifier extracts the single transaction id targeted by a
// destructive command, either from the first text argument or from an
// identifier filter. Only single identifiers are supported.
func resolveTargetIdentifier(parsed parser.ParsedCmdLine, command string) (string, error) {
	if len(parsed.Args) > 0 {
		if text, ok := parsed.Args[0].(parser.ArgText); ok && text.Text != "" {
			return text.Text, nil
		}
		return "", fmt.Errorf("%s requires a transaction id", command)
	}
	for _, arg := range parsed.Filters {
		attr, ok := arg.(parser.ArgAttribute)
		if ok && attr.Key == "identifier" {
			if attr.Value.ValueShape != parser.AttributeValueShapeSingle {
				return "", fmt.Errorf("%s accepts a single identifier, not lists or ranges", command)
			}
			if attr.Value.Raw != "" {
				return attr.Value.Raw, nil
			}
		}
	}
	return "", fmt.Errorf("%s requires a transaction id", command)
}
