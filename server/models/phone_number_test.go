package models

import (
	"github.com/nyaruka/phonenumbers"
	"github.com/stretchr/testify/require"
	"testing"
)

func mustParsePhoneNumber(t *testing.T, number string) {
	num, err := phonenumbers.Parse(number, "US")
	require.NoError(t, err, "`%s` should have parsed successfully", number)
	require.NotNil(t, num, "resulting phone number cannot be nil")
}

func TestParsePhoneNumberLibrary(t *testing.T) {
	numbers := []string{
		"+1 612-123-5423",
		"612-123-5423",
		"6121235423",
		"1-612-123-5423",
	}
	for _, number := range numbers {
		mustParsePhoneNumber(t, number)
	}
}
