package recurring

import (
	"maps"
	"math"
	"slices"
	"sort"
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"
)

const (
	// Days of the month after the 28th don't exist in every month, and RRULE
	// skips a month entirely when the day doesn't exist. So days are capped at
	// the 28th, which is never later than the real day.
	latestMonthDay = 28
)

var (
	// Indexed by time.Weekday, which starts on Sunday.
	weekdays = [...]rrule.Weekday{
		rrule.SU,
		rrule.MO,
		rrule.TU,
		rrule.WE,
		rrule.TH,
		rrule.FR,
		rrule.SA,
	}
)

// GenerateRuleSet will build a ruleset for transactions that were detected as
// recurring at the specified frequency. Whenever a day has to be picked, the
// earliest day the transactions support is used. That way something funded by
// this rule is never funded later than the transaction actually happens, only
// sometimes a few days early. Twice a month is the exception, it uses the most
// common days instead, see [semiMonthlyMajorityDays].
func GenerateRuleSet(
	frequency int,
	members []models.Transaction,
	timezone *time.Location,
) (*models.RuleSet, error) {
	if len(members) == 0 {
		return nil, errors.New("cannot generate a ruleset without any transactions")
	}

	days := make([]time.Time, len(members))
	for i := range members {
		days[i] = util.Midnight(members[i].Date, timezone).In(timezone)
	}

	// Every transaction is placed on a circle representing one period of the
	// frequency, as a fraction from 0 to 1 of the way through the period.
	var position func(day time.Time) float64
	switch frequency {
	case 7, 14:
		position = func(day time.Time) float64 {
			return float64(daysSinceEpoch(day)%frequency) / float64(frequency)
		}
	case 15:
		position = func(day time.Time) float64 {
			return math.Mod(monthFraction(day), 0.5) / 0.5
		}
	case 30, 60, 90:
		interval := frequency / 30
		position = func(day time.Time) float64 {
			months := day.Year()*12 + int(day.Month()) - 1
			return (float64(months%interval) + monthFraction(day)) / float64(interval)
		}
	case 365:
		position = func(day time.Time) float64 {
			daysInYear := time.Date(day.Year(), 12, 31, 0, 0, 0, 0, day.Location()).YearDay()
			return float64(day.YearDay()-1) / float64(daysInYear)
		}
	default:
		return nil, errors.Errorf("cannot generate a ruleset for a frequency of [%d] days", frequency)
	}

	// The earliest transaction is the one right after the largest gap on that
	// circle. Just taking the smallest day of the month doesn't work, a charge on
	// the 1st that sometimes posts on the 31st would pick the 1st. This also
	// ignores a single late outlier, since that leaves the largest gap in front
	// of the usual day.
	earliest := earliestOnCircle(days, position)

	// Start the rule at the beginning of the month the earliest transaction is
	// in, at midnight in the account's timezone. Starting on the transaction
	// itself would skip that month's occurrence whenever the day was capped.
	// Weekly rules need to start on the transaction to keep their phase.
	monthStart := time.Date(earliest.Year(), earliest.Month(), 1, 0, 0, 0, 0, timezone)

	var options rrule.ROption
	switch frequency {
	case 7, 14:
		// The phase of an every other week rule comes from DTSTART, which is the
		// earliest transaction itself.
		options = rrule.ROption{
			Freq:      rrule.WEEKLY,
			Interval:  frequency / 7,
			Byweekday: []rrule.Weekday{weekdays[earliest.Weekday()]},
			Dtstart:   earliest,
		}
	case 15:
		first, second, err := semiMonthlyMajorityDays(days)
		if err != nil {
			return nil, err
		}

		options = rrule.ROption{
			Freq:       rrule.MONTHLY,
			Interval:   1,
			Bymonthday: []int{first, second},
			Dtstart:    monthStart,
		}
	case 30, 60, 90:
		// For every two or three months the phase comes from DTSTART being in the
		// same month as the earliest transaction.
		options = rrule.ROption{
			Freq:       rrule.MONTHLY,
			Interval:   frequency / 30,
			Bymonthday: []int{min(earliest.Day(), latestMonthDay)},
			Dtstart:    monthStart,
		}
	case 365:
		options = rrule.ROption{
			Freq:       rrule.YEARLY,
			Interval:   1,
			Bymonth:    []int{int(earliest.Month())},
			Bymonthday: []int{min(earliest.Day(), latestMonthDay)},
			Dtstart:    monthStart,
		}
	}

	// DTSTART is stored in UTC like every other ruleset, a DTSTART in the
	// account's timezone would be written with a TZID instead.
	options.Dtstart = options.Dtstart.UTC()
	rule, err := rrule.NewRRule(options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate rule")
	}

	ruleset := new(models.RuleSet)
	ruleset.RRule(rule)

	return ruleset, nil
}

