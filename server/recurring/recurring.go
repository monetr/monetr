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
	Period    int
	Rule      models.RuleSet
}

type RecurringTransactionResult struct {
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

	// result := calc.FastFourierTransform(series)

	return nil, nil
}
