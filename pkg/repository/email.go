package repository

import (
	"context"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type EmailVerification = bool

const (
	EmailVerified    EmailVerification = true
	EmailNotVerified EmailVerification = false
)

type EmailRepository interface {
	SetEmailVerified(ctx context.Context, emailAddress string) error
}

var (
	_ EmailRepository = &emailRepositoryBase{}
)

type emailRepositoryBase struct {
	log *logrus.Entry
	db  pg.DBI
}

func NewEmailRepository(log *logrus.Entry, db pg.DBI) EmailRepository {
	return &emailRepositoryBase{
		log: log,
		db:  db,
	}
}

func (e *emailRepositoryBase) SetEmailVerified(ctx context.Context, emailAddress string) error {
	span := sentry.StartSpan(ctx, "SetEmailVerified")
	defer span.Finish()

	var login models.Login
	result, err := e.db.ModelContext(span.Context(), &login).
		Set(`"is_email_verified" = ?`, EmailVerified).       // Change the verification to true.
		Where(`"login"."email" = ?`, strings.ToLower(emailAddress)). // Only for a login with this email.
		Where(`"login"."is_enabled" = ?`, true).                     // Only if the login is actually enabled.
		Where(`"login"."is_email_verified" = ?`, EmailNotVerified).  // And only if the login is not already verified.
		Limit(1).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to verify email")
	}

	if result.RowsAffected() != 1 {
		return errors.New("email cannot be verified")
	}

	return nil
}
