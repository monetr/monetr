package platypus

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/mock_plaid"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
)

func TestPlaid_CreateLinkToken(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		mock_plaid.MockCreateLinkToken(t)

		platypus := NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
			OAuthDomain:  "localhost",
		})

		linkToken, err := platypus.CreateLinkToken(context.Background(), LinkTokenOptions{
			ClientUserID:             "1234",
			LegalName:                gofakeit.Name(),
			PhoneNumber:              nil,
			PhoneNumberVerifiedTime:  nil,
			EmailAddress:             gofakeit.Email(),
			EmailAddressVerifiedTime: nil,
			UpdateMode:               false,
		})
		assert.NoError(t, err, "should not return an error creating a link token")
		assert.NotEmpty(t, linkToken.Token(), "must not be empty")
	})
}
