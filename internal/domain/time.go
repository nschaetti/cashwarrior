package domain

import (
	"fmt"
	"strconv"
	"strings"
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

var namedMonths = map[string]time.Month{
	"january":   time.January,
	"february":  time.February,
	"march":     time.March,
	"april":     time.April,
	"may":       time.May,
	"june":      time.June,
	"july":      time.July,
	"august":    time.August,
	"september": time.September,
	"october":   time.October,
	"november":  time.November,
	"december":  time.December,
}

func namedMonthRange(name string) (time.Time, time.Time, error) {
	month, ok := namedMonths[strings.ToLower(name)]
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("unknown month shortcut: %s", name)
	}

	now := time.Now()
	year := now.Year()
	if month >= now.Month() {
		year--
	}

	start := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end, nil
}

func numberedWeekRange(shortcut string) (time.Time, time.Time, error) {
	value := strings.TrimPrefix(strings.ToLower(shortcut), "week")
	weekNumber, err := strconv.Atoi(value)
	if err != nil || weekNumber <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("unknown week shortcut: %s", shortcut)
	}

	now := time.Now()
	location := now.Location()
	startOfYear := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, location)
	firstMonday := startOfYear
	for firstMonday.Weekday() != time.Monday {
		firstMonday = firstMonday.AddDate(0, 0, 1)
	}

	start := firstMonday.AddDate(0, 0, (weekNumber-1)*7)
	if start.Year() != now.Year() {
		return time.Time{}, time.Time{}, fmt.Errorf("unknown week shortcut: %s", shortcut)
	}
	end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
	return start, end, nil
}

func GetTimeShortcut(shortcut string) (time.Time, time.Time, error) {
	if shortcutFunc, ok := timeShortcutFunc[shortcut]; ok {
		from, to := shortcutFunc()
		return from, to, nil
	}
	if _, ok := namedMonths[strings.ToLower(shortcut)]; ok {
		return namedMonthRange(shortcut)
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "week") && len(shortcut) > len("week") {
		return numberedWeekRange(shortcut)
	}
	return time.Time{}, time.Time{}, fmt.Errorf("unknown time shortcut: %s", shortcut)
}

func IsTimeShortcut(shortcut string) bool {
	if _, ok := timeShortcutFunc[shortcut]; ok {
		return true
	}
	if _, ok := namedMonths[strings.ToLower(shortcut)]; ok {
		return true
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "week") && len(shortcut) > len("week") {
		_, _, err := numberedWeekRange(shortcut)
		return err == nil
	}
	return false
}
