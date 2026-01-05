package mock_plaid

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/stretchr/testify/require"
)

func MockDeactivateItemTokenSuccess(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/item/remove"),
		func(t *testing.T, request *http.Request) (any, int) {
			ValidatePlaidAuthentication(t, request, RequireAccessToken)
			var itemRemoveRequest plaid.ItemRemoveRequest
			require.NoError(t, json.NewDecoder(request.Body).Decode(&itemRemoveRequest), "must decode request")

			return plaid.ItemRemoveResponse{
				RequestId: gofakeit.UUID(),
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}
