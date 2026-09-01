package similar

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/monetr/monetr/server/models"
)

// syntheticTransactions builds a corpus that behaves the way a real bank account
// does as it grows. Each transaction comes from one of merchants distinct
// merchants and carries a couple of throwaway words alongside it. The number of
// meaningful words in any single transaction stays roughly constant no matter
// how big the account gets, but the vocabulary across the whole account keeps
// growing, which is what makes the vectors get wider and emptier over time.
func syntheticTransactions(count, merchants int) []models.Transaction {
	rng := rand.New(rand.NewSource(int64(count * merchants)))
	prefixes := []string{"POS Debit", "ACH Debit", "Card Purchase", "Recurring"}
	cities := []string{"Omaha", "Lincoln", "Bellevue", "Papillion", "Gretna"}

	transactions := make([]models.Transaction, 0, count)
	for range count {
		merchant := rng.Intn(merchants)
		transactions = append(transactions, models.Transaction{
			TransactionId: models.NewID[models.Transaction](),
			OriginalName: fmt.Sprintf(
				"%s Merchant%d Store%d %s",
				prefixes[rng.Intn(len(prefixes))],
				merchant,
				merchant%97,
				cities[rng.Intn(len(cities))],
			),
			Name: fmt.Sprintf("Merchant%d", merchant),
		})
	}

	return transactions
}

// BenchmarkClusteringAtScale measures the clustering of a whole account end to
// end at a few different sizes. The fixtures only produce a vocabulary of a few
// dozen words, which is small enough that the vectors are not really sparse at
// all, so they do not show what happens to an account that has been running for
// a few years.
func BenchmarkClusteringAtScale(bench *testing.B) {
	for _, size := range []struct {
		count     int
		merchants int
	}{
		{1000, 60},
		{4000, 200},
		{8000, 400},
	} {
		transactions := syntheticTransactions(size.count, size.merchants)

		bench.Run(fmt.Sprintf("txns=%d/merchants=%d", size.count, size.merchants), func(bench *testing.B) {
			processor := NewTransactionTFIDF()
			for i := range transactions {
				processor.AddTransaction(&transactions[i])
			}
			documents := processor.GetDocuments(context.Background())
			bench.ReportMetric(float64(len(documents[0].Vector)), "vectorSize")
			ctx := context.Background()

			bench.ResetTimer()
			for bench.Loop() {
				// A fresh DBSCAN every time, it keeps the visited labels from the
				// previous run otherwise and there would be no work left to do.
				_ = NewDBSCAN(documents, Epsilon, MinNeighbors).Calculate(ctx)
			}
		})
	}
}
