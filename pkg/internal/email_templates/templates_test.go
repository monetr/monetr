package email_templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEmailTemplate(t *testing.T) {
	t.Run("verify email", func(t *testing.T) {
		verifyEmailTemplate, err := GetEmailTemplate(VerifyEmailTemplate)
		assert.NoError(t, err, "should succeed")
		assert.NotNil(t, verifyEmailTemplate, "should return a valid template")
	})

	t.Run("forgot password", func(t *testing.T) {
		forgotPasswordTemplate, err := GetEmailTemplate(ForgotPasswordTemplate)
		assert.NoError(t, err, "should succeed")
		assert.NotNil(t, forgotPasswordTemplate, "should return a valid template")
	})

	t.Run("missing template", func(t *testing.T) {
		verifyEmailTemplate, err := GetEmailTemplate("templates/i_dont_exist.html")
		assert.EqualError(t, err, "failed to open email template (templates/i_dont_exist.html): open templates/i_dont_exist.html: file does not exist")
		assert.Nil(t, verifyEmailTemplate, "should not return a template if the file is missing")
	})
}
