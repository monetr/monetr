package mock_plaid

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/mock_http_helper"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
)

func GenerateTransactions(t *testing.T, start, end time.Time, numberOfTransactions int, bankAccountIds []string) []plaid.Transaction {
	transactions := make([]plaid.Transaction, numberOfTransactions*len(bankAccountIds))
	for i := 0; i < len(transactions); i++ {
		bankAccountId := bankAccountIds[i%len(bankAccountIds)]

		transaction := plaid.Transaction{}
		transaction.SetAmount(gofakeit.Float32Range(0.99, 100))
		transaction.SetCategory([]string{
			"Bank Fees",
		})
		transaction.SetCategoryId("10000000")
		transaction.SetAccountId(bankAccountId)
		transaction.SetIsoCurrencyCode("USD")
		transaction.SetUnofficialCurrencyCode("USD")

		// Should break down transaction dates evenly over the provided range.
		down := end.Add(-(end.Sub(start) / time.Duration(numberOfTransactions*len(bankAccountIds))) * time.Duration(i))

		transaction.SetDate(down.Format("2006-01-02"))
		transaction.SetName(gofakeit.Company())
		transaction.SetTransactionId(gofakeit.Generate("?????????????????????"))

		transactions[i] = transaction
	}

	return transactions
}

func MockGetRandomTransactions(t *testing.T, start, end time.Time, numberOfTransactions int, bankAccountIds []string) {
	transactions := GenerateTransactions(t, start, end, numberOfTransactions, bankAccountIds)
	MockGetTransactions(t, transactions)
}

func MockGetTransactions(t *testing.T, transactions []plaid.Transaction) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/transactions/get"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
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
				return plaid.TransactionsGetResponse{}, http.StatusOK
			}

			offset := getTransactionsRequest.Options.Offset
			count := getTransactionsRequest.Options.Count

			filteredTransactions := make([]plaid.Transaction, 0, len(transactions))
			for _, transaction := range transactions {
				transactionDate, err := time.Parse("2006-01-02", transaction.GetDate())
				require.NoError(t, err, "must be able to parse transaction date")
				if transactionDate.Before(filterEnd.Add(1*time.Second)) && transactionDate.After(filterStart.Add(-1*time.Second)) {
					filteredTransactions = append(filteredTransactions, transaction)
				}
			}

			endingOffset := myownsanity.Min(len(filteredTransactions), offset+count)
			data := filteredTransactions[offset:endingOffset]

			return plaid.TransactionsGetResponse{
				Accounts:             nil, // Add some basic reporting here too
				Transactions:         data,
				TotalTransactions:    int32(len(filteredTransactions)),
				Item:                 plaid.Item{},
				RequestId:            gofakeit.UUID(),
				AdditionalProperties: nil,
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}
