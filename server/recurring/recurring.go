package recurring

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

const minimumNumberOfTransactions = 3
const paddingDays = 3

var (
	ErrInsufficientTransactionData = errors.New("not enough transactions, minimum of 3 required to detect recurring")
)

type Frequency struct {
	StartDate time.Time
	Frequency int
	Rule      models.RuleSet
}

type FrequencyScore struct {
	Frequency      int
	EstimatedIndex float64
	Index          float64
	Conclusion     float64
	Confidence     float64
}

type RecurringTransactionResult struct {
	Best    *Frequency
	Results []FrequencyScore
}

func DetectRecurringTransactions(
	ctx context.Context,
	transactions []models.Transaction,
) (*RecurringTransactionResult, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// We need at least 3 transactions in order to detect a pattern. Fewer than
	// this the data will be garbage.
	if len(transactions) < minimumNumberOfTransactions {
		return nil, errors.WithStack(ErrInsufficientTransactionData)
	}

	// We need to make sure that the transactions are sorted in ascending order
	// before we begin. This makes sure our start and end calculations are
	// correct.
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.After(transactions[j].Date)
	})

	// Size is the number of items in the time series we are going to build for
	// the fourier transform.
	size := calc.FourierSize
	padding := paddingDays

	// Start and end are the earliest and latest dates in the transaction dataset
	// with the padding added on.
	start := transactions[0].Date.AddDate(0, 0, -padding)
	end := transactions[len(transactions)-1].Date.AddDate(0, 0, padding)

	// How many total seconds between the start and the end.
	window := int64(end.Sub(start).Seconds())

	// How many seconds elapse for each data point in the time series.
	segment := float64(window) / float64(size)

	crumbs.Debug(span.Context(), "Detecting recurring transactions", map[string]interface{}{
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
		series[index] += complex(1, 0)
	}

	// Frequencies represents the number of days between each transaction, we will
	// evaluate the resulting frequency spectrum from the fourier transform for
	// these specific frequencies. If a group of transactions clearly show a
	// specific frequency then that will be the end result. Some frequencies are
	// allowed to overlap or tie for "first". Like 14,15,16.
	frequencies := []int{
		7,
		14,
		15,
		16,
		30,
		60,
		90,
	}

	result := calc.FastFourierTransform(series)

	scores := make([]FrequencyScore, len(frequencies))
	for f := range frequencies {
		frequency := frequencies[f]
		// Period is the frequency adjusted to the scale of our current time series.
		period := ((time.Duration(frequency) * 24 * time.Hour).Seconds()) / segment
		// Estimated index is a floating point number which indicates where in the
		// resulting frequency spectrum this frequency would be located.
		estimatedIndex := (1 / period) * float64(size)
		// Index is the actual index we will use for this frequency, this index
		// might be the same for multiple frequencies depending on the number of
		// transactions in the time series. We round so that we favor the higher or
		// lower index depending on the decimal of the estimated index that we
		// calculated above.
		index := math.Round(estimatedIndex)
		score := FrequencyScore{
			Frequency:      f,
			EstimatedIndex: estimatedIndex,
			Index:          index,
			Conclusion:     0,
			Confidence:     0,
		}

		// We will only have valid magnitudes up to N + 1 where N is the number of
		// transactions we provided to the time series. This way we eliminate a ton
		// of extra data that might be misleading. Also if our index ends up being 0
		// then this frequency is definitely not valid as 0 is always a useless
		// frequency on the spectrum.
		if index > float64(len(transactions))+1 || index == 0 {
			scores[f] = score
			continue
		}

		value := result[int(index)]
		real := real(value)
		imaginary := imag(value)
		magnitude := math.Sqrt((real * real) + (imaginary * imaginary))
		score.Conclusion = magnitude
		// TODO Add some form of confidence calculation
		scores[f] = score
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Conclusion > scores[j].Conclusion
	})

	// TODO Determine if the top score is actually the best, or if it is tied with
	// other scores. If its tied but its a compatible score (such as 14, 15 and
	// 16) then use the top score. Otherwise return no recurrence detected.

	// TODO Try to determine which transactions are part of the recurrence as well
	// as what the start date for the actual recurrence is.

	return nil, nil
}
