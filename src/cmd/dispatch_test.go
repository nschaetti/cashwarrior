package cmd

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestAllKnownCommandsHaveHandlers(t *testing.T) {
	for _, name := range parser.KnownCommands() {
		if _, ok := Handlers[name]; !ok {
			t.Fatalf("missing handler for command %q", name)
		}
	}
}
