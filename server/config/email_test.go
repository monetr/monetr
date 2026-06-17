package config_test

import (
	"testing"

	"github.com/monetr/monetr/server/config"
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
			BlockedDomains: []string{"example.com"},
		}
		domain, blocked := email.BlockedEmailDomain("person@example.com")
		assert.True(t, blocked, "example.com is on the blocklist so it should be blocked")
		assert.Equal(t, "example.com", domain, "should return the domain it matched on")
	})

	t.Run("is case insensitive", func(t *testing.T) {
		email := config.Email{
			BlockedDomains: []string{"Example.COM"},
		}
		domain, blocked := email.BlockedEmailDomain("Person@EXAMPLE.com")
		assert.True(t, blocked, "the match should not care about casing on either side")
		assert.Equal(t, "example.com", domain, "the returned domain should be normalized to lowercase")
	})

	t.Run("does not block a subdomain", func(t *testing.T) {
		// We intentionally only match the domain exactly, blocking example.com
		// should NOT also block mail.example.com.
		email := config.Email{
			BlockedDomains: []string{"example.com"},
		}
		domain, blocked := email.BlockedEmailDomain("person@mail.example.com")
		assert.False(t, blocked, "a subdomain of a blocked domain should not be blocked")
		assert.Empty(t, domain)
	})

	t.Run("tolerates a leading @ in the config entry", func(t *testing.T) {
		// Someone might write the blocklist entry as @example.com, we should treat
		// that the same as example.com.
		email := config.Email{
			BlockedDomains: []string{"@example.com"},
		}
		_, blocked := email.BlockedEmailDomain("person@example.com")
		assert.True(t, blocked, "a leading @ in the config entry should be ignored")
	})

	t.Run("an address with no domain is not blocked", func(t *testing.T) {
		email := config.Email{
			BlockedDomains: []string{"example.com"},
		}
		domain, blocked := email.BlockedEmailDomain("not-an-email")
		assert.False(t, blocked, "if there is no domain to compare we should not block it")
		assert.Empty(t, domain)
	})

	t.Run("an unrelated domain is not blocked", func(t *testing.T) {
		email := config.Email{
			BlockedDomains: []string{"example.com"},
		}
		_, blocked := email.BlockedEmailDomain("person@monetr.app")
		assert.False(t, blocked, "a domain that is not on the list should be allowed through")
	})

	t.Run("blocks when one of several domains matches", func(t *testing.T) {
		email := config.Email{
			BlockedDomains: []string{
				"mailinator.com",
				"guerrillamail.com",
				"example.com",
			},
		}
		_, blocked := email.BlockedEmailDomain("person@guerrillamail.com")
		assert.True(t, blocked, "a match against any entry in the list should block")
	})
}
