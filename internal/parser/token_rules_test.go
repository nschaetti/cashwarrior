package parser

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/domain"
)

func TestClassifyNegativeTag(t *testing.T) {
	tok, ok := classifyNegativeTag("-@food")
	if !ok {
		t.Fatal("classifyNegativeTag did not match valid negative tag")
	}
	if tok.Kind != TokenTagNegative || tok.Raw != "-@food" {
		t.Fatalf("unexpected token: %#v", tok)
	}

	if _, ok := classifyNegativeTag("-@"); ok {
		t.Fatal("classifyNegativeTag matched invalid short value")
	}
}

func TestClassifyTag(t *testing.T) {
	tok, ok := classifyTag("@rent")
	if !ok {
		t.Fatal("classifyTag did not match valid tag")
	}
	if tok.Kind != TokenTag || tok.Raw != "@rent" {
		t.Fatalf("unexpected token: %#v", tok)
	}
}

func TestClassifyAmount(t *testing.T) {
	tok, ok := classifyAmount("-12.50")
	if !ok {
		t.Fatal("classifyAmount did not match valid amount")
	}
	if tok.Kind != TokenAmount || tok.Amount != float32(-12.5) {
		t.Fatalf("unexpected token: %#v", tok)
	}

	if _, ok := classifyAmount("12.50"); ok {
		t.Fatal("classifyAmount matched amount without sign")
	}
}

func TestClassifyAttribute(t *testing.T) {
	tok, ok := classifyAttribute("account:cash")
	if !ok {
		t.Fatal("classifyAttribute did not match key:value")
	}
	if tok.Kind != TokenAttribute || tok.Key != "account" || tok.Value != "cash" {
		t.Fatalf("unexpected token: %#v", tok)
	}

	clearTok, ok := classifyAttribute("account:")
	if !ok {
		t.Fatal("classifyAttribute did not match key:")
	}
	if clearTok.Kind != TokenAttributeClear || clearTok.Key != "account" {
		t.Fatalf("unexpected clear token: %#v", clearTok)
	}
}

func TestClassifyID(t *testing.T) {
	tok, ok := classifyID("2026.05.02")
	if !ok {
		t.Fatal("classifyID did not match valid id")
	}
	if tok.Kind != TokenID {
		t.Fatalf("unexpected kind: %v", tok.Kind)
	}
	if tok.TransID != (domain.TransactionID{Year: 2026, Month: 5, Num: 2}) {
		t.Fatalf("unexpected id: %#v", tok.TransID)
	}
}

func TestClassifyPeriod(t *testing.T) {
	tok, ok := classifyPeriod("today")
	if !ok {
		t.Fatal("classifyPeriod did not match valid period")
	}
	if tok.Kind != TokenPeriod || tok.Period != domain.PeriodToday {
		t.Fatalf("unexpected token: %#v", tok)
	}
}

func TestClassifyText(t *testing.T) {
	tok, ok := classifyText("coffee")
	if !ok {
		t.Fatal("classifyText should always match")
	}
	if tok.Kind != TokenText || tok.Raw != "coffee" {
		t.Fatalf("unexpected token: %#v", tok)
	}
}

func TestClassifyToken_PriorityRegression(t *testing.T) {
	tests := []struct {
		input string
		want  TokenKind
	}{
		{"-@food", TokenTagNegative},
		{"@food", TokenTag},
		{"-12.50", TokenAmount},
		{"2026.05.02", TokenID},
		{"account:cash", TokenAttribute},
		{"today", TokenPeriod},
		{"randomtext", TokenText},
	}

	for _, tt := range tests {
		got := ClassifyToken(tt.input)
		if got.Kind != tt.want {
			t.Fatalf("ClassifyToken(%q) kind = %v, want %v", tt.input, got.Kind, tt.want)
		}
	}
}
