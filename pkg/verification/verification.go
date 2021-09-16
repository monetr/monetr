package verification

import (
	"context"

	"github.com/monetr/rest-api/pkg/repository"
	"github.com/sirupsen/logrus"
)

type Verification interface {
	CreateEmailVerificationToken(ctx context.Context, emailAddress string) (token string, _ error)
	UseEmailVerificationToken(ctx context.Context, token string)
}

type verificationBase struct {
	log       *logrus.Entry
	emailRepo repository.EmailRepository
	tokens    EmailVerificationTokenGenerator
}
