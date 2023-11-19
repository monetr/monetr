package recurring

import (
	"time"

	"github.com/monetr/monetr/server/util"
	"github.com/teambition/rrule-go"
)

type Window struct {
	// Start is the beginning of the window, all dates generated by the window will come after this one.
	Start time.Time
	// Rule is the recurrence set used to generate the subsequent dates for the window.
	Rule *rrule.Set
	// Fuzzy is the number of days in either direction from the generated recurrence date that an input date might be
	// considered valid as part of this window. Fuzzy days cannot overlap with another recurrence of the same window and
	// must always be less than the delta of the recurrence.
	Fuzzy int
}

const day = 24 * time.Hour

func (w Window) GetDeviation(date time.Time) (absoluteDays int, ok bool) {
	// If the provided date comes before this window even starts then we return -1 and false since it is not a valid match
	// against the current window.
	if date.Before(w.Start) {
		return -1, false
	}

	end := w.Rule.After(date, true)
	start := w.Rule.Before(date, true)

	// TODO Does not account for daylight savings time.

	{ // From the end date to the provided date, the end date is after.
		endDiff := end.Sub(date)
		if endDiff >= 0 {
			days := int(endDiff.Hours() / 24)
			if days <= w.Fuzzy {
				return days, true
			}
		}
	}

	{ // From the provided date to the start date, which is before the provided.
		startDiff := date.Sub(start)
		if startDiff >= 0 {
			days := int(startDiff.Hours() / 24)
			if days <= w.Fuzzy {
				return days, true
			}
		}
	}

	return -1, false
}

func GetWindowsForDate(date time.Time, timezone *time.Location) []Window {
	date = util.Midnight(date, timezone)
	windows := make([]Window, 0)
	switch getDayOfMonth(date) {
	case 1:
		windows = append(windows,
			windowFirstAndFifthteenth(date),
		)
	case 15:
		windows = append(windows,
			windowFirstAndFifthteenth(date),
			windowFifthteenthAndTheLastDay(date),
		)
	case -1: // Last day of the month
		windows = append(windows,
			windowFifthteenthAndTheLastDay(date),
		)
	}

	windows = append(windows,
		windowMonthly(date),
		windowBiMonthly(date),
		windowQuarterly(date),
		windowSemiYearly(date),
		windowYearly(date),
		windowWeekly(date),
		windowBiWeekly(date),
	)

	return windows
}

// getDayOfMonth returns the day of the month, if the day of the month is the last day then -1 will be returned.
func getDayOfMonth(date time.Time) int {
	tomorrow := date.AddDate(0, 0, 1)
	if tomorrow.Month() != date.Month() {
		return -1
	}

	return date.Day()
}

func getDayOfWeek(date time.Time) rrule.Weekday {
	switch date.Weekday() {
	case time.Sunday:
		return rrule.SU
	case time.Monday:
		return rrule.MO
	case time.Tuesday:
		return rrule.TU
	case time.Wednesday:
		return rrule.WE
	case time.Thursday:
		return rrule.TH
	case time.Friday:
		return rrule.FR
	case time.Saturday:
		return rrule.SA
	default:
		panic("new day of the week has been invented")
	}
}

func windowFirstAndFifthteenth(date time.Time) Window {
	set, _ := rrule.StrToRRuleSet("RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1,15")
	set.DTStart(date)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 5,
	}
}

func windowFifthteenthAndTheLastDay(date time.Time) Window {
	set, _ := rrule.StrToRRuleSet("RRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
	set.DTStart(date)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 5,
	}
}

func windowMonthly(date time.Time) Window {
	option := rrule.ROption{
		Freq:     rrule.MONTHLY,
		Dtstart:  date,
		Interval: 1,
		Bymonthday: []int{
			date.Day(),
		},
	}

	rule, _ := rrule.NewRRule(option)
	set := &rrule.Set{}
	set.RRule(rule)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 7,
	}
}

func windowBiMonthly(date time.Time) Window {
	option := rrule.ROption{
		Freq:     rrule.MONTHLY,
		Dtstart:  date,
		Interval: 2,
		Bymonthday: []int{
			date.Day(),
		},
	}

	rule, _ := rrule.NewRRule(option)
	set := &rrule.Set{}
	set.RRule(rule)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 7,
	}
}

func windowQuarterly(date time.Time) Window {
	option := rrule.ROption{
		Freq:     rrule.MONTHLY,
		Dtstart:  date,
		Interval: 3,
		Bymonthday: []int{
			date.Day(),
		},
	}

	rule, _ := rrule.NewRRule(option)
	set := &rrule.Set{}
	set.RRule(rule)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 7,
	}
}

func windowSemiYearly(date time.Time) Window {
	option := rrule.ROption{
		Freq:     rrule.MONTHLY,
		Dtstart:  date,
		Interval: 6,
		Bymonthday: []int{
			date.Day(),
		},
	}

	rule, _ := rrule.NewRRule(option)
	set := &rrule.Set{}
	set.RRule(rule)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 7,
	}
}

func windowYearly(date time.Time) Window {
	option := rrule.ROption{
		Freq:     rrule.YEARLY,
		Dtstart:  date,
		Interval: 1,
		// TODO this might not be right
		Bymonthday: []int{
			date.Day(),
		},
		Bymonth: []int{
			int(date.Month()),
		},
	}

	rule, _ := rrule.NewRRule(option)
	set := &rrule.Set{}
	set.RRule(rule)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 14,
	}
}

func windowWeekly(date time.Time) Window {
	option := rrule.ROption{
		Freq:     rrule.WEEKLY,
		Dtstart:  date,
		Interval: 1,
		Byweekday: []rrule.Weekday{
			getDayOfWeek(date),
		},
	}

	rule, _ := rrule.NewRRule(option)
	set := &rrule.Set{}
	set.RRule(rule)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 2,
	}
}

func windowBiWeekly(date time.Time) Window {
	option := rrule.ROption{
		Freq:     rrule.WEEKLY,
		Dtstart:  date,
		Interval: 2,
		Byweekday: []rrule.Weekday{
			getDayOfWeek(date),
		},
	}

	rule, _ := rrule.NewRRule(option)
	set := &rrule.Set{}
	set.RRule(rule)

	return Window{
		Start: date,
		Rule:  set,
		Fuzzy: 3,
	}
}
