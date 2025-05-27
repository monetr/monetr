package captcha

import (
	"context"
	"errors"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/validation"
)

var (
	_ validation.Rule            = &captchaValidator{}
	_ validation.RuleWithContext = &captchaValidator{}
)

type captchaValidator struct {
	verification Verification
}

// ValidateWithContext implements validation.RuleWithContext.
func (c *captchaValidator) ValidateWithContext(ctx context.Context, value any) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	str, ok := value.(string)
	if !ok {
		return errors.New("ReCAPTCHA value must be a valid string")
	}

	if err := c.verification.VerifyCaptcha(span.Context(), str); err != nil {
		return errors.New("ReCAPTCHA is not valid")
	}

	return nil
}

// Validate implements validation.Rule.
func (c *captchaValidator) Validate(value any) error {
	return c.ValidateWithContext(context.Background(), value)
}

func Validate(verification Verification) validation.Rule {
	return &captchaValidator{
		verification: verification,
	}
}
