package output

import (
	"encoding/json"
	"io"
)

// JSONRenderer writes one JSON result followed by a newline.
type JSONRenderer struct{}

func (JSONRenderer) Render(w io.Writer, result Result) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}
