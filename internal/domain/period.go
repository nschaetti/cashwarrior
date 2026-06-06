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
	// PeriodLastWeek matches entries from the previous week.
	PeriodLastWeek
	// PeriodMonth matches entries from the current month.
	PeriodMonth
	// PeriodLastMonth matches entries from the previous month.
	PeriodLastMonth
	// PeriodYear matches entries from the current year.
	PeriodYear
	// PeriodLastYear matches entries from the previous year.
	PeriodLastYear
	// PeriodMonday matches entries from last Monday.
	PeriodMonday
	// PeriodTuesday matches entries from last Tuesday.
	PeriodTuesday
	// PeriodWednesday matches entries from last Wednesday.
	PeriodWednesday
	// PeriodThursday matches entries from last Thursday.
	PeriodThursday
	// PeriodFriday matches entries from last Friday.
	PeriodFriday
	// PeriodSaturday matches entries from last Saturday.
	PeriodSaturday
	// PeriodSunday matches entries from last Sunday.
	PeriodSunday
	// PeriodJanuary matches entries from January.
	PeriodJanuary
	// PeriodFebruary matches entries from February.
	PeriodFebruary
	// PeriodMarch matches entries from March.
	PeriodMarch
	// PeriodApril matches entries from April.
	PeriodApril
	// PeriodMay matches entries from May.
	PeriodMay
	// PeriodJune matches entries from June.
	PeriodJune
	// PeriodJuly matches entries from July.
	PeriodJuly
	// PeriodAugust matches entries from August.
	PeriodAugust
	// PeriodSeptember matches entries from September.
	PeriodSeptember
	// PeriodOctober matches entries from October.
	PeriodOctober
	// PeriodNovember matches entries from November.
	PeriodNovember
	// PeriodDecember matches entries from December.
	PeriodDecember
)

var periodNames = map[string]PeriodKind{
	"today":     PeriodToday,
	"yesterday": PeriodYesterday,
	"week":      PeriodWeek,
	"lastweek":  PeriodLastWeek,
	"month":     PeriodMonth,
	"lastmonth": PeriodLastMonth,
	"year":      PeriodYear,
	"lastyear":  PeriodLastYear,
	// Day of week
	"monday":    PeriodMonday,
	"tuesday":   PeriodTuesday,
	"wednesday": PeriodWednesday,
	"thursday":  PeriodThursday,
	"friday":    PeriodFriday,
	"saturday":  PeriodSaturday,
	"sunday":    PeriodSunday,
	// Months
	"january":   PeriodJanuary,
	"february":  PeriodFebruary,
	"march":     PeriodMarch,
	"april":     PeriodApril,
	"may":       PeriodMay,
	"june":      PeriodJune,
	"july":      PeriodJuly,
	"august":    PeriodAugust,
	"september": PeriodSeptember,
	"october":   PeriodOctober,
	"november":  PeriodNovember,
	"december":  PeriodDecember,
}

var periodIDs = map[PeriodKind]string{
	PeriodToday:     "today",
	PeriodYesterday: "yesterday",
	PeriodWeek:      "week",
	PeriodLastWeek:  "lastweek",
	PeriodMonth:     "month",
	PeriodLastMonth: "lastmonth",
	PeriodYear:      "year",
	PeriodLastYear:  "lastyear",
	// Day of week
	PeriodMonday:    "monday",
	PeriodTuesday:   "tuesday",
	PeriodWednesday: "wednesday",
	PeriodThursday:  "thursday",
	PeriodFriday:    "friday",
	PeriodSaturday:  "saturday",
	PeriodSunday:    "sunday",
	// Months
	PeriodJanuary:   "january",
	PeriodFebruary:  "february",
	PeriodMarch:     "march",
	PeriodApril:     "april",
	PeriodMay:       "may",
	PeriodJune:      "june",
	PeriodJuly:      "july",
	PeriodAugust:    "august",
	PeriodSeptember: "september",
	PeriodOctober:   "october",
	PeriodNovember:  "november",
	PeriodDecember:  "december",
}

// ParsePeriod converts a period identifier (for example, "today") into PeriodKind.
func ParsePeriod(s string) (PeriodKind, error) {
	kind, ok := periodNames[s]
	if ok {
		return kind, nil
	}
	if IsTimeShortcut(s) {
		if len(s) >= 4 && s[:4] == "week" {
			return s, nil
		}
		return s, nil
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
