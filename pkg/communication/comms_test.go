package communication

import (
	"context"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/mock_mail"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUserCommunicationBase_SendPasswordResetEmail(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		smtpMock := mock_mail.NewMockMail()
		options := config.Configuration{
			UIDomainName: "monetr.mini",
			Email: config.Email{
				Domain: "monetr.mini",
			},
		}
		log := testutils.GetLog(t)

		comms := NewUserCommunication(log, options, smtpMock)
		assert.NotNil(t, comms, "communication interface must not be nil")

		params := ForgotPasswordParams{
			Login:    models.Login{
				LoginId:   1234,
				Email:     gofakeit.Email(),
				FirstName: gofakeit.FirstName(),
				LastName:  gofakeit.LastName(),
			},
			ResetURL: fmt.Sprintf("https://app.monetr.mini/reset/%s", gofakeit.Generate("????????????")),
		}

		err := comms.SendPasswordResetEmail(context.Background(), params)
		assert.NoError(t, err, "must send email successfully")
		assert.Len(t, smtpMock.Sent, 1, "should have sent 1 email")
		assert.Equal(t, "no-reply@monetr.mini", smtpMock.Sent[0].From, "from address should be a no-reply")
	})

	t.Run("failure", func(t *testing.T) {
		smtpMock := mock_mail.NewMockMail()
		smtpMock.ShouldFail = true
		options := config.Configuration{
			UIDomainName: "monetr.mini",
			Email: config.Email{
				Domain: "monetr.mini",
			},
		}
		log := testutils.GetLog(t)

		comms := NewUserCommunication(log, options, smtpMock)
		assert.NotNil(t, comms, "communication interface must not be nil")

		params := VerifyEmailParams{
			Login: models.Login{
				LoginId:   1234,
				Email:     gofakeit.Email(),
				FirstName: gofakeit.FirstName(),
				LastName:  gofakeit.LastName(),
			},
			VerifyURL: fmt.Sprintf("https://app.monetr.mini/verify/%s", gofakeit.Generate("????????????")),
		}

		err := comms.SendVerificationEmail(context.Background(), params)
		assert.EqualError(t, err, "failed to send verification email: cannot send email")
		assert.Empty(t, smtpMock.Sent, "should not have sent any emails")
	})
}

func TestUserCommunicationBase_SendVerificationEmail(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		smtpMock := mock_mail.NewMockMail()
		options := config.Configuration{
			UIDomainName: "monetr.mini",
			Email: config.Email{
				Domain: "monetr.mini",
			},
		}
		log := testutils.GetLog(t)

		comms := NewUserCommunication(log, options, smtpMock)
		assert.NotNil(t, comms, "communication interface must not be nil")

		params := VerifyEmailParams{
			Login: models.Login{
				LoginId:   1234,
				Email:     gofakeit.Email(),
				FirstName: gofakeit.FirstName(),
				LastName:  gofakeit.LastName(),
			},
			VerifyURL: fmt.Sprintf("https://app.monetr.mini/verify/%s", gofakeit.Generate("????????????")),
		}

		err := comms.SendVerificationEmail(context.Background(), params)
		assert.NoError(t, err, "must send email successfully")
		assert.Len(t, smtpMock.Sent, 1, "should have sent 1 email")
		assert.Equal(t, "no-reply@monetr.mini", smtpMock.Sent[0].From, "from address should be a no-reply")
	})

	t.Run("failure", func(t *testing.T) {
		smtpMock := mock_mail.NewMockMail()
		smtpMock.ShouldFail = true
		options := config.Configuration{
			UIDomainName: "monetr.mini",
			Email: config.Email{
				Domain: "monetr.mini",
			},
		}
		log := testutils.GetLog(t)

		comms := NewUserCommunication(log, options, smtpMock)
		assert.NotNil(t, comms, "communication interface must not be nil")

		params := VerifyEmailParams{
			Login: models.Login{
				LoginId:   1234,
				Email:     gofakeit.Email(),
				FirstName: gofakeit.FirstName(),
				LastName:  gofakeit.LastName(),
			},
			VerifyURL: fmt.Sprintf("https://app.monetr.mini/verify/%s", gofakeit.Generate("????????????")),
		}

		err := comms.SendVerificationEmail(context.Background(), params)
		assert.EqualError(t, err, "failed to send verification email: cannot send email")
		assert.Empty(t, smtpMock.Sent, "should not have sent any emails")
	})
}
