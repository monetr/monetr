package config

import (
	"strings"
	"time"

	"github.com/monetr/monetr/server/internal/myownsanity"
)

type Email struct {
	// Enabled controls whether the API can send emails at all. In order to
	// support things like forgot password links or email verification this must
	// be enabled.
	Enabled        bool              `yaml:"enabled"`
	Verification   EmailVerification `yaml:"verification"`
	ForgotPassword ForgotPassword    `yaml:"forgotPassword"`
	// Domain specifies the actual domain name used to send emails. Emails will
	// always be sent from `no-reply@{domain}`.
	Domain string `yaml:"domain"`
	// Email is sent via SMTP. If you want to send emails it is required to
	// include an SMTP configuration.
	SMTP SMTPClient `yaml:"smtp"`
	// BlockedDomains is the set of email domains that are NOT allowed to sign up
	// for a new account on this server. This only restricts sign up. Existing
	// users on one of these domains can still sign in, reset their password and
	// verify their email. An empty list turns the feature off. This applies even
	// when email sending (`enabled`) is turned off, it has nothing to do with
	// SMTP being configured.
	BlockedDomains myownsanity.Set[string] `yaml:"blockedDomains"`
}

type EmailVerification struct {
	// If you want to verify email addresses when a new user signs up then this
	// should be enabled. This will require a user to verify that they own (or at
	// least have proper access to) the email address that they used when they
	// signed up.
	Enabled bool `yaml:"enabled"`
	// Specify the amount of time that an email verification link is valid.
	TokenLifetime time.Duration `yaml:"tokenLifetime"`
}

type ForgotPassword struct {
	// If you want to allow people to reset their passwords then we need to be
	// able to send them a password reset link.
	Enabled bool `yaml:"enabled"`
	// Specify the amount of time that a password reset link will be valid.
	TokenLifetime time.Duration `yaml:"tokenLifetime"`
}

type SMTPClient struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
}

func (s Email) ShouldVerifyEmails() bool {
	return s.Enabled && s.Verification.Enabled
}

func (s Email) AllowPasswordReset() bool {
	return s.Enabled && s.ForgotPassword.Enabled
}

// BlockedEmailDomain returns the normalized domain portion of the email address
// along with true when that domain is on the sign up blocklist. When it is not
// blocked (no blocklist, no domain, or just not a match) it returns "", false.
// We hand the domain back so the caller can log which domain got blocked
// without having to parse the address a second time. We only call this during
// sign up so an operator can keep throwaway/disposable email providers from
// creating accounts. The comparison is case insensitive and matches the domain
// exactly, blocking `example.com` does NOT block `mail.example.com`.
// TODO Should we match subdomains too? For a disposable email blocklist an
// exact match is usually what you want, revisit this if someone needs the
// allow-list style behavior instead.
func (s Email) BlockedEmailDomain(emailAddress string) (string, bool) {
	if len(s.BlockedDomains) == 0 {
		return "", false
	}

	// The register controller only trims the email, it does not lowercase it
	// until the login is actually created, so normalize it ourselves here.
	emailAddress = strings.ToLower(strings.TrimSpace(emailAddress))
	at := strings.LastIndex(emailAddress, "@")
	if at < 0 {
		return "", false
	}
	domain := emailAddress[at+1:]
	if domain == "" {
		return "", false
	}

	if s.BlockedDomains.Has(domain) {
		return domain, true
	}

	return "", false
}
