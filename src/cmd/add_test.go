package cmd

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestGetTransactionDescription_UsesTextTokenNotFirstArg(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Args: []parser.Token{
			{Raw: "+0.03", Kind: parser.TokenAmount, Amount: 0.03},
			{Raw: "account:swissquote", Kind: parser.TokenAttribute, Key: "account", Value: "swissquote"},
			{Raw: "store:test", Kind: parser.TokenAttribute, Key: "store", Value: "test"},
			{Raw: "Un", Kind: parser.TokenText},
			{Raw: "peu", Kind: parser.TokenText},
			{Raw: "de", Kind: parser.TokenText},
			{Raw: "Bitcoin", Kind: parser.TokenText},
			{Raw: "date:2025-11-12@08:00:00", Kind: parser.TokenAttribute, Key: "date", Value: "2025-11-12@08:00:00"},
		},
	}

	counts := map[parser.TokenKind]int{parser.TokenText: 4}
	got := getTransactionDescription(parsed, counts)
	want := "Un peu de Bitcoin"
	if got != want {
		t.Fatalf("getTransactionDescription() = %q, want %q", got, want)
	}
}

func TestGetTransactionDescription_SingleTextToken(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Args: []parser.Token{
			{Raw: "+1.00", Kind: parser.TokenAmount, Amount: 1.0},
			{Raw: "coffee", Kind: parser.TokenText},
		},
	}

	counts := map[parser.TokenKind]int{parser.TokenText: 1}
	got := getTransactionDescription(parsed, counts)
	if got != "coffee" {
		t.Fatalf("getTransactionDescription() = %q, want %q", got, "coffee")
	}
}
