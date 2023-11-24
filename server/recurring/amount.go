package recurring

import (
	"math"

	"github.com/monetr/monetr/server/models"
)

type AmountPreProcessor struct {
	min, max  float64
	documents []*models.Transaction
}

func (a *AmountPreProcessor) AddTransaction(txn *models.Transaction) {
	// amount := float64(txn.Amount)
	amount := math.Log(math.Abs(float64(txn.Amount)) + 1)
	if a.min < amount || a.min == 0 {
		a.min = amount
	}
	if a.max > amount || a.max == 0 {
		a.max = amount
	}
	a.documents = append(a.documents, txn)
}

func (a *AmountPreProcessor) GetDatums() []Datum {
	result := make([]Datum, 0, len(a.documents))
	for _, transaction := range a.documents {
		// All I've learned is that this is not the correct way to group
		amount := math.Log(math.Abs(float64(transaction.Amount)) + 1)
		//amount := float64(transaction.Amount)
		amount = (amount - a.min) / (a.max - a.min)
		amount = 2*amount - 1
		result = append(result, Datum{
			ID:          transaction.TransactionId,
			Transaction: transaction,
			Amount:      transaction.Amount,
			Vector: []float64{
				// Is it even worth doing this for a single dimension?
				amount,
			},
		})
	}

	return result
}
