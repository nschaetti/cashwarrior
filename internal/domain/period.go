package domain

import "fmt"

// PeriodKind identifies a predefined relative period used in filters.
type PeriodKind int

const (
	// PeriodToday matches entries from the current day.
	PeriodToday PeriodKind = iota
	// PeriodYesterday matches entries from the previous day.
	PeriodYesterday
	// PeriodWeek matches entries from the current week.
	PeriodWeek
	// PeriodMonth matches entries from the current month.
	PeriodMonth
	// PeriodYear matches entries from the current year.
	PeriodYear
	// PeriodLastMonday matches entries from last Monday.
	PeriodLastMonday
	// PeriodLastTuesday matches entries from last Tuesday.
	PeriodLastTuesday
	// PeriodLastWednesday matches entries from last Wednesday.
	PeriodLastWednesday
	// PeriodLastThursday matches entries from last Thursday.
	PeriodLastThursday
	// PeriodLastFriday matches entries from last Friday.
	PeriodLastFriday
	// PeriodLastSaturday matches entries from last Saturday.
	PeriodLastSaturday
	// PeriodLastSunday matches entries from last Sunday.
	PeriodLastSunday
	// PeriodNamedMonth matches shortcuts like january.
	PeriodNamedMonth
	// PeriodWeekNumber matches shortcuts like week12.
	PeriodWeekNumber
)

var periodNames = map[string]PeriodKind{
	"today":         PeriodToday,
	"yesterday":     PeriodYesterday,
	"week":          PeriodWeek,
	"month":         PeriodMonth,
	"year":          PeriodYear,
	"lastmonday":    PeriodLastMonday,
	"lasttuesday":   PeriodLastTuesday,
	"lastwednesday": PeriodLastWednesday,
	"lastthursday":  PeriodLastThursday,
	"lastfriday":    PeriodLastFriday,
	"lastsaturday":  PeriodLastSaturday,
	"lastsunday":    PeriodLastSunday,
}

var periodIDs = map[PeriodKind]string{
	PeriodToday:         "today",
	PeriodYesterday:     "yesterday",
	PeriodWeek:          "week",
	PeriodMonth:         "month",
	PeriodYear:          "year",
	PeriodLastMonday:    "lastmonday",
	PeriodLastTuesday:   "lasttuesday",
	PeriodLastWednesday: "lastwednesday",
	PeriodLastThursday:  "lastthursday",
	PeriodLastFriday:    "lastfriday",
	PeriodLastSaturday:  "lastsaturday",
	PeriodLastSunday:    "lastsunday",
	PeriodNamedMonth:    "namedmonth",
	PeriodWeekNumber:    "weeknumber",
}

// ParsePeriod converts a period identifier (for example "today") into PeriodKind.
func ParsePeriod(s string) (PeriodKind, error) {
	kind, ok := periodNames[s]
	if ok {
		return kind, nil
	}
	if IsTimeShortcut(s) {
		if len(s) >= 4 && s[:4] == "week" {
			return PeriodWeekNumber, nil
		}
		return PeriodNamedMonth, nil
	}
	return 0, fmt.Errorf("invalid period: %s", s)
}

// IsPeriod reports whether s is a supported period identifier.
func IsPeriod(s string) bool {
	_, ok := periodNames[s]
	if ok {
		return true
	}
	return IsTimeShortcut(s)
}

// String returns the canonical identifier for the period kind.
func (p PeriodKind) String() string {
	return fmt.Sprintf("%s", periodIDs[p])
}
