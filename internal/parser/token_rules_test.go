package parser

import "testing"

func TestArgRules(t *testing.T) {
	tests := []struct {
		name  string
		rule  ArgRule
		input string
		want  bool
	}{
		{name: "negative tag", rule: classifyNegativeTag, input: "-@food", want: true},
		{name: "short negative tag", rule: classifyNegativeTag, input: "-@", want: false},
		{name: "tag", rule: classifyTag, input: "@rent", want: true},
		{name: "attribute", rule: classifyAttribute, input: "account:cash", want: true},
		{name: "clear attribute", rule: classifyAttribute, input: "account:", want: true},
		{name: "text", rule: classifyText, input: "coffee", want: true},
	}

	for _, tt := range tests {
		if got := tt.rule(tt.input); got != tt.want {
			t.Errorf("%s rule(%q) = %t, want %t", tt.name, tt.input, got, tt.want)
		}
	}
}

func TestClassifyArgPriority(t *testing.T) {
	tests := []struct {
		input string
		want  ArgKind
	}{
		{"--help", ArgKindFlag},
		{"-@food", ArgKindTagNegative},
		{"@food", ArgKindTag},
		{"amount:-12.50", ArgKindAttribute},
		{"identifier:2026.05.02", ArgKindAttribute},
		{"date:today", ArgKindAttribute},
		{"randomtext", ArgKindText},
	}

	for _, tt := range tests {
		if got := ClassifyArg(tt.input); got != tt.want {
			t.Fatalf("ClassifyArg(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
