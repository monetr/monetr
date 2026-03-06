package mock_plaid

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/stretchr/testify/require"
)

func MockGetWebhookVerificationKey(t *testing.T, clk clock.Clock, kid string, publicKey *ecdsa.PublicKey) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", Path(t, "/webhook_verification_key/get"),
		func(t *testing.T, request *http.Request) (any, int) {
			ValidatePlaidAuthentication(t, request, DoNotRequireAccessToken)
			var requestBody plaid.WebhookVerificationKeyGetRequest
			require.NoError(t, json.NewDecoder(request.Body).Decode(&requestBody), "must decode request")
			require.Equal(t, kid, requestBody.KeyId, "key ID in request must match expected kid")

			x := base64.RawURLEncoding.EncodeToString(publicKey.X.Bytes())
			y := base64.RawURLEncoding.EncodeToString(publicKey.Y.Bytes())

			return plaid.WebhookVerificationKeyGetResponse{
				Key: plaid.JWKPublicKey{
					Alg:       "ES256",
					Crv:       "P-256",
					Kid:       kid,
					Kty:       "EC",
					Use:       "sig",
					X:         x,
					Y:         y,
					CreatedAt: int32(clk.Now().Unix()),
					ExpiredAt: *plaid.NewNullableInt32(myownsanity.Pointer(int32(clk.Now().Add(30 * time.Minute).Unix()))),
				},
				RequestId:            gofakeit.UUID(),
				AdditionalProperties: nil,
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}

func MockGetWebhookVerificationKeyFailure(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(t,
		"POST", Path(t, "/webhook_verification_key/get"),
		func(t *testing.T, request *http.Request) (any, int) {
			ValidatePlaidAuthentication(t, request, DoNotRequireAccessToken)
			var requestBody plaid.WebhookVerificationKeyGetRequest
			require.NoError(t, json.NewDecoder(request.Body).Decode(&requestBody), "must decode request")

			return plaid.PlaidError{
				ErrorType:      "API_ERROR",
				ErrorCode:      "INTERNAL_SERVER_ERROR",
				DisplayMessage: *plaid.NewNullableString(myownsanity.Pointer("Something went wrong.")),
				RequestId:      myownsanity.Pointer(gofakeit.UUID()),
			}, http.StatusInternalServerError
		},
		PlaidHeaders,
	)
}

func SignWebhookJWT(t *testing.T, clk clock.Clock, privateKey *ecdsa.PrivateKey, kid string) string {
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(clk.Now()),
		ExpiresAt: jwt.NewNumericDate(clk.Now().Add(5 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(privateKey)
	require.NoError(t, err, "must sign webhook JWT")
	return signed
}
