package identification_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/identification"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("test a few", func(t *testing.T) {
		kinds := []identification.Kind{
			identification.LoginKind,
			identification.UserKind,
			identification.AccountKind,
			identification.LinkKind,
			identification.PlaidLinkKind,
			identification.TellerLinkKind,
			identification.BankAccountKind,
			identification.PlaidBankAccountKind,
			identification.TellerBankAccountKind,
			identification.PlaidSyncKind,
			identification.TellerSyncKind,
			identification.TransactionKind,
			identification.PlaidTransactionKind,
			identification.TellerTransactionKind,
			identification.TransactionClusterKind,
			identification.SecretKind,
			identification.SpendingKind,
			identification.FundingScheduleKind,
			identification.FileKind,
			identification.CronJobKind,
			identification.JobKind,
			identification.BetaKind,
		}

		for i := range kinds {
			kind := kinds[i]

			id := identification.New(kind)
			outKind := id.Kind()
			assert.Equal(t, kind, outKind, "kinds should match!")
		}
	})
}
