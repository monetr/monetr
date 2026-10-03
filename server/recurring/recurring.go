package recurring

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

const minimumNumberOfTransactions = 3
const paddingDays = 3
const individualMagnitude float64 = 1024
const confidenceMinimum float64 = 0.6

// A frequency is only considered when the transactions span at least this many
// of its periods. This is two periods with some slack, since months aren't all
// the same length and charges move around by a few days. Three monthly charges
// from January 2nd to March 2nd only span 1.94 months.
const minimumPeriods float64 = 1.8

// A frequency is only considered when at least this much of the occurrences it
// would expect over the span of the transactions are actually there.
const minimumOccurrenceRatio float64 = 0.75

// A frequency is only considered when the median gap between transactions is
// within this much of its period, 0.25 is plus or minus 25%.
const gapTolerance float64 = 0.25

// Yearly gets a tighter gap tolerance. Plus or minus 25% of a year is three
// months, which lets a few visits to the same place that happen to be roughly a
// year apart look yearly. Plus or minus 10% is still over a month either way,
// which is plenty for a renewal or a yearly tax payment.
const yearlyGapTolerance float64 = 0.10

// A gap between transactions that is longGapMultiple times the median gap, and
// at least longGapMinimumDays long, means the recurrence stopped for a while.
const longGapMultiple float64 = 3
const longGapMinimumDays float64 = 60

var (
	ErrInsufficientTransactionData = errors.New("not enough transactions, minimum of 3 required to detect recurring")
)

type Frequency struct {
	StartDate  time.Time
	EndDate    *time.Time
	Frequency  int
	Confidence float32
	Rule       models.RuleSet
}

type FrequencyScore struct {
	Frequency      int
	EstimatedIndex float64
	Index          float64
	Conclusion     float64
	Confidence     float64
}

type RecurringTransactionResult struct {
	Direction models.Direction
	Best      *Frequency
	// RuleSet is only present when Best is, see GenerateRuleSet.
	RuleSet *models.RuleSet
	Members []models.Transaction
	Results []FrequencyScore
}

// DetectRecurringTransactions splits the transactions by direction, debits and
// credits, and detects recurrence for each direction on its own. A recurring
// group can only be one direction, so this returns at most one result per
// direction. A direction with fewer than minimumNumberOfTransactions is
// skipped, so this can return no results at all. Transactions with an amount of
// zero have no direction and are ignored. The timezone should be the account's,
// it is used to generate the ruleset for each direction that recurs.
func DetectRecurringTransactions(
	ctx context.Context,
	clock clock.Clock,
	timezone *time.Location,
	transactions []models.Transaction,
) ([]RecurringTransactionResult, error) {
	debits := make([]models.Transaction, 0, len(transactions))
	credits := make([]models.Transaction, 0, len(transactions))
	for i := range transactions {
		switch {
		case transactions[i].Amount > 0:
			debits = append(debits, transactions[i])
		case transactions[i].Amount < 0:
			credits = append(credits, transactions[i])
		}
	}

	results := make([]RecurringTransactionResult, 0, 2)
	for _, group := range []struct {
		direction    models.Direction
		transactions []models.Transaction
	}{
		{models.DebitDirection, debits},
		{models.CreditDirection, credits},
	} {
		result, err := detectRecurringTransactions(
			ctx,
			clock,
			timezone,
			group.direction,
			group.transactions,
		)
		if errors.Is(err, ErrInsufficientTransactionData) {
			continue
		}
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}

	return results, nil
}

