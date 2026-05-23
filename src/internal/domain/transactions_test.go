package domain

import "testing"

func TestTransactionIDString(t *testing.T) {
	id := TransactionID{Year: 2026, Month: 5, Num: 2}
	if got := id.String(); got != "2026.5.2" {
		t.Fatalf("TransactionID.String() = %q, want %q", got, "2026.5.2")
	}
}

func TestParseTransactionID_Valid(t *testing.T) {
	got, err := ParseTransactionID("2026.05.02")
	if err != nil {
		t.Fatalf("ParseTransactionID returned error: %v", err)
	}

	want := TransactionID{Year: 2026, Month: 5, Num: 2}
	if got != want {
		t.Fatalf("ParseTransactionID = %#v, want %#v", got, want)
	}
}

func TestParseTransactionID_Invalid(t *testing.T) {
	invalid := []string{
		"2026.5.02",
		"26.05.02",
		"abcd.05.02",
		"2026-05-02",
	}

	for _, input := range invalid {
		if _, err := ParseTransactionID(input); err == nil {
			t.Fatalf("ParseTransactionID(%q) expected error, got nil", input)
		}
	}
}
