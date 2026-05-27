package domain

import (
	"fmt"
	"time"
)

func ShortcutToday() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)
	return start, end
}

func ShortcutYesterday() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)
	return start, end
}

func ShortcutWeek() (time.Time, time.Time) {
	now := time.Now()
	startOfDay := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0,
		now.Location(),
	)
	daysSinceMonday := (int(startOfDay.Weekday()) + 6) % 7
	startOfWeek := startOfDay.AddDate(0, 0, -daysSinceMonday)
	endOfWeek := startOfWeek.AddDate(0, 0, 7).Add(-time.Second)
	return startOfWeek, endOfWeek
}

func ShortcutLastWeek() (time.Time, time.Time) {
	start, end := ShortcutWeek()
	return start.AddDate(0, 0, -7), end.AddDate(0, 0, -7)
}

func ShortcutMonth() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return start, now
}

func ShortcutLastMonth() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)
	return start, end
}

func ShortcutYear() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	return start, now
}

func ShortcutLastYear() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(1, 0, 0)
	return start, end
}

var timeShortcutFunc = map[string]func() (time.Time, time.Time){
	"today":     ShortcutToday,
	"yesterday": ShortcutYesterday,
	"week":      ShortcutWeek,
	"lastweek":  ShortcutLastWeek,
	"month":     ShortcutMonth,
	"lastmonth": ShortcutLastMonth,
	"year":      ShortcutYear,
	"lastyear":  ShortcutLastYear,
}

func GetTimeShortcut(shortcut string) (time.Time, time.Time, error) {
	if shortcutFunc, ok := timeShortcutFunc[shortcut]; ok {
		from, to := shortcutFunc()
		return from, to, nil
	}
	return time.Time{}, time.Time{}, fmt.Errorf("unknown time shortcut: %s", shortcut)
}

func IsTimeShortcut(shortcut string) bool {
	_, ok := timeShortcutFunc[shortcut]
	return ok
}
