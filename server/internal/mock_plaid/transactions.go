package mock_plaid

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/monetr/monetr/server/internal/myownsanity"
	plaid "github.com/plaid/plaid-go/v49"
	"github.com/stretchr/testify/require"
)

func GenerateTransactions(_ *testing.T, start, end time.Time, numberOfTransactions int, bankAccountIds []string) []*plaid.Transaction {
	transactions := make([]*plaid.Transaction, numberOfTransactions*len(bankAccountIds))
	for i := 0; i < len(transactions); i++ {
		bankAccountId := bankAccountIds[i%len(bankAccountIds)]

		// Should break down transaction dates evenly over the provided range.
		down := end.Add(-(end.Sub(start) / time.Duration(numberOfTransactions*len(bankAccountIds))) * time.Duration(i))
		date := time.Date(down.Year(), down.Month(), down.Day(), 0, 0, 0, 0, time.UTC)

		transactions[i] = &plaid.Transaction{
			Amount: gofakeit.Float64Range(0.99, 100),
			Category: []string{
				"Bank Fees",
			},
			CategoryID:             new("10000000"),
			AccountID:              bankAccountId,
			IsoCurrencyCode:        new("USD"),
			UnofficialCurrencyCode: new("USD"),
			Date:                   date,
			Name:                   gofakeit.Company(),
			TransactionID:          gofakeit.Lexify("?????????????????????"),
		}
	}

	return transactions
}

func MockGetRandomTransactions(t *testing.T, start, end time.Time, numberOfTransactions int, bankAccountIds []string) {
	transactions := GenerateTransactions(t, start, end, numberOfTransactions, bankAccountIds)
	MockGetTransactions(t, transactions)
}

func MockGetTransactions(t *testing.T, transactions []*plaid.Transaction) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/transactions/get"),
		func(t *testing.T, request *http.Request) (any, int) {
			ValidatePlaidAuthentication(t, request, RequireAccessToken)
			var getTransactionsRequest struct {
				Options struct {
					AccountIds []string `json:"account_ids"`
					Count      int      `json:"count"`
					Offset     int      `json:"offset"`
				} `json:"options"`
				StartDate string `json:"start_date"`
				EndDate   string `json:"end_date"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&getTransactionsRequest), "must decode request")

			// Make sure our request dates are valid.
			filterStart, err := time.Parse("2006-01-02", getTransactionsRequest.StartDate)
			require.NoError(t, err, "must provide a valid start date")
			filterEnd, err := time.Parse("2006-01-02", getTransactionsRequest.EndDate)
			require.NoError(t, err, "must provide a valid end date")

			if getTransactionsRequest.Options.Offset > len(transactions) {
				return &plaid.TransactionsGetResponse{}, http.StatusOK
			}

			offset := getTransactionsRequest.Options.Offset
			count := getTransactionsRequest.Options.Count

			filteredTransactions := make([]*plaid.Transaction, 0, len(transactions))
			for _, transaction := range transactions {
				transactionDate := transaction.Date
				if transactionDate.Before(filterEnd.Add(1*time.Second)) && transactionDate.After(filterStart.Add(-1*time.Second)) {
					filteredTransactions = append(filteredTransactions, transaction)
				}
			}

			endingOffset := myownsanity.Min(len(filteredTransactions), offset+count)
			data := filteredTransactions[offset:endingOffset]

			return &plaid.TransactionsGetResponse{
				Accounts:          nil, // Add some basic reporting here too
				Transactions:      data,
				TotalTransactions: len(filteredTransactions),
				Item:              &plaid.Item{},
				RequestID:         gofakeit.UUID(),
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}
