package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintGlobalHelpIncludesCommandsDescriptionsAndSubcommands(t *testing.T) {
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}
	os.Stdout = writer
	printGlobalHelp()
	_ = writer.Close()
	os.Stdout = original

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, reader); err != nil {
		t.Fatalf("io.Copy returned error: %v", err)
	}
	_ = reader.Close()
	result := buffer.String()

	for _, expected := range []string{
		"accounts   ",
		"List and manage accounts",
		"accounts list",
		"accounts add",
		"accounts initial-balance",
		"list transactions",
		"stores rename",
		"transfer add",
		"--format table|json",
		"--yes",
	} {
		if !strings.Contains(result, expected) {
			t.Fatalf("global help missing %q:\n%s", expected, result)
		}
	}
}