// detectRecurringTransactions detects recurrence for transactions that are all
// the specified direction.
func detectRecurringTransactions(
	ctx context.Context,
	_ clock.Clock,
	timezone *time.Location,
	direction models.Direction,
	transactions []models.Transaction,
) (*RecurringTransactionResult, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	numberOfTransactions := len(transactions)

	// We need at least 3 transactions in order to detect a pattern. Fewer than
	// this the data will be garbage.
	if numberOfTransactions < minimumNumberOfTransactions {
		return nil, errors.WithStack(ErrInsufficientTransactionData)
	}

	// We need to make sure that the transactions are sorted in ascending order
	// before we begin. This makes sure our start and end calculations are
	// correct.
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.Before(transactions[j].Date)
	})

	// If the recurrence stopped for a while and then started again, or there are
	// a few old one off transactions mixed in, the long gap throws off every
	// frequency below. So only look at the transactions after the most recent
	// long gap, as long as there are still enough of them.
	transactions = trimToMostRecentRun(transactions)
	numberOfTransactions = len(transactions)

	// How many days the transactions span and the median gap between them, these
	// are used to rule out frequencies that the transactions can't support.
	spanDays := transactions[numberOfTransactions-1].Date.Sub(transactions[0].Date).Hours() / 24
	medianGap := medianGapDays(transactions)

	// Size is the number of items in the time series we are going to build for
	// the fourier transform.
	size := calc.FourierSize
	padding := paddingDays

	// Start and end are the earliest and latest dates in the transaction dataset
	// with the padding added on.
	start := transactions[0].Date.AddDate(0, 0, -padding)
	end := transactions[numberOfTransactions-1].Date.AddDate(0, 0, padding)

	// How many total seconds between the start and the end.
	window := int64(end.Sub(start).Seconds())

	// How many seconds elapse for each data point in the time series.
	segment := float64(window) / float64(size)

	crumbs.Debug(span.Context(), "Detecting recurring transactions", map[string]any{
		"start":   start,
		"end":     end,
		"segment": segment,
		"window":  window,
		"size":    size,
		"padding": padding,
		"count":   len(transactions),
	})

	// Build our time series of transaction items.
	series := make([]complex128, size)
	for i := range transactions {
		txn := transactions[i]
		// Calculate the index by taking the number of seconds after the start
		// timestamp. Multiplying that by our segment size, and rounding down to get
		// our index.
		secondsSinceStart := float64(txn.Date.Sub(start).Seconds())
		// Then we can divide the number of seconds by our segment size; this will
		// tell us the index we want to use.
		index := int(math.Round(secondsSinceStart / segment))
		// Store the transaction at it's index, if there are multiple transactions
		// on the same "day" then this will increment the "count" of the
		// transactions on that day by incrementing the real part of the complex
		// number.
		series[index] += complex(individualMagnitude, 0)
	}

	// Frequencies represents the number of days between each transaction, we will
	// evaluate the resulting frequency spectrum from the fourier transform for
	// these specific frequencies. If a group of transactions clearly show a
	// specific frequency then that will be the end result. Anything based on
	// months or years uses their average length instead of a whole number of
	// days, otherwise over a few years of data the frequency drifts away from
	// where a calendar month actually lands on the spectrum.
	frequencies := []struct {
		Frequency int
		Days      float64
		// Tolerance is how far the median gap between transactions can be from
		// Days, as a fraction of Days.
		Tolerance float64
	}{
		{7, 7, gapTolerance},              // Weekly
		{14, 14, gapTolerance},            // Every 2 weeks
		{15, 365.25 / 24, gapTolerance},   // Twice a month
		{30, 365.25 / 12, gapTolerance},   // Monthly
		{60, 365.25 / 6, gapTolerance},    // Every 2 months
		{90, 365.25 / 4, gapTolerance},    // Quarterly
		{365, 365.25, yearlyGapTolerance}, // Yearly
	}

	result := calc.FastFourierTransform(series)

	scores := make([]FrequencyScore, len(frequencies))
	for f := range frequencies {
		frequency := frequencies[f]
		// Period is the frequency adjusted to the scale of our current time series.
		period := (frequency.Days * 24 * 60 * 60) / segment
		// Estimated index is a floating point number which indicates where in the
		// resulting frequency spectrum this frequency would be located.
		estimatedIndex := (1 / period) * float64(size)
		score := FrequencyScore{
			Frequency:      frequency.Frequency,
			EstimatedIndex: estimatedIndex,
			Index:          math.Round(estimatedIndex),
			Conclusion:     0,
			Confidence:     0,
		}

		// Rule out frequencies the transactions can't support. They have to span
		// about two periods, most of the occurrences we would expect over that span
		// have to actually be there, and the typical gap between them has to be
		// close to the period. Otherwise a handful of transactions, or a merchant
		// that just gets visited a lot, can line up with a frequency by chance.
		periods := spanDays / frequency.Days
		if periods < minimumPeriods ||
			float64(numberOfTransactions) < minimumOccurrenceRatio*(periods+1) ||
			math.Abs(medianGap-frequency.Days) > frequency.Tolerance*frequency.Days {
			scores[f] = score
			continue
		}

		// The estimated index is almost never a whole number, and the frequency can
		// land on either side of it. So check the indexes on both sides and keep
		// whichever one is stronger.
		for _, index := range []float64{math.Floor(estimatedIndex), math.Ceil(estimatedIndex)} {
			value := result[int(index)]
			realPart := real(value)
			imaginary := imag(value)
			magnitude := math.Sqrt((realPart * realPart) + (imaginary * imaginary))
			// Confidence is the magnitude over the maximum potential magnitude. If we
			// have 3 transactions each with an individual magnitude of 1024, then the
			// maximum achievable magnitude is 3072. So we can take the magnitude of
			// the frequency we are checking against over the maximum magnitude
			// possible and determine how much of the transaction data is represented
			// by that frequency. We can then throw out frequencies that represent a
			// lower portion of the overall transaction dataset.
			confidence := magnitude / (individualMagnitude * float64(numberOfTransactions))
			if confidence > score.Confidence {
				score.Index = index
				score.Conclusion = magnitude
				score.Confidence = confidence
			}
		}
		scores[f] = score
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Confidence > scores[j].Confidence
	})

	frequency := scores[0]
	if frequency.Confidence < confidenceMinimum {
		return &RecurringTransactionResult{
			Direction: direction,
			Best:      nil,
			Members:   nil,
			Results:   scores,
		}, nil
	}

	// This index represents the spike in the frequency spectrum for the frequency
	// that we want to isolate, whichever side of the estimated index scored
	// higher above.
	index := int(frequency.Index)

	// Create a new series based on the output of the original fourier transform.
	n := len(result)
	isolatedSeries := make([]complex128, n)
	copy(isolatedSeries, result)

	// Then zero out everything in the isolated series except for the index that
	// we want to isolate. We need to isolate the mirror of our index as well
	// because the result of a fourier transform is symetrical. Isolating only one
	// side will fuck up the results of the inverse transform we will be
	// performing below.
	for i := range isolatedSeries {
		if i == index || i == (n-index)%n {
			continue
		}
		isolatedSeries[i] = complex(0, 0)
	}

	// Now we perform the inverse fourier transform on our data. This will return
	// a waveform that only represents the frequency we have selected above.
	invertedSeries := calc.InverseFastFourierTransform(isolatedSeries)
	// Take only the real portion of our inverted series and put that into its own
	// array.
	signal := make([]float64, len(invertedSeries))
	for i := range signal {
		signal[i] = real(invertedSeries[i])
	}

	// Now we need to calculate a threshold for peak detection of our waveform.
	// Peaks represent the actually recurrences in our original dataset.
	var threshold float64
	{
		duplicate := make([]float64, len(signal))
		copy(duplicate, signal)
		sort.Float64s(duplicate)
		// Sort the points from the resulting signal least to greatest, then take
		// the value of the point at 66% through the sorted results. This value
		// represents a point that is higher than 66% of the magnitudes, and thus
		// any value greater than this might be considered a peak. Essentially we
		// want to isolate the top 33% of the waveform.
		cut := int(float64(len(duplicate)) / 3)
		threshold = duplicate[len(duplicate)-cut]
	}

	// Now we can scan over the waveform and find all of the ranges of peaks on
	// the wave. These peaks should correlate strongly with indicies of
	// transactions in our original dataset when converted to a time series.
	ranges := make([][2]int, 0, numberOfTransactions)
	startOfRange := -1
	for x, y := range signal {
		// If we are under the threshold and not currently observing a range of
		// indicies then just keep going.
		if y < threshold && startOfRange == -1 {
			continue
		}

		if startOfRange == -1 {
			startOfRange = x
		} else if threshold > y {
			ranges = append(ranges, [2]int{
				startOfRange,
				x,
			})
			startOfRange = -1
		}
	}
	if startOfRange != -1 {
		ranges = append(ranges, [2]int{
			startOfRange,
			len(signal) - 1,
		})
	}

	// Now take our ranges and isolate the transactions that belong to this
	// frequency so we can include those in our result.
	members := make([]models.Transaction, 0, numberOfTransactions)
	lastIndex := 0
	for _, txnRange := range ranges {
		a, b := txnRange[0], txnRange[1]
		for i := lastIndex; i < numberOfTransactions; i++ {
			txn := transactions[i]
			secondsSinceStart := float64(txn.Date.Sub(start).Seconds())
			index := int(math.Round(secondsSinceStart / segment))

			if index >= a && index <= b {
				// If the transaction falls in our peak range then add it to the member
				// array.
				members = append(members, txn)
				// This way on the next range we don't need to reread transactions, we
				// can just jump right to the spot we havent read.
				lastIndex = i
			}
		}
	}

	// TODO Determine if the top score is actually the best, or if it is tied with
	// other scores. If its tied but its a compatible score (such as 14, 15 and
	// 16) then use the top score. Otherwise return no recurrence detected.

	ruleset, err := GenerateRuleSet(frequency.Frequency, members, timezone)
	if err != nil {
		return nil, err
	}

	return &RecurringTransactionResult{
		Direction: direction,
		Best: &Frequency{
			StartDate:  members[0].Date,
			Frequency:  frequency.Frequency,
			Confidence: float32(frequency.Confidence),
		},
		RuleSet: ruleset,
		Members: members,
		Results: scores,
	}, nil
}

