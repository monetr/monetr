package verification

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Verification interface {
	// CreateEmailVerificationToken is a stateless method that will generate a token for the provided email address
	// regardless of whether that email address is even associated with a login in the application.
	CreateEmailVerificationToken(ctx context.Context, emailAddress string) (token string, _ error)
	// RegenerateEmailVerificationToken is similar to CreateEmailVerificationToken, except that it requires that the
	// provided email address is 1. Not already verified, 2. Associated with a valid login. It will return an error if
	// both of these conditions are not met.
	RegenerateEmailVerificationToken(ctx context.Context, emailAddress string) (login *models.Login, token string, _ error)
	// UseEmailVerificationToken consumes a provided token, and will associate it with a specific login in the
	// application. If there is no login that the token can be associated with (as in the email for the token does not
	// exist in the application) then an error is returned. If the token provided is both valid, and is associated with
	// a login in the application (and that login is not already verified) then the method will succeed and return
	// successfully.
	UseEmailVerificationToken(ctx context.Context, token string) error
}

var (
	_ Verification = &verificationBase{}
)

type verificationBase struct {
	log       *logrus.Entry
	lifespan  time.Duration
	emailRepo repository.EmailRepository
	tokens    TokenGenerator
}

func NewEmailVerification(log *logrus.Entry, lifespan time.Duration, emailRepo repository.EmailRepository, tokens TokenGenerator) Verification {
	return &verificationBase{
		log:       log,
		lifespan:  lifespan,
		emailRepo: emailRepo,
		tokens:    tokens,
	}
}

func (v *verificationBase) CreateEmailVerificationToken(ctx context.Context, emailAddress string) (token string, err error) {
	span := sentry.StartSpan(ctx, "CreateEmailVerificationToken")
	defer span.Finish()

	return v.tokens.GenerateToken(span.Context(), emailAddress, v.lifespan)
}

func (v *verificationBase) RegenerateEmailVerificationToken(ctx context.Context, emailAddress string) (_ *models.Login, token string, err error) {
	span := sentry.StartSpan(ctx, "CreateEmailVerificationToken")
	defer span.Finish()

	login, err := v.emailRepo.GetLoginForEmail(span.Context(), emailAddress)
	if err != nil {
		return nil, "", err
	}

	if login.GetEmailIsVerified() {
		return nil, "", errors.New("email is already verified")
	}

	token, err = v.tokens.GenerateToken(span.Context(), emailAddress, v.lifespan)
	return login, token, err
}

func (v *verificationBase) UseEmailVerificationToken(ctx context.Context, token string) error {
	span := sentry.StartSpan(ctx, "UseEmailVerificationToken")
	defer span.Finish()

	// Verify the token is valid and get the email address out of it.
	emailAddress, err := v.tokens.ValidateToken(span.Context(), token)
	if err != nil {
		span.Status = sentry.SpanStatusInvalidArgument
		return errors.Wrap(err, "could not verify email")
	}

	// Then set that email as verified.
	if err = v.emailRepo.SetEmailVerified(span.Context(), emailAddress); err != nil {
		// If there is a problem then that means the token is not valid or the email has already been verified.
		span.Status = sentry.SpanStatusNotFound
		return errors.Wrap(err, "failed to mark email as verified")
	}

	return nil
}
