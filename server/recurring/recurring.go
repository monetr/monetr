package recurring

import (
	"context"
	"io"
	"sort"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type Detection struct {
	nameComparator       TransactionNameComparator
	minimumNameScore     float64
	merchantComparator   TransactionMerchantComparator
	minimumMerchantScore float64
	base                 models.Transaction
	transactions         []models.Transaction
	windows              []Window
}

func NewDetection(
	name TransactionNameComparator,
	minimumNameScore float64,
	merchant TransactionMerchantComparator,
	minimumMerchantScore float64,
	base models.Transaction,
) *Detection {
	return &Detection{
		nameComparator:       name,
		minimumNameScore:     minimumNameScore,
		merchantComparator:   merchant,
		minimumMerchantScore: minimumMerchantScore,
		base:                 base,
		transactions:         make([]models.Transaction, 0),
		windows:              make([]Window, 0),
	}
}

func (d *Detection) Load(ctx context.Context, reader TransactionReader) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for {
		transaction, err := reader.Read()
		if err == io.EOF {
			// TODO Log that we reached the end?
			break
		} else if err != nil {
			return errors.Wrap(err, "failed to read next transaction")
		}

		nameScore := d.nameComparator.CompareTransactionName(d.base, *transaction)
		merchantScore := d.merchantComparator.CompareTransactionMerchant(d.base, *transaction)
		if nameScore > d.minimumNameScore || merchantScore > d.minimumMerchantScore {
			d.transactions = append(d.transactions, *transaction)
		}
	}

	// Sort transactions that match by date. This way we can walk forward incrementally somewhere else to start to figure
	// out what recurrence the transaction might be.
	sort.SliceStable(d.transactions, func(i, j int) bool {
		return d.transactions[i].Date.Before(d.transactions[j].Date)
	})

	return nil
}
