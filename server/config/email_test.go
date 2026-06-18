package config_test

import (
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

func TestEmail_BlockedEmailDomain(t *testing.T) {
	t.Run("empty list never blocks", func(t *testing.T) {
		email := config.Email{}
		domain, blocked := email.BlockedEmailDomain("person@example.com")
		assert.False(t, blocked, "an empty blocklist should never block anything")
		assert.Empty(t, domain, "no domain should be returned when nothing is blocked")
	})

	t.Run("blocks an exact domain match", func(t *testing.T) {
		email := config.Email{
			BlockedDomains: myownsanity.NewSet("example.com"),
		}
		domain, blocked := email.BlockedEmailDomain("person@example.com")
		assert.True(t, blocked, "example.com is on the blocklist so it should be blocked")
		assert.Equal(t, "example.com", domain, "should return the domain it matched on")
	})

	t.Run("normalizes the incoming email", func(t *testing.T) {
		// We only normalize the address the person is signing up with, the config
		// entries are expected to already be lowercase. So a person signing up as
		// EXAMPLE.com should still match the example.com entry.
		email := config.Email{
			BlockedDomains: myownsanity.NewSet("example.com"),
		}
		domain, blocked := email.BlockedEmailDomain("Person@EXAMPLE.com")
		assert.True(t, blocked, "the incoming email casing should not matter")
		assert.Equal(t, "example.com", domain, "the returned domain should be normalized to lowercase")
	})

	t.Run("does not block a subdomain", func(t *testing.T) {
		// We intentionally only match the domain exactly, blocking example.com
		// should NOT also block mail.example.com.
		email := config.Email{
			BlockedDomains: myownsanity.NewSet("example.com"),
		}
		domain, blocked := email.BlockedEmailDomain("person@mail.example.com")
		assert.False(t, blocked, "a subdomain of a blocked domain should not be blocked")
		assert.Empty(t, domain)
	})

	t.Run("an address with no domain is not blocked", func(t *testing.T) {
		email := config.Email{
			BlockedDomains: myownsanity.NewSet("example.com"),
		}
		domain, blocked := email.BlockedEmailDomain("not-an-email")
		assert.False(t, blocked, "if there is no domain to compare we should not block it")
		assert.Empty(t, domain)
	})

	t.Run("an unrelated domain is not blocked", func(t *testing.T) {
		email := config.Email{
			BlockedDomains: myownsanity.NewSet("example.com"),
		}
		_, blocked := email.BlockedEmailDomain("person@monetr.app")
		assert.False(t, blocked, "a domain that is not on the list should be allowed through")
	})

	t.Run("a config entry that is not normalized will not match", func(t *testing.T) {
		// It is on the operator to configure the blocklist with lowercase domains
		// and no leading @. We do NOT normalize the config side, so a sloppy entry
		// like this just silently wont match. This test is here to document that
		// expectation rather than to bless it.
		email := config.Email{
			BlockedDomains: myownsanity.NewSet("@Example.COM"),
		}
		_, blocked := email.BlockedEmailDomain("person@example.com")
		assert.False(t, blocked, "a config entry that was not normalized should not match")
	})

	t.Run("blocks when one of several domains matches", func(t *testing.T) {
		email := config.Email{
			BlockedDomains: myownsanity.NewSet(
				"mailinator.com",
				"guerrillamail.com",
				"example.com",
			),
		}
		_, blocked := email.BlockedEmailDomain("person@guerrillamail.com")
		assert.True(t, blocked, "a match against any entry in the list should block")
	})
}
