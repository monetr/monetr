package mock_plaid

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/rest-api/pkg/internal/mock_http_helper"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func MockGetWebhookVerificationKey(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", Path(t, "/webhook_verification_key/get"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			ValidatePlaidAuthentication(t, request, DoNotRequireAccessToken)
			var getWebhookVerificationKeyRequest struct {
				KeyId string `json:"kid"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&getWebhookVerificationKeyRequest), "must decode request")

			return plaid.WebhookVerificationKeyGetResponse{
				Key: plaid.JWKPublicKey{
					Alg:       "",
					Crv:       "",
					Kid:       "",
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
