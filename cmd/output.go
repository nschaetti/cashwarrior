package cmd

import (
	"fmt"
	"os"

	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func commandOutputFormat(parsed parser.ParsedCmdLine) (output.Format, error) {
	if parsed.HasFlag("json") {
		return output.FormatJSON, nil
	}
	value, ok := parsed.GetFlagString("format")
	if !ok {
		return output.FormatTable, nil
	}
	return output.ParseFormat(value)
}

func renderJSON(resultType string, data any, count int) error {
	return renderJSONResult(output.SuccessResult(resultType, data, count))
}

func renderJSONResult(result output.Result) error {
	renderer, err := output.NewRenderer(output.FormatJSON)
	if err != nil {
		return err
	}
	return renderer.Render(os.Stdout, result)
}

func isJSONOutput(parsed parser.ParsedCmdLine) bool {
	format, err := commandOutputFormat(parsed)
	return err == nil && format == output.FormatJSON
}

func requireYesForJSON(parsed parser.ParsedCmdLine) error {
	if isJSONOutput(parsed) && !parsed.HasFlag("yes") {
		return fmt.Errorf("--yes is required for mutating commands in JSON mode")
	}
	return nil
}
