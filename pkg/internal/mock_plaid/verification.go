package mock_plaid

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/rest-api/pkg/internal/mock_http_helper"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
)

func MockGetWebhookVerificationKey(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", Path(t, "/webhook_verification_key/get"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			ValidatePlaidAuthentication(t, request, DoNotRequireAccessToken)
			var requestBody plaid.WebhookVerificationKeyGetRequest
			require.NoError(t, json.NewDecoder(request.Body).Decode(&requestBody), "must decode request")

			// TODO webhooks: Properly implement testing for webhook verification keys.
			// I have little to know idea how to actually implement the key issuing side of this code. I will need to learn
			// how it actually works and then build a mock system off of it. In the mean time I will leave this blank for
			// now.
			return plaid.WebhookVerificationKeyGetResponse{
				Key: plaid.JWKPublicKey{
					Alg:       "",
					Crv:       "",
					Kid:       requestBody.KeyId,
					Kty:       "",
					Use:       "",
					X:         "",
					Y:         "",
					CreatedAt: int32(time.Now().Unix()),
					ExpiredAt: *plaid.NewNullableInt32(myownsanity.Int32P(int32(time.Now().Add(10 * time.Second).Unix()))),
				},
				RequestId:            gofakeit.UUID(),
				AdditionalProperties: nil,
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}
