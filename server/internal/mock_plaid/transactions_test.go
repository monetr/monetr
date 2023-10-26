package mock_plaid

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTransactions(t *testing.T) {
	t.Run("2 days", func(t *testing.T) {
		numberOfTransactions := 600
		bankAccounts := []string{
			"1234",
			"5678",
		}
		end := time.Now()
		start := time.Now().Add(-25 * time.Hour)
		transactions := GenerateTransactions(t, start, end, numberOfTransactions, bankAccounts)
		assert.Len(t, transactions, len(bankAccounts)*numberOfTransactions)
		assert.Equal(t, end.Format("2006-01-02"), transactions[0].GetDate(), "date of first transaction should be end")
		assert.Equal(t, start.Format("2006-01-02"), transactions[len(transactions)-1].GetDate(), "date of last transaction should be start")
	})

	t.Run("30 days", func(t *testing.T) {
		numberOfTransactions := 3000
		bankAccounts := []string{
			"1234",
			"5678",
		}
		end := time.Now().UTC().Truncate(time.Hour)
		start := time.Now().UTC().Add(-30 * 24 * time.Hour).Truncate(time.Hour)
		transactions := GenerateTransactions(t, start, end, numberOfTransactions, bankAccounts)
		assert.Len(t, transactions, len(bankAccounts)*numberOfTransactions)
		assert.Equal(t, end.Format("2006-01-02"), transactions[0].GetDate(), "date of first transaction should be end")
		assert.Equal(t, start.Format("2006-01-02"), transactions[len(transactions)-1].GetDate(), "date of last transaction should be start")
	})
}
