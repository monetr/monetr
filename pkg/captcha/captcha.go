package captcha

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

type Verification interface {
	VerifyCaptcha(ctx context.Context, captcha string) error
}

var (
	_ Verification = &captchaBase{}
)

type captchaBase struct {
	verification recaptcha.ReCAPTCHA
}

func NewReCAPTCHAVerification(privateKey string) (Verification, error) {
	captcha, err := recaptcha.NewReCAPTCHA(
		privateKey,
		recaptcha.V2,
		15*time.Second,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ReCAPTCHA verification")
	}

	return &captchaBase{
		verification: captcha,
	}, nil
}

func (c *captchaBase) VerifyCaptcha(ctx context.Context, captcha string) error {
	span := sentry.StartSpan(ctx, "VerifyCaptcha")
	defer span.Finish()

	return errors.Wrap(c.verification.Verify(captcha), "invalid ReCAPTCHA")
}
