package mock_plaid

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/consts"
	"github.com/monetr/monetr/pkg/internal/mock_http_helper"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
)

func MockCreateLinkToken(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/link/token/create"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			ValidatePlaidAuthentication(t, request, DoNotRequireAccessToken)
			var createLinkTokenRequest plaid.LinkTokenCreateRequest
			require.NoError(t, json.NewDecoder(request.Body).Decode(&createLinkTokenRequest), "must decode request")
			require.Equal(t, consts.PlaidClientName, createLinkTokenRequest.ClientName, "client name must match the shared const")
			require.NotEmpty(t, createLinkTokenRequest.Language, "language is required")

			if createLinkTokenRequest.AccessToken != nil {
				require.Empty(t, createLinkTokenRequest.Products, "products array must be empty when updating a link")
			}

			if createLinkTokenRequest.Webhook != nil {
				webhookUrl, err := url.Parse(*createLinkTokenRequest.Webhook)
				require.NoError(t, err, "webhook URL provided must be valid")
				require.NotEmpty(t, webhookUrl.String(), "webhook URL must be properly parsed")
			}

			return plaid.LinkTokenCreateResponse{
				LinkToken:  gofakeit.UUID(),
				Expiration: time.Now().Add(30 * time.Second),
				RequestId:  gofakeit.UUID(),
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}

func MockCreateLinkTokenFailure(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/link/token/create"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			ValidatePlaidAuthentication(t, request, DoNotRequireAccessToken)
			var createLinkTokenRequest plaid.LinkTokenCreateRequest
			require.NoError(t, json.NewDecoder(request.Body).Decode(&createLinkTokenRequest), "must decode request")
			require.Equal(t, consts.PlaidClientName, createLinkTokenRequest.ClientName, "client name must match the shared const")
			require.NotEmpty(t, createLinkTokenRequest.Language, "language is required")

			if createLinkTokenRequest.AccessToken != nil {
				require.Empty(t, createLinkTokenRequest.Products, "products array must be empty when updating a link")
			}

			return plaid.Error{
				ErrorType:      "API_ERROR",
				ErrorCode:      "INTERNAL_SERVER_ERROR",
				DisplayMessage: *plaid.NewNullableString(myownsanity.StringP("Something went wrong.")),
				RequestId:      myownsanity.StringP(gofakeit.UUID()),
			}, http.StatusInternalServerError
		},
		PlaidHeaders,
	)
}
