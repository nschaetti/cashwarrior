package domain

import "testing"

func TestParsePeriod_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  PeriodKind
	}{
		{"today", PeriodToday},
		{"yesterday", PeriodYesterday},
		{"week", PeriodWeek},
		{"month", PeriodMonth},
		{"year", PeriodYear},
		{"lastmonday", PeriodLastMonday},
		{"lastsunday", PeriodLastSunday},
		{"january", PeriodNamedMonth},
		{"week3", PeriodWeekNumber},
	}

	for _, tt := range tests {
		got, err := ParsePeriod(tt.input)
		if err != nil {
			t.Fatalf("ParsePeriod(%q) returned error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("ParsePeriod(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParsePeriod_Invalid(t *testing.T) {
	_, err := ParsePeriod("tomorrow")
	if err == nil {
		t.Fatal("ParsePeriod(\"tomorrow\") expected error, got nil")
	}
}

func TestIsPeriod(t *testing.T) {
	if !IsPeriod("week") {
		t.Fatal("IsPeriod(\"week\") = false, want true")
	}
	if !IsPeriod("march") {
		t.Fatal("IsPeriod(\"march\") = false, want true")
	}
	if !IsPeriod("week8") {
		t.Fatal("IsPeriod(\"week8\") = false, want true")
	}
	if IsPeriod("invalid") {
		t.Fatal("IsPeriod(\"invalid\") = true, want false")
	}
}

func TestPeriodKindString(t *testing.T) {
	if PeriodToday.String() != "today" {
		t.Fatalf("PeriodToday.String() = %q, want %q", PeriodToday.String(), "today")
	}
	if PeriodLastFriday.String() != "lastfriday" {
		t.Fatalf("PeriodLastFriday.String() = %q, want %q", PeriodLastFriday.String(), "lastfriday")
	}
}
