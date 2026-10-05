package similar

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimilarTransactions_TFIDF_DBSCAN(t *testing.T) {
	t.Run("monetr mercury dataset", func(t *testing.T) {
		fixtureJson := fixtures.LoadFile(t, "monetr_sample_data_1.json")
		var data []models.Transaction
		require.NoError(t, json.Unmarshal(fixtureJson, &data), "must be able to decode fixture data")
		log := testutils.GetLog(t)
		detector := NewSimilarTransactions_TFIDF_DBSCAN(log)

		for i := range data {
			detector.AddTransaction(&data[i])
		}

		groups := detector.DetectSimilarTransactions(t.Context())
		assert.NotEmpty(t, groups, "must return an array of groups of similar transactions")
		for _, group := range groups {
			assert.NotEmpty(t, group.Members, "a groups matches should not be empty!")
			assert.NotEmpty(t, group.Name, "a groups name should not be empty!")
			assert.NotEmpty(t, group.Signature, "a groups signature should not be empty!")
		}
		// TODO, add specific assertions here about what the groups are.
		j, _ := json.MarshalIndent(groups, "", "  ")
		fmt.Println(string(j))
	})

	t.Run("name word missing from the centroid", func(t *testing.T) {
		// The payroll deposits are a minority of the cluster, but payroll ranks
		// highest so it is the only word in the name. The centroid is one of the
		// plain deposits which doesn't have payroll in it, so the original word for
		// payroll has to come from one of the other members instead.
		data := make([]models.Transaction, 0, 240)
		add := func(count int, name, merchant string) {
			for range count {
				data = append(data, models.Transaction{
					TransactionId:        models.ID[models.Transaction](fmt.Sprintf("txn_%03d", len(data))),
					OriginalName:         name,
					OriginalMerchantName: merchant,
				})
			}
		}
		add(22, "Deposit 5176 Treasury Pr  Deposit", "Deposit 5176 Treasury Pr Deposit")
		add(11, "Deposit 5176 Treasury Pr  Payroll", "Deposit 5176 Treasury Pr Payroll")
		// The merchant names sometimes get cut off, these are what bridge the plain
		// deposits and the payroll deposits into a single cluster.
		add(1, "Deposit 5176 Treasury Pr  Deposit", "Deposit 5176 Treasury Pr  Deposi")
		add(1, "Deposit 5176 Treasury Pr  Payroll", "Deposit 5176 Treasury Pr  Payrol")
		// Other transactions in the account so that treasury and deposit aren't in
		// every single transaction.
		for _, merchant := range []string{
			"Sentry", "Github", "Discord", "Plaid", "Twilio", "Freshbooks", "Google", "Lunchflow",
		} {
			name := "ACH Debit - Pwp " + merchant + " Privacycom 2111508"
			add(20, name, name)
		}
		add(20, "POS DEBIT-BDC 2069 PLOVER DIGITAL PLOVER", "POS DEBIT-BDC 2069 PLOVER DIGITAL PLOVER")
		add(25, "Dividend", "Dividend")

		log := testutils.GetLog(t)
		detector := NewSimilarTransactions_TFIDF_DBSCAN(log)
		for i := range data {
			detector.AddTransaction(&data[i])
		}

		groups := detector.DetectSimilarTransactions(t.Context())
		var treasury *models.TransactionCluster
		for i := range groups {
			assert.NotEmpty(t, groups[i].Name, "a groups name should not be empty!")
			if len(groups[i].Members) == 35 {
				treasury = &groups[i]
			}
		}
		require.NotNil(t, treasury, "the deposits and payroll deposits must be clustered together")
		assert.Equal(t, "Payroll", treasury.Name)
	})

	t.Run("amazon dataset", func(t *testing.T) {
		fixtureJson := fixtures.LoadFile(t, "amazon_sample_data_1.json")
		var data []models.Transaction
		require.NoError(t, json.Unmarshal(fixtureJson, &data), "must be able to decode fixture data")
		log := testutils.GetLog(t)
		detector := NewSimilarTransactions_TFIDF_DBSCAN(log)

		for i := range data {
			detector.AddTransaction(&data[i])
		}

		groups := detector.DetectSimilarTransactions(t.Context())
		assert.NotEmpty(t, groups, "must return an array of groups of similar transactions")

		j, _ := json.MarshalIndent(groups, "", "  ")
		fmt.Println(string(j))
	})
}
