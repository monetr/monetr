package communication

import (
	"bytes"
	"context"
	"fmt"

	"github.com/monetr/rest-api/pkg/config"
	"github.com/monetr/rest-api/pkg/mail"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/internal/email_templates"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type VerifyEmailParams struct {
	Login     models.Login
	VerifyURL string
}

type UserCommunication interface {
	SendVerificationEmail(ctx context.Context, params VerifyEmailParams) error
}

type userCommunicationBase struct {
	log     *logrus.Entry
	options config.Email
	mail    mail.Communication
}

func NewUserCommunication(log *logrus.Entry, options config.Email, client mail.Communication) UserCommunication {
	return &userCommunicationBase{
		log:     log,
		options: options,
		mail:    client,
	}
}

func (u *userCommunicationBase) SendVerificationEmail(ctx context.Context, params VerifyEmailParams) error {
	span := sentry.StartSpan(ctx, "SendVerificationEmail")
	defer span.Finish()

	emailContent, err := u.getVerificationEmailContent(span.Context(), params)
	if err != nil {
		return err
	}

	log := u.log.WithContext(ctx).WithFields(logrus.Fields{
		"loginId": params.Login.LoginId,
	})

	log.Debug("sending verification email")

	if err = u.mail.Send(span.Context(), mail.SendEmailRequest{
		From:    fmt.Sprintf("no-reply@%s", u.options.Domain),
		To:      params.Login.Email,
		Subject: "Verify Your Email Address",
		Content: emailContent,
		IsHTML:  true,
	}); err != nil {
		log.WithError(err).Error("failed to send verification email")
		return errors.Wrap(err, "failed to send verification email")
	}

	return nil
}

func (u *userCommunicationBase) getVerificationEmailContent(ctx context.Context, params VerifyEmailParams) (string, error) {
	span := sentry.StartSpan(ctx, "getVerificationEmailContent")
	defer span.Finish()

	log := u.log.WithContext(ctx).WithFields(logrus.Fields{
		"loginId": params.Login.LoginId,
	})

	verifyTemplate, err := email_templates.GetEmailTemplate(email_templates.VerifyEmailTemplate)
	if err != nil {
		log.WithError(err).Error("failed to retrieve verification email template")
		return "", errors.Wrap(err, "failed to retrieve verification email template")
	}

	buffer := bytes.NewBuffer(nil)

	if err = verifyTemplate.Execute(buffer, params); err != nil {
		log.WithError(err).Error("failed to execute verification email template")
		return "", errors.Wrap(err, "failed to execute verification email template")
	}

	return buffer.String(), nil
}
