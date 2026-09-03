package cmd

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func testArg(t *testing.T, raw string) parser.Arg {
	t.Helper()

	args, err := parser.ExtractArgs([]string{raw}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ExtractArgs(%q) returned error: %v", raw, err)
	}
	return args[0]
}

func testStringAttribute(raw, key, value string) parser.Arg {
	return parser.ArgAttribute{
		Raw: raw,
		Key: key,
		Value: parser.AttributeValue{
			Raw:        value,
			ValueShape: parser.AttributeValueShapeSingle,
			Value:      parser.StringItem{Raw: value, Value: value},
		},
	}
}

func yesFlag() []parser.Arg {
	return []parser.Arg{parser.ArgFlag{Raw: "--yes", Key: "yes"}}
}
