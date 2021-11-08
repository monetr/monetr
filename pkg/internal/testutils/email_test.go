package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUniqueEmail(t *testing.T) {
	numberOfEmails := 1000
	emails := map[string]struct{}{}
	for i := 0; i < numberOfEmails; i++ {
		email := GetUniqueEmail(t)
		assert.NotEmpty(t, email, "generated email must not be blank")
		assert.NotContains(t, emails, email, "email must not have been seen before")
		emails[email] = struct{}{}
	}
	assert.Len(t, emails, numberOfEmails, "should have the expected total for unique emails")
}
