package domain

import (
	"testing"
	"time"
)

func TestGetTimeShortcutNamedMonth(t *testing.T) {
	from, to, err := GetTimeShortcut("january")
	if err != nil {
		t.Fatalf("GetTimeShortcut(january) returned error: %v", err)
	}

	now := time.Now()
	wantYear := now.Year()
	if time.January >= now.Month() {
		wantYear--
	}

	if from.Year() != wantYear || from.Month() != time.January || from.Day() != 1 {
		t.Fatalf("from = %v, want January 1st of %d", from, wantYear)
	}
	if to.Month() != time.January || to.Year() != wantYear {
		t.Fatalf("to = %v, want last instant of January %d", to, wantYear)
	}
}

func TestGetTimeShortcutWeekNumber(t *testing.T) {
	from, to, err := GetTimeShortcut("week1")
	if err != nil {
		t.Fatalf("GetTimeShortcut(week1) returned error: %v", err)
	}

	now := time.Now()
	want := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
	for want.Weekday() != time.Monday {
		want = want.AddDate(0, 0, 1)
	}

	if !from.Equal(want) {
		t.Fatalf("from = %v, want %v", from, want)
	}
	if !to.Equal(want.AddDate(0, 0, 7).Add(-time.Nanosecond)) {
		t.Fatalf("to = %v, want %v", to, want.AddDate(0, 0, 7).Add(-time.Nanosecond))
	}
}

func TestIsTimeShortcutSupportsNamedMonthAndWeekNumber(t *testing.T) {
	if !IsTimeShortcut("february") {
		t.Fatal("IsTimeShortcut(february) = false, want true")
	}
	if !IsTimeShortcut("week12") {
		t.Fatal("IsTimeShortcut(week12) = false, want true")
	}
	if IsTimeShortcut("week0") {
		t.Fatal("IsTimeShortcut(week0) = true, want false")
	}
}
