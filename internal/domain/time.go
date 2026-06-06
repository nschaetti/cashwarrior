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

func ShortcutJanuary() (time.Time, time.Time) {
	start, end, err := namedMonthRange("january")
	if err != nil {
		panic(">" + err.Error() + "<")
	}
	return start, end
}

func ShortcutFebruary() (time.Time, time.Time) {
	start, end, err := namedMonthRange("february")
	if err != nil {
		panic(">" + err.Error() + "<")
	}
	return start, end
}

func ShortcutMarch() (time.Time, time.Time) {
	start, end, err := namedMonthRange("march")
	if err != nil {
		panic(">" + err.Error() + "<")
	}
	return start, end
}

func ShortcutApril() (time.Time, time.Time) {
	start, end, err := namedMonthRange("april")
	if err != nil {
		panic(">" + err.Error() + "<")
	}
	return start, end
}

func ShortcutMay() (time.Time, time.Time) {
	start, end, err := namedMonthRange("may")
	if err != nil {
		panic(">" + err.Error() + "<")
	}
	return start, end
}

func ShortcutJune() (time.Time, time.Time) {
	start, end, err := namedMonthRange("june")
	if err != nil {
		panic(">" + err.Error() + "<")
	}
	return start, end
}

func ShortcutJuly() (time.Time, time.Time) {
	start, end, err := namedMonthRange("july")
	if err != nil {
		panic(">" + err.Error() + "<")
	}
	return start, end
}

func ShortcutAugust() (time.Time, time.Time) {
	start, end, err := namedMonthRange("august")
	if err != nil {
		panic(">" + err.Error() + "<")
	}
	return start, end
}

var timeShortcutFunc = map[string]func() (time.Time, time.Time){
	// Main
	"today":     ShortcutToday,
	"yesterday": ShortcutYesterday,
	"week":      ShortcutWeek,
	"lastweek":  ShortcutLastWeek,
	"month":     ShortcutMonth,
	"lastmonth": ShortcutLastMonth,
	"year":      ShortcutYear,
	"lastyear":  ShortcutLastYear,
	// Day of week
	"monday":    ShortcutMonday,
	"tuesday":   ShortcutTuesday,
	"wednesday": ShortcutWednesday,
	"thursday":  ShortcutThursday,
	"friday":    ShortcutFriday,
	"saturday":  ShortcutSaturday,
	"sunday":    ShortcutSunday,
	// Month
	"january":   ShortcutJanuary,
	"february":  ShortcutFebruary,
	"march":     ShortcutMarch,
	"april":     ShortcutApril,
	"may":       ShortcutMay,
	"june":      ShortcutJune,
	"july":      ShortcutJuly,
	"august":    ShortcutAugust,
	"september": ShortcutSeptember,
	"october":   ShortcutOctober,
	"november":  ShortcutNovember,
	"december":  ShortcutDecember,
}

var namedMonths = map[string]time.Month{
	"jan":       time.January,
	"january":   time.January,
	"feb":       time.February,
	"february":  time.February,
	"mar":       time.March,
	"march":     time.March,
	"apr":       time.April,
	"april":     time.April,
	"may":       time.May,
	"jun":       time.June,
	"june":      time.June,
	"jul":       time.July,
	"july":      time.July,
	"aug":       time.August,
	"august":    time.August,
	"sep":       time.September,
	"september": time.September,
	"oct":       time.October,
	"october":   time.October,
	"nov":       time.November,
	"november":  time.November,
	"dec":       time.December,
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

func numberedMonthRange(shortcut string) (time.Time, time.Time, error) {
	value := strings.TrimPrefix(strings.ToLower(shortcut), "month")
	monthNumber, err := strconv.Atoi(value)
	if err != nil || monthNumber <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("unknown month shortcut: %s", shortcut)
	}

	now := time.Now()
	location := now.Location()

	// First day of the month
	startOfMonth := time.Date(now.Year(), time.Month(monthNumber), 1, 0, 0, 0, 0, location)

	// December
	endOfMonth := time.Date(now.Year(), time.December, 31, 0, 0, 0, 0, location)
	if monthNumber < 12 {
		endOfMonth = time.Date(now.Year(), time.Month(monthNumber+1), 1, 0, 0, 0, 0, location).
			AddDate(0, 0, -1)
	}

	return startOfMonth, endOfMonth, nil
}

// numberedWeekRange returns the start and end of the week with the given number.
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
	if strings.HasPrefix(strings.ToLower(shortcut), "week") && len(shortcut) > len("week") {
		return numberedWeekRange(shortcut)
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "month") && len(shortcut) > len("month") {
		return numberedMonthRange(shortcut)
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "year") && len(shortcut) > len("year") {
		return numberedYearRange(shortcut)
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "day") && len(shortcut) > len("day") {
		return numberedDayRange(shortcut)
	}
	return time.Time{}, time.Time{}, fmt.Errorf("unknown time shortcut: %s", shortcut)
}

func IsTimeShortcut(shortcut string) bool {
	if _, ok := timeShortcutFunc[shortcut]; ok {
		return true
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "week") && len(shortcut) > len("week") {
		_, _, err := numberedWeekRange(shortcut)
		return err == nil
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "month") && len(shortcut) > len("month") {
		_, _, err := numberedMonthRange(shortcut)
		return err == nil
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "year") && len(shortcut) > len("year") {
		_, _, err := numberedYearRange(shortcut)
		return err == nil
	}
	if strings.HasPrefix(strings.ToLower(shortcut), "day") && len(shortcut) > len("day") {
		_, _, err := numberedDayRange(shortcut)
		return err == nil
	}
	return false
}
