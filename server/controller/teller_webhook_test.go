package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/teller"
	"github.com/stretchr/testify/assert"
)

func TestPostTellerWebhook(t *testing.T) {
	t.Run("disconnection", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		conf.Teller.Enabled = true
		conf.Teller.ApplicationId = gofakeit.Generate("app_????????")
		conf.Teller.Certificate = "bogus.pem"
		conf.Teller.PrivateKey = "bogus.pem"
		conf.Teller.WebhookSigningSecret = []string{
			gofakeit.Generate("????????????????????????????????"),
		}
		app, e := NewTestApplicationWithConfig(t, conf)

		user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveATellerLink(t, app.Clock, user)

		var webhook struct {
			Id      string `json:"id"`
			Payload struct {
				EnrollmentId string `json:"enrollment_id"`
				Reason       string `json:"reason"`
			} `json:"payload"`
			Timestamp time.Time `json:"timestamp"`
			Type      string    `json:"type"`
		}
		webhook.Id = gofakeit.Generate("wh_????????")
		webhook.Payload.EnrollmentId = link.TellerLink.EnrollmentId
		webhook.Payload.Reason = "disconnected"
		webhook.Timestamp = app.Clock.Now()
		webhook.Type = "enrollment.disconnected"

		body, err := json.Marshal(webhook)
		assert.NoError(t, err, "must convert webhook body to json")

		timestamp := app.Clock.Now()
		signatures := teller.GenerateWebhookSignatures(
			timestamp,
			body,
			conf.Teller.WebhookSigningSecret,
		)

		parts := []string{
			fmt.Sprintf("t=%d", timestamp.Unix()),
		}
		for i := range signatures {
			signature := signatures[i]
			parts = append(parts, fmt.Sprintf("v1=%s", signature))
		}

		updatedLink := *link.TellerLink
		updatedLink = testutils.MustDBRead(t, updatedLink)
		assert.Equal(t, models.TellerLinkStatusSetup, updatedLink.Status, "teller link should be setup before webhook")

		{
			response := e.POST("/api/teller/webhook").
				WithJSON(webhook).
				WithHeader("Teller-Signature", strings.Join(parts, ",")).
				Expect()
			response.Status(http.StatusOK)
		}

		updatedLink = testutils.MustDBRead(t, updatedLink)
		assert.Equal(t, models.TellerLinkStatusDisconnected, updatedLink.Status, "teller link should now be disconnected")
	})

	t.Run("invalid signature", func(t *testing.T) {
		conf := NewTestApplicationConfig(t)
		conf.Teller.Enabled = true
		conf.Teller.ApplicationId = gofakeit.Generate("app_????????")
		conf.Teller.Certificate = "bogus.pem"
		conf.Teller.PrivateKey = "bogus.pem"
		conf.Teller.WebhookSigningSecret = []string{
			gofakeit.Generate("????????????????????????????????"),
		}
		app, e := NewTestApplicationWithConfig(t, conf)

		user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveATellerLink(t, app.Clock, user)

		var webhook struct {
			Id      string `json:"id"`
			Payload struct {
				EnrollmentId string `json:"enrollment_id"`
				Reason       string `json:"reason"`
			} `json:"payload"`
			Timestamp time.Time `json:"timestamp"`
			Type      string    `json:"type"`
		}
		webhook.Id = gofakeit.Generate("wh_????????")
		webhook.Payload.EnrollmentId = link.TellerLink.EnrollmentId
		webhook.Payload.Reason = "disconnected"
		webhook.Timestamp = app.Clock.Now()
		webhook.Type = "enrollment.disconnected"

		body, err := json.Marshal(webhook)
		assert.NoError(t, err, "must convert webhook body to json")

		timestamp := app.Clock.Now()
		signatures := teller.GenerateWebhookSignatures(
			timestamp,
			body,
			// Generate the signatures using a different secret from above.
			[]string{
				gofakeit.Generate("????????????????????????????????"),
			},
		)

		parts := []string{
			fmt.Sprintf("t=%d", timestamp.Unix()),
		}
		for i := range signatures {
			signature := signatures[i]
			parts = append(parts, fmt.Sprintf("v1=%s", signature))
		}

		updatedLink := *link.TellerLink
		updatedLink = testutils.MustDBRead(t, updatedLink)
		assert.Equal(t, models.TellerLinkStatusSetup, updatedLink.Status, "teller link should be setup before webhook")

		{
			response := e.POST("/api/teller/webhook").
				WithJSON(webhook).
				WithHeader("Teller-Signature", strings.Join(parts, ",")).
				Expect()
			response.Status(http.StatusUnauthorized)
		}

		updatedLink = testutils.MustDBRead(t, updatedLink)
		assert.Equal(t, models.TellerLinkStatusSetup, updatedLink.Status, "link status should not have changed for a webhook with an invalid signature")
	})
}
