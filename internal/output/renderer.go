package output

import (
	"fmt"
	"io"
)

// Renderer serializes a command result to an output stream.
type Renderer interface {
	Render(w io.Writer, result Result) error
}

// NewRenderer returns the renderer for a supported format.
// The table renderer remains in the command/UI layer until the next block.
func NewRenderer(format Format) (Renderer, error) {
	switch format {
	case FormatJSON:
		return JSONRenderer{}, nil
	default:
		return nil, fmt.Errorf("output format %q is not implemented", format)
	}
}
