package mock_plaid

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/mock_http_helper"
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
			require.NotEmpty(t, createLinkTokenRequest.ClientName, "client name is required")
			require.NotEmpty(t, createLinkTokenRequest.Language, "language is required")

			if createLinkTokenRequest.AccessToken != nil {
				require.Empty(t, createLinkTokenRequest.Products, "products array must be empty when updating a link")
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
