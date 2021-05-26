package plaid_helper

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetrapp/rest-api/pkg/internal/mock_plaid"
	"github.com/monetrapp/rest-api/pkg/internal/testutils"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPlaidClient_GetAccounts(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		account := plaid.Account{
			AccountID: gofakeit.UUID(),
			Balances: plaid.AccountBalances{
				Available:              10.00,
				Current:                10.00,
				Limit:                  10.00,
				ISOCurrencyCode:        "USD",
				UnofficialCurrencyCode: "USD",
			},
			Mask:               "1234",
			Name:               "Checking account",
			OfficialName:       "Super duper checking",
			Subtype:            "checking",
			Type:               "depository",
			VerificationStatus: "verified?",
		}

		mock_plaid.MockGetAccounts(t, []plaid.Account{
			account,
		})

		client := NewPlaidClient(testutils.GetLog(t), plaid.ClientOptions{
			ClientID:    gofakeit.UUID(),
			Secret:      gofakeit.UUID(),
			Environment: plaid.Sandbox,
			HTTPClient: &http.Client{
				Transport: httpmock.DefaultTransport,
			},
		})

		result, err := client.GetAccounts(context.Background(), gofakeit.UUID(), plaid.GetAccountsOptions{
			AccountIDs: []string{
				account.AccountID,
			},
		})
		assert.NoError(t, err, "should succeed")
		assert.NotEmpty(t, result)
	})

	t.Run("error NO_ACCOUNTS", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		mock_plaid.MockGetAccountsError(t, plaid.Error{
			APIResponse: plaid.APIResponse{
				RequestID: gofakeit.UUID(),
			},
			ErrorType:      "ITEM_ERROR",
			ErrorCode:      "NO_ACCOUNTS",
			ErrorMessage:   "no valid accounts were found for this item",
			DisplayMessage: "No valid accounts were found at the financial institution. Please visit your financial institution's website to confirm accounts are available.",
			StatusCode:     http.StatusBadRequest,
		})

		client := NewPlaidClient(testutils.GetLog(t), plaid.ClientOptions{
			ClientID:    gofakeit.UUID(),
			Secret:      gofakeit.UUID(),
			Environment: plaid.Sandbox,
			HTTPClient: &http.Client{
				Transport: httpmock.DefaultTransport,
			},
		})

		result, err := client.GetAccounts(context.Background(), gofakeit.UUID(), plaid.GetAccountsOptions{
			AccountIDs: []string{
				gofakeit.UUID(),
			},
		})
		assert.Error(t, err, "should fail with NO_ACCOUNTS error")
		assert.Empty(t, result)

		cause := errors.Cause(err)
		assert.IsType(t, plaid.Error{}, cause, "should be a plaid error")
	})
}
