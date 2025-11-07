package mock_plaid

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/stretchr/testify/require"
)

func MockSync(t *testing.T, transactions []plaid.Transaction) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/transactions/sync"),
		func(t *testing.T, request *http.Request) (any, int) {
			ValidatePlaidAuthentication(t, request, RequireAccessToken)
			var syncTransactionsRequest struct {
				Cursor string `json:"cursor"`
				Count  int    `json:"count"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&syncTransactionsRequest), "must decode request")
			count := syncTransactionsRequest.Count

			filteredTransactions := make([]plaid.Transaction, 0, len(transactions))
			for _, transaction := range transactions {
				filteredTransactions = append(filteredTransactions, transaction)
			}

			endingOffset := myownsanity.Min(len(filteredTransactions), count)
			data := filteredTransactions[0:endingOffset]

			return plaid.TransactionsSyncResponse{
				Accounts:             nil, // Add some basic reporting here too
				Added:                data,
				HasMore:              false,
				RequestId:            gofakeit.UUID(),
				AdditionalProperties: nil,
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}

func MockSyncError(t *testing.T, error plaid.PlaidError) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/transactions/sync"),
		func(t *testing.T, request *http.Request) (any, int) {
			ValidatePlaidAuthentication(t, request, RequireAccessToken)
			var syncTransactionsRequest struct {
				Cursor string `json:"cursor"`
				Count  int    `json:"count"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&syncTransactionsRequest), "must decode request")

			var status int
			if s := error.Status.Get(); s != nil {
				status = int(*s)
			} else {
				status = http.StatusInternalServerError
			}

			return error, status
		},
		PlaidHeaders,
	)
}
