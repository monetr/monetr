package platypus

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
)

func TestPlaidTransaction_GetDates(t *testing.T) {
	transaction := PlaidTransaction{
		Amount:        128,
		BankAccountId: gofakeit.Generate("?????"),
		Category: []string{
			"Bank Fee",
		},
		Date:                   time.Date(2021, 9, 16, 0, 0, 0, 0, time.UTC),
		ISOCurrencyCode:        "USD",
		UnofficialCurrencyCode: "USD",
		IsPending:              false,
		MerchantName:           "Arby's",
		Name:                   "Arby's",
		OriginalDescription:    "ARBYS",
		PendingTransactionId:   nil,
		TransactionId:          gofakeit.Generate("??????????"),
	}

	timezone, err := time.LoadLocation("America/Chicago")
	assert.NoError(t, err, "must retrieve timezone")
	assert.NotNil(t, timezone, "timezone cannot be nil")

	t.Run("GetDate", func(t *testing.T) {
		assert.Equal(t,
			"2021-09-16T00:00:00Z", transaction.GetDate().Format(time.RFC3339Nano),
			"should match value without transforming timezone",
		)
	})

	t.Run("GetDateLocal", func(t *testing.T) {
		assert.Equal(t,
			"2021-09-16T00:00:00-05:00", transaction.GetDateLocal(timezone).Format(time.RFC3339Nano),
			"should match value when transforming timezone",
		)
	})
}
