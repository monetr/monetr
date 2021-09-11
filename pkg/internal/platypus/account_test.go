package platypus

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
)

func TestNewPlaidBankAccountBalances(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var available, current, limit float32
		available = 101.34
		current = 110.25
		limit = 500.13

		plaidBalances := plaid.AccountBalance{
			Available:              *plaid.NewNullableFloat32(&available),
			Current:                *plaid.NewNullableFloat32(&current),
			Limit:                  *plaid.NewNullableFloat32(&limit),
			IsoCurrencyCode:        *plaid.NewNullableString(myownsanity.StringP("USD")),
			UnofficialCurrencyCode: *plaid.NewNullableString(myownsanity.StringP("USD")),
			LastUpdatedDatetime:    plaid.NullableTime{}, // Leave this be so that is has no value.
			AdditionalProperties:   map[string]interface{}{},
		}

		balances, err := NewPlaidBankAccountBalances(plaidBalances)
		assert.NoError(t, err, "must be able to convert balances")
		assert.NotEmpty(t, balances, "balances must not be empty")
		assert.EqualValues(t, available*100, balances.GetAvailable(), "available should be converted to cents")
		assert.EqualValues(t, current*100, balances.GetCurrent(), "current should be converted to cents")
		assert.EqualValues(t, limit*100, balances.GetLimit(), "limit should be converted to cents")
		assert.EqualValues(t, "USD", balances.GetIsoCurrencyCode(), "ISO currency code should match USD")
		assert.EqualValues(t, "USD", balances.GetUnofficialCurrencyCode(), "unofficial currency code should match USD")
	})

	t.Run("missing value", func(t *testing.T) {
		plaidBalances := plaid.AccountBalance{
			Available:              *plaid.NewNullableFloat32(nil),
			Current:                *plaid.NewNullableFloat32(nil),
			Limit:                  *plaid.NewNullableFloat32(nil),
			IsoCurrencyCode:        *plaid.NewNullableString(nil),
			UnofficialCurrencyCode: *plaid.NewNullableString(nil),
		}

		balances, err := NewPlaidBankAccountBalances(plaidBalances)
		assert.NoError(t, err, "must be able to convert balances")
		assert.EqualValues(t, 0, balances.GetAvailable(), "available should be 0 when no value is present")
		assert.EqualValues(t, 0, balances.GetCurrent(), "current should be 0 when no value is present")
		assert.EqualValues(t, 0, balances.GetLimit(), "limit should be 0 when no value is present")
		assert.EqualValues(t, "", balances.GetIsoCurrencyCode(), "ISO currency code should be empty if no value is present")
		assert.EqualValues(t, "", balances.GetUnofficialCurrencyCode(), "unofficial currency code should be empty if no value is present")
	})
}

func TestNewPlaidBankAccount(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		subType := plaid.ACCOUNTSUBTYPE_CHECKING
		plaidBank := plaid.AccountBase{
			AccountId:    gofakeit.UUID(),
			Balances:     plaid.AccountBalance{},
			Mask:         *plaid.NewNullableString(myownsanity.StringP("1234")),
			Name:         "Checking Account",
			OfficialName: *plaid.NewNullableString(myownsanity.StringP("CHECKING - 1234")),
			Type:         plaid.ACCOUNTTYPE_DEPOSITORY,
			Subtype:      *plaid.NewNullableAccountSubtype(&subType),
		}

		bank, err := NewPlaidBankAccount(plaidBank)
		assert.NoError(t, err, "must be able to convert bank account")
		assert.NotEmpty(t, bank, "bank account must not be empty")
		assert.EqualValues(t, plaidBank.GetAccountId(), bank.GetAccountId(), "account Id must match")
		assert.EqualValues(t, "1234", bank.GetMask(), "mask must match")
		assert.EqualValues(t, "Checking Account", bank.GetName(), "name must match")
		assert.EqualValues(t, "CHECKING - 1234", bank.GetOfficialName(), "official name must match")
		assert.EqualValues(t, "depository", bank.GetType(), "account type must match")
		assert.EqualValues(t, "checking", bank.GetSubType(), "account sub-type must match")

		assert.IsType(t, PlaidBankAccountBalances{}, bank.GetBalances(), "must return plaid bank account balances")
	})
}