// medianGapDays returns the median number of days between transactions, which
// must already be sorted by date. Transactions on the same day are skipped, so
// something that charges a few times on the same day every month still has a
// median gap of a month.
func medianGapDays(transactions []models.Transaction) float64 {
	gaps := make([]float64, 0, len(transactions))
	for i := 1; i < len(transactions); i++ {
		if gap := transactions[i].Date.Sub(transactions[i-1].Date).Hours() / 24; gap > 0 {
			gaps = append(gaps, gap)
		}
	}
	if len(gaps) == 0 {
		return 0
	}
	sort.Float64s(gaps)
	return gaps[len(gaps)/2]
}

// trimToMostRecentRun drops every transaction before the most recent long gap,
// see longGapMultiple. The transactions must already be sorted by date. If that
// would leave fewer than minimumNumberOfTransactions then nothing is dropped.
func trimToMostRecentRun(transactions []models.Transaction) []models.Transaction {
	longGap := math.Max(longGapMultiple*medianGapDays(transactions), longGapMinimumDays)
	for i := len(transactions) - 1; i > 0; i-- {
		if transactions[i].Date.Sub(transactions[i-1].Date).Hours()/24 > longGap {
			if len(transactions)-i >= minimumNumberOfTransactions {
				return transactions[i:]
			}
			break
		}
	}
	return transactions
}