// earliestOnCircle will return the day right after the largest gap between the
// positions of the days.
func earliestOnCircle(
	days []time.Time,
	position func(time.Time) float64,
) time.Time {
	type point struct {
		day      time.Time
		position float64
	}

	points := make([]point, len(days))
	for i, day := range days {
		points[i] = point{
			day:      day,
			position: position(day),
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].position < points[j].position
	})

	// Start with the gap that wraps around from the last point to the first.
	earliest := points[0]
	largestGap := points[0].position + 1 - points[len(points)-1].position
	for i := 1; i < len(points); i++ {
		if gap := points[i].position - points[i-1].position; gap > largestGap {
			largestGap = gap
			earliest = points[i]
		}
	}

	return earliest.day
}

// semiMonthlyMajorityDays will return the most common day in each half of the
// month for a twice a month rule. Twice a month is usually a payroll on the
// 15th and the last day of the month, which moves to the previous business day
// on weekends and holidays. The earliest day would put it on the 12th or 13th,
// the most common day keeps it on the 15th. A day that is the last day of its
// month counts as -1, so the 30th, 31st and the end of February all count
// towards the same day.
func semiMonthlyMajorityDays(days []time.Time) (int, int, error) {
	// Splitting the month on the 15th doesn't work when both days are in the same
	// half, like a payroll on the 1st and the 15th. The second half would be
	// empty and come back as day 0. So instead put every day of the month we saw
	// on a 31 day circle, the two largest gaps between them are where each half
	// starts. For the 1st and the 15th the gaps are 14 and 17 days, so the halves
	// start on the 1st and the 15th. Weekend shifts like the 13th or 14th only
	// add 1 or 2 day gaps so they never move where the halves start.
	present := map[int]bool{}
	for _, day := range days {
		present[day.Day()] = true
	}
	values := slices.Sorted(maps.Keys(present))
	if len(values) < 2 {
		return 0, 0, errors.New("cannot generate a twice a month ruleset from a single day of the month")
	}

	type gap struct {
		start int
		size  int
	}
	gaps := make([]gap, len(values))
	for i := range values {
		next := values[(i+1)%len(values)]
		gaps[i] = gap{
			start: next,
			size:  (next - values[i] + 31) % 31,
		}
	}
	sort.SliceStable(gaps, func(i, j int) bool {
		return gaps[i].size > gaps[j].size
	})
	firstStart := gaps[0].start
	firstLength := (gaps[1].start - firstStart + 31) % 31

	first, second := map[int]int{}, map[int]int{}
	for _, day := range days {
		daysInMonth := time.Date(day.Year(), day.Month()+1, 0, 0, 0, 0, 0, day.Location()).Day()
		value := day.Day()
		if value == daysInMonth {
			value = -1
		} else {
			value = min(value, latestMonthDay)
		}

		if (day.Day()-firstStart+31)%31 < firstLength {
			first[value]++
		} else {
			second[value]++
		}
	}

	// Either half could have been found first, keep the earlier day first so the
	// rule always reads the same way
	a, b := mostCommonDay(first), mostCommonDay(second)
	if earlierDay(b, a) {
		a, b = b, a
	}

	return a, b, nil
}

// mostCommonDay will return the day with the highest count, the earlier day
// wins a tie so that ties still lean towards funding early.
func mostCommonDay(counts map[int]int) int {
	best, bestCount := 0, -1
	for day, count := range counts {
		if count > bestCount || (count == bestCount && earlierDay(day, best)) {
			best, bestCount = day, count
		}
	}

	return best
}

// earlierDay treats -1 as later than any other day of the month.
func earlierDay(a, b int) bool {
	if a == -1 {
		return false
	}

	return b == -1 || a < b
}

// monthFraction places the day of the month on a 31 day circle, from 0 to 1.
// This uses 31 days for every month instead of the length of the day's month so
// that the earliest day is the smallest day number. Otherwise the 29th of a 31
// day month would come before the 27th of February, and a rule on the 28th
// would be a day late in February.
func monthFraction(day time.Time) float64 {
	return float64(day.Day()-1) / 31
}

// daysSinceEpoch counts whole calendar days since 1970-01-01, ignoring the
// timezone offset so that every day is exactly one apart.
func daysSinceEpoch(day time.Time) int {
	return int(time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC).Unix() / 86400)
}
