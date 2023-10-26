package mock_plaid

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/plaid/plaid-go/v14/plaid"
	"github.com/stretchr/testify/require"
)

// MockExchangePublicToken will create an httpmock responder for the development environment of plaid. It returns the
// public token that should be provided in the request. If the request's public token does not match the one returned
// here then an error is returned.
func MockExchangePublicToken(t *testing.T) string {
	publicToken := gofakeit.UUID()

	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/item/public_token/exchange"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			ValidatePlaidAuthentication(t, request, DoNotRequireAccessToken)
			var exchangeRequest struct {
				PublicToken string `json:"public_token"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&exchangeRequest), "must decode request")

			requestId := gofakeit.UUID()
			if exchangeRequest.PublicToken != publicToken {
				return plaid.PlaidError{
					RequestId:      &requestId,
					ErrorType:      "INVALID_REQUEST",
					ErrorCode:      "1234",
					ErrorMessage:   "public_token is not valid",
					DisplayMessage: *plaid.NewNullableString(myownsanity.StringP("public_token is not valid")),
					Status:         *plaid.NewNullableInt32(myownsanity.Int32P(http.StatusBadRequest)),
				}, http.StatusBadRequest
			}

			return plaid.ItemPublicTokenExchangeResponse{
				RequestId:   requestId,
				AccessToken: gofakeit.UUID(),
				ItemId:      gofakeit.UUID(),
			}, http.StatusOK
		},
		PlaidHeaders,
	)

	return publicToken
}
