package mock_plaid

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/internal/mock_http_helper"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// MockExchangePublicToken will create an httpmock responder for the development environment of plaid. It returns the
// public token that should be provided in the request. If the request's public token does not match the one returned
// here then an error is returned.
func MockExchangePublicToken(t *testing.T) string {
	publicToken := gofakeit.UUID()

	path := fmt.Sprintf("%s/item/public_token/exchange", plaid.Development)
	mock_http_helper.NewHttpMockJsonResponder(t, "POST", path, func(t *testing.T, request *http.Request) (interface{}, int) {
		var exchangeRequest struct {
			ClientID    string `json:"client_id"`
			Secret      string `json:"secret"`
			PublicToken string `json:"public_token"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&exchangeRequest), "must decode request")

		if exchangeRequest.PublicToken != publicToken {
			return plaid.Error{
				APIResponse: plaid.APIResponse{
					RequestID: gofakeit.UUID(),
				},
				ErrorType:      "INVALID_REQUEST",
				ErrorCode:      "1234",
				ErrorMessage:   "public_token is not valid",
				DisplayMessage: "public_token is not valid",
				StatusCode:     http.StatusBadRequest,
			}, http.StatusBadRequest
		}

		return plaid.ExchangePublicTokenResponse{
			APIResponse: plaid.APIResponse{
				RequestID: gofakeit.UUID(),
			},
			AccessToken: gofakeit.UUID(),
			ItemID:      gofakeit.UUID(),
		}, http.StatusOK
	})

	return publicToken
}
