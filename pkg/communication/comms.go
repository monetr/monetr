package communication

import (
	"bytes"
	"context"
	"fmt"

	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/mail"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/internal/email_templates"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type (
	VerifyEmailParams struct {
		Login     models.Login
		VerifyURL string
	}

	ForgotPasswordParams struct {
		Login    models.Login
		ResetURL string
	}
)

var (
	_ Configuration = config.Configuration{}
)

type Configuration interface {
	GetUIDomainName() string
	GetEmail() config.Email
}

type UserCommunication interface {
	SendVerificationEmail(ctx context.Context, params VerifyEmailParams) error
	SendPasswordResetEmail(ctx context.Context, params ForgotPasswordParams) error
}

type userCommunicationBase struct {
	log     *logrus.Entry
	options Configuration
	mail    mail.Communication
}

func NewUserCommunication(log *logrus.Entry, options Configuration, client mail.Communication) UserCommunication {
	return &userCommunicationBase{
		log:     log,
		options: options,
		mail:    client,
	}
}

func (u *userCommunicationBase) SendVerificationEmail(ctx context.Context, params VerifyEmailParams) error {
	span := sentry.StartSpan(ctx, "SendVerificationEmail")
	defer span.Finish()

	emailContent, err := u.getEmailContent(span.Context(), email_templates.VerifyEmailTemplate, params)
	if err != nil {
		return err
	}

	log := u.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"loginId": params.Login.LoginId,
	})

	log.Debug("sending verification email")

	if err = u.mail.Send(span.Context(), mail.SendEmailRequest{
		From:    fmt.Sprintf("no-reply@%s", u.options.GetEmail().Domain),
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

func (u *userCommunicationBase) SendPasswordResetEmail(ctx context.Context, params ForgotPasswordParams) error {
	span := sentry.StartSpan(ctx, "SendPasswordResetEmail")
	defer span.Finish()

	emailContent, err := u.getEmailContent(span.Context(), email_templates.ForgotPasswordTemplate, params)
	if err != nil {
		return err
	}

	log := u.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"loginId": params.Login.LoginId,
	})

	log.Debug("sending password reset email")

	if err = u.mail.Send(span.Context(), mail.SendEmailRequest{
		From:    fmt.Sprintf("no-reply@%s", u.options.GetEmail().Domain),
		To:      params.Login.Email,
		Subject: "Password Reset",
		Content: emailContent,
		IsHTML:  true,
	}); err != nil {
		log.WithError(err).Error("failed to send password reset email")
		return errors.Wrap(err, "failed to send password reset email")
	}

	return nil
}

func (u *userCommunicationBase) getEmailContent(ctx context.Context, templateName string, params interface{}) (string, error) {
	span := sentry.StartSpan(ctx, "getEmailContent")
	defer span.Finish()

	log := u.log.WithContext(span.Context()).WithField("emailTemplate", templateName)

	verifyTemplate, err := email_templates.GetEmailTemplate(templateName)
	if err != nil {
		log.WithError(err).Error("failed to retrieve email template")
		return "", errors.Wrap(err, "failed to retrieve email template")
	}

	buffer := bytes.NewBuffer(nil)

	if err = verifyTemplate.Execute(buffer, params); err != nil {
		log.WithError(err).Error("failed to execute verification email template")
		return "", errors.Wrap(err, "failed to execute verification email template")
	}

	return buffer.String(), nil
}
