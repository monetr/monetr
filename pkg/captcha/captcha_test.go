package captcha

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/pkg/internal/mock_http_helper"
	"github.com/stretchr/testify/assert"
)

func TestNewReCAPTCHAVerification(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		verification, err := NewReCAPTCHAVerification("")
		assert.Error(t, err, "should fail to create a verifier without a private key")
		assert.Nil(t, verification, "verification should not be returned if there was a problem")
	})
}

func TestCaptchaBase_VerifyCaptcha(t *testing.T) {
	t.Run("mock success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_http_helper.NewHttpMockJsonResponder(t,
			"POST", "https://www.google.com/recaptcha/api/siteverify",
			func(t *testing.T, request *http.Request) (interface{}, int) {
				return map[string]interface{}{
					"success":      true,
					"challenge_ts": time.Now(),
					"hostname":     "monetr.mini",
					"score":        1.0,
				}, http.StatusOK
			},
			nil,
		)

		verification, err := NewReCAPTCHAVerification("test")
		assert.NoError(t, err, "must be able to create captcha verification")

		err = verification.VerifyCaptcha(context.Background(), "abc123")
		assert.NoError(t, err, "must verify captcha")
	})

	t.Run("mock failure", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_http_helper.NewHttpMockJsonResponder(t,
			"POST", "https://www.google.com/recaptcha/api/siteverify",
			func(t *testing.T, request *http.Request) (interface{}, int) {
				return map[string]interface{}{
					"success":      false,
					"challenge_ts": time.Now(),
					"hostname":     "monetr.mini",
					"score":        0,
				}, http.StatusOK
			},
			nil,
		)

		verification, err := NewReCAPTCHAVerification("test")
		assert.NoError(t, err, "must be able to create captcha verification")

		err = verification.VerifyCaptcha(context.Background(), "abc123")
		assert.EqualError(t, err, "invalid ReCAPTCHA: invalid challenge solution")
	})
}
