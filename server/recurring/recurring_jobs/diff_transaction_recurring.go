package recurring_jobs

import (
	"context"
	"slices"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/recurring"
)

// RecurringDiff is everything that changed between the recurring transactions
// we have stored for a cluster and the ones we just detected.
type RecurringDiff struct {
	// Every recurring transaction that was detected. If one already existed for
	// the same direction then it keeps the existing ID.
	UpsertRecurring []models.TransactionRecurring
	// Existing recurring transactions for a direction that doesn't recur
	// anymore, these can just be deleted.
	DeleteRecurringIds []models.ID[models.TransactionRecurring]
	// Transactions in the cluster whose recurring ID needs to change. Members of
	// a recurring transaction point at it, everything else in the cluster has
	// its recurring ID removed.
	UpdateMembers []models.Transaction
}

// DiffTransactionRecurring figures out what needs to be written to get the
// database in line with the recurring transactions we just detected for a
// cluster. There is only ever one recurring transaction per cluster and
// direction, so a detected one is matched up with an existing one by its
// direction and updates it instead of creating a new one. The transactions are
// every transaction in the cluster as they are in the database right now, they
// are used to figure out which ones need their recurring ID changed.
func DiffTransactionRecurring(
	ctx context.Context,
	existing []models.TransactionRecurring,
	results []recurring.RecurringTransactionResult,
	transactions []models.Transaction,
	now time.Time,
	bankAccountId models.ID[models.BankAccount],
	transactionClusterId models.ID[models.TransactionCluster],
) RecurringDiff {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	existingByDirection := make(map[models.Direction]models.TransactionRecurring, len(existing))
	for _, item := range existing {
		existingByDirection[item.Direction] = item
	}

	diff := RecurringDiff{
		UpsertRecurring:    make([]models.TransactionRecurring, 0, len(results)),
		DeleteRecurringIds: make([]models.ID[models.TransactionRecurring], 0, len(existing)),
		UpdateMembers:      make([]models.Transaction, 0, len(transactions)),
	}
	// Which recurring transaction each member should point at.
	newOwner := make(map[models.ID[models.Transaction]]models.ID[models.TransactionRecurring], len(transactions))
	for _, result := range results {
		// A direction that doesn't recur has no recurring transaction, if there was
		// one before then it gets deleted below.
		if result.Best == nil {
			continue
		}

		item := newTransactionRecurring(result, now, bankAccountId, transactionClusterId)
		if old, ok := existingByDirection[result.Direction]; ok {
			item.TransactionRecurringId = old.TransactionRecurringId
			item.CreatedAt = old.CreatedAt
			delete(existingByDirection, result.Direction)
		} else {
			// New recurring transactions get their ID now instead of when they are
			// inserted, that way the members can point at it.
			item.TransactionRecurringId = models.NewID[models.TransactionRecurring]()
		}
		diff.UpsertRecurring = append(diff.UpsertRecurring, item)

		for _, member := range result.Members {
			newOwner[member.TransactionId] = item.TransactionRecurringId
		}
	}

	// Anything left didn't match a direction that still recurs.
	for _, item := range existingByDirection {
		diff.DeleteRecurringIds = append(diff.DeleteRecurringIds, item.TransactionRecurringId)
	}
	// Map iteration is random, keep the order stable.
	slices.Sort(diff.DeleteRecurringIds)

	// Compare what each transaction in the cluster points at now with what it
	// should point at, and only update the ones that changed.
	for _, txn := range transactions {
		var want *models.ID[models.TransactionRecurring]
		if owner, ok := newOwner[txn.TransactionId]; ok {
			want = &owner
		}
		have := txn.TransactionRecurringId
		if (want == nil && have == nil) || (want != nil && have != nil && *want == *have) {
			continue
		}
		diff.UpdateMembers = append(diff.UpdateMembers, models.Transaction{
			TransactionId:          txn.TransactionId,
			TransactionRecurringId: want,
		})
	}

	return diff
}

// newTransactionRecurring builds the recurring transaction for a detection
// result that recurs. Result must have Best and RuleSet present.
func newTransactionRecurring(
	result recurring.RecurringTransactionResult,
	now time.Time,
	bankAccountId models.ID[models.BankAccount],
	transactionClusterId models.ID[models.TransactionCluster],
) models.TransactionRecurring {
	// Members are sorted by date already, but the same transaction can show up
	// more than once.
	seen := make(map[models.ID[models.Transaction]]struct{}, len(result.Members))
	amounts := make(map[int64]int, len(result.Members))
	var first, last models.Transaction
	for _, member := range result.Members {
		if _, ok := seen[member.TransactionId]; ok {
			continue
		}
		seen[member.TransactionId] = struct{}{}
		amounts[member.Amount]++
		if first.TransactionId.IsZero() || member.Date.Before(first.Date) {
			first = member
		}
		if last.TransactionId.IsZero() || !member.Date.Before(last.Date) {
			last = member
		}
	}

	// The recurrence has ended once it is past the occurrence that should have
	// come after the last transaction, with half a period of slack since charges
	// move around by a few days.
	expected := result.RuleSet.After(last.Date, false)
	ended := now.After(expected.AddDate(0, 0, max(result.Best.Frequency/2, 3)))

	return models.TransactionRecurring{
		BankAccountId:        bankAccountId,
		TransactionClusterId: transactionClusterId,
		Window:               windowType(result),
		RuleSet:              result.RuleSet,
		First:                first.Date,
		Last:                 last.Date,
		Next:                 result.RuleSet.After(now, false),
		Ended:                ended,
		Confidence:           result.Best.Confidence,
		Direction:            result.Direction,
		Amounts:              amounts,
		LastAmount:           last.Amount,
	}
}

// windowType maps the detected frequency to the window type of the recurring
// transaction.
func windowType(result recurring.RecurringTransactionResult) models.WindowType {
	switch result.Best.Frequency {
	case 7:
		return models.WeeklyWindowType
	case 14:
		return models.BiWeeklyWindowType
	case 15:
		// Twice a month is either the 1st and 15th or the 15th and the last day of
		// the month, the ruleset has the last day as -1.
		if slices.Contains(result.RuleSet.GetRRule().OrigOptions.Bymonthday, -1) {
			return models.FifteenthAndLastWindowType
		}
		return models.FirstAndFifteenthWindowType
	case 60:
		return models.BiMonthlyWindowType
	case 90:
		return models.QuarterlyWindowType
	case 365:
		return models.YearlyWindowType
	default:
		return models.MonthlyWindowType
	}
}
