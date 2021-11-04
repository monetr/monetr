package config

import (
	"fmt"
	"time"

	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/viper"
)

const EnvironmentPrefix = "MONETR"

type PlaidEnvironment string

const (
	PlaidSandbox     PlaidEnvironment = "sandbox"
	PlaidDevelopment PlaidEnvironment = "development"
	PlaidProduction  PlaidEnvironment = "production"
)

type SendGridTemplate string

const (
	VerifyEmailAddressTemplate SendGridTemplate = "verifyEmailTemplate"
	ForgotPasswordTemplate     SendGridTemplate = "forgotPasswordTemplate"
)

type Configuration struct {
	Name          string
	ListenPort    int
	StatsPort     int
	Environment   string
	UIDomainName  string
	APIDomainName string
	AllowSignUp   bool
	Beta          Beta
	CORS          CORS
	JWT           JWT
	Logging       Logging
	Plaid         Plaid
	PostgreSQL    PostgreSQL
	ReCAPTCHA     ReCAPTCHA
	Redis         Redis
	Email         Email
	Sentry        Sentry
	Stripe        Stripe
	Vault         Vault
}

func (c Configuration) GetUIDomainName() string {
	return c.UIDomainName
}

type Beta struct {
	EnableBetaCodes bool
}

type JWT struct {
	LoginJwtSecret        string
	RegistrationJwtSecret string
}

type PostgreSQL struct {
	Address            string
	Port               int
	Username           string
	Password           string
	Database           string
	InsecureSkipVerify bool
	CACertificatePath  string
	KeyPath            string
	CertificatePath    string
}

func (c Configuration) GetEmail() Email {
	return c.Email
}

type Email struct {
	// Enabled controls whether the API can send emails at all. In order to support things like forgot password links or
	// email verification this must be enabled.
	Enabled      bool
	Verification EmailVerification
	// Domain specifies the actual domain name used to send emails. Emails will always be sent from `no-reply@{domain}`.
	Domain string
	// Email is sent via SMTP. If you want to send emails it is required to include an SMTP configuration.
	SMTP SMTPClient
}

type EmailVerification struct {
	// If you want to verify email addresses when a new user signs up then this should be enabled. This will require a
	// user to verify that they own (or at least have proper access to) the email address that they used when they
	// signed up.
	Enabled bool
	// Specify the amount of time that an email verification link is valid.
	TokenLifetime time.Duration
	// The secret used to generate verification tokens and validate them.
	TokenSecret string
}

type SMTPClient struct {
	Identity string
	Username string
	Password string
	Host     string
	Port     int
}

func (s Email) ShouldVerifyEmails() bool {
	return s.Enabled && s.Verification.Enabled
}

type ReCAPTCHA struct {
	Enabled    bool
	PublicKey  string
	PrivateKey string
	Version    int // Currently only version 2 is supported by the UI.

	VerifyLogin    bool
	VerifyRegister bool
}

func (r ReCAPTCHA) ShouldVerifyLogin() bool {
	return r.Enabled && r.VerifyLogin
}

func (r ReCAPTCHA) ShouldVerifyRegistration() bool {
	return r.Enabled && r.VerifyRegister
}

type Plaid struct {
	ClientID     string
	ClientSecret string
	Environment  plaid.Environment
	// This does not seem to be a scope within the documentation. Per the
	// documentation "balance is not a valid product" and is enabled
	// automatically. It is not clear if that includes this beta feature though.
	EnableBalanceTransfers bool

	// EnableReturningUserExperience changes the required data for sign up. If
	// this is enabled then the user must provide their full legal name as well
	// as their phone number.
	// If enabled; email address and phone number verification is REQUIRED.
	EnableReturningUserExperience bool

	// EnableBirthdatePrompt will allow users to provide their birthday during
	// sign up or afterwards in their user settings. This is used by plaid for
	// future products. At the time of writing this it does not do anything.
	EnableBirthdatePrompt bool

	WebhooksEnabled bool
	WebhooksDomain  string
	// OAuthDomain is used to specify the domain name that the user will be brought to upon returning to monetr after
	// authenticating to a bank that requires OAuth. This will typically be a UI domain name and should not include a
	// protocol or a path. The protocol is auto inserted as `https` as it is the only protocol supported. The path is
	// currently hard coded until a need for different paths arises?
	OAuthDomain string
}

func (p Plaid) GetWebhooksURL() string {
	return fmt.Sprintf("https://%s/api/plaid/webhook", p.WebhooksDomain)
}

type CORS struct {
	AllowedOrigins []string
	Debug          bool
}

// Redis defines the config used to connect to a redis for our worker pool. If these are left blank or default then we
// will instead use a mock redis pool that is internal only. This is fine for single instance deployments, but anytime
// more than one instance of the API is running a redis instance will be required.
type Redis struct {
	Enabled   bool
	Address   string
	Port      int
	Namespace string
}

type Logging struct {
	Level       string
	Format      string
	StackDriver StackDriverLogging
}

type StackDriverLogging struct {
	Enabled   bool
	ProjectID string
	LogName   string
}

type Sentry struct {
	Enabled         bool
	DSN             string
	SampleRate      float64
	TraceSampleRate float64
}

type Stripe struct {
	Enabled         bool
	APIKey          string
	PublicKey       string
	WebhooksEnabled bool
	WebhooksDomain  string
	WebhookSecret   string
	InitialPlan     *Plan
	Plans           []Plan
	BillingEnabled  bool
}

// IsBillingEnabled will return true if both Stripe and Billing are enabled. It will return false any other time.
func (s Stripe) IsBillingEnabled() bool {
	return s.Enabled && s.BillingEnabled
}

type Vault struct {
	Enabled            bool
	Address            string
	Auth               string
	Token              string
	TokenFile          string
	Username, Password string
	Role               string
	CertificatePath    string
	KeyPath            string
	CACertificatePath  string
	InsecureSkipVerify bool
	Timeout            time.Duration
	IdleConnTimeout    time.Duration
}

func LoadConfiguration(configFilePath *string) Configuration {
	v := viper.GetViper()

	v.SetEnvPrefix(EnvironmentPrefix)
	v.AutomaticEnv()

	setupDefaults(v)
	setupEnv(v)

	if configFilePath != nil {
		viper.SetConfigName(*configFilePath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/monetr/")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("failed to read in config from file: %+v\n", err)
	}

	var config Configuration
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	return config
}

func setupDefaults(v *viper.Viper) {
	v.SetDefault("Name", "monetr")
	v.SetDefault("ListenPort", 4000)
	v.SetDefault("StatsPort", 9000)
	v.SetDefault("Environment", "development")
	v.SetDefault("UIDomainName", "localhost:3000")
	v.SetDefault("APIDomainName", "localhost:4000")
	v.SetDefault("AllowSignUp", true)
	v.SetDefault("Email.Verification.TokenLifetime", 10*time.Minute)
	v.SetDefault("Logging.Level", "info")
	v.SetDefault("Logging.Format", "text")
	v.SetDefault("PostgreSQL.Address", "localhost")
	v.SetDefault("PostgreSQL.Database", "postgres")
	v.SetDefault("PostgreSQL.Port", 5432)
	v.SetDefault("PostgreSQL.Username", "postgres")
	v.SetDefault("ReCAPTCHA.Enabled", false)
	v.SetDefault("Vault.Auth", "kubernetes")
	v.SetDefault("Vault.IdleConnTimeout", 9*time.Minute)
	v.SetDefault("Vault.Timeout", 30*time.Second)
	v.SetDefault("Vault.TokenFile", "/var/run/secrets/kubernetes.io/serviceaccount/token")
}

func setupEnv(v *viper.Viper) {
	_ = v.BindEnv("Name", "MONETR_NAME")
	_ = v.BindEnv("ListenPort", "MONETR_PORT")
	_ = v.BindEnv("StatsPort", "MONETR_STATS_PORT")
	_ = v.BindEnv("Environment", "MONETR_ENVIRONMENT")
	_ = v.BindEnv("UIDomainName", "MONETR_UI_DOMAIN_NAME")
	_ = v.BindEnv("APIDomainName", "MONETR_API_DOMAIN_NAME")
	_ = v.BindEnv("AllowSignUp", "MONETR_ALLOW_SIGN_UP")
	_ = v.BindEnv("EnableWebhooks", "MONETR_ENABLE_WEBHOOKS")
	_ = v.BindEnv("Beta.EnableBetaCodes", "MONETR_ENABLE_BETA_CODES")
	_ = v.BindEnv("Cors.AllowedOrigins", "MONETR_CORS_ALLOWED_ORIGINS")
	_ = v.BindEnv("Cors.Debug", "MONETR_CORS_DEBUG")
	_ = v.BindEnv("Email.Enabled", "MONETR_EMAIL_ENABLED")
	_ = v.BindEnv("Email.Domain", "MONETR_EMAIL_DOMAIN")
	_ = v.BindEnv("Email.Verification.Enabled", "MONETR_EMAIL_VERIFICATION_ENABLED")
	_ = v.BindEnv("Email.Verification.TokenLifetime", "MONETR_EMAIL_VERIFICATION_TOKEN_LIFETIME")
	_ = v.BindEnv("Email.Verification.TokenSecret", "MONETR_EMAIL_VERIFICATION_TOKEN_SECRET")
	_ = v.BindEnv("Email.SMTP.Identity", "MONETR_EMAIL_SMTP_IDENTITY")
	_ = v.BindEnv("Email.SMTP.Username", "MONETR_EMAIL_SMTP_USERNAME")
	_ = v.BindEnv("Email.SMTP.Password", "MONETR_EMAIL_SMTP_PASSWORD")
	_ = v.BindEnv("Email.SMTP.Host", "MONETR_EMAIL_SMTP_HOST")
	_ = v.BindEnv("Email.SMTP.Port", "MONETR_EMAIL_SMTP_PORT")
	_ = v.BindEnv("JWT.LoginJwtSecret", "MONETR_JWT_LOGIN_SECRET")
	_ = v.BindEnv("JWT.RegistrationJwtSecret", "MONETR_JWT_REGISTRATION_SECRET")
	_ = v.BindEnv("Logging.Level", "MONETR_LOG_LEVEL")
	_ = v.BindEnv("Logging.Format", "MONETR_LOG_FORMAT")
	_ = v.BindEnv("Plaid.ClientID", "MONETR_PLAID_CLIENT_ID")
	_ = v.BindEnv("Plaid.ClientSecret", "MONETR_PLAID_CLIENT_SECRET")
	_ = v.BindEnv("Plaid.Environment", "MONETR_PLAID_ENVIRONMENT")
	_ = v.BindEnv("Plaid.EnableBirthdatePrompt", "MONETR_PLAID_BIRTHDATE_PROMPT")
	_ = v.BindEnv("Plaid.EnableReturningUserExperience", "MONETR_PLAID_RETURNING_EXPERIENCE")
	_ = v.BindEnv("Plaid.WebhooksEnabled", "MONETR_PLAID_WEBHOOKS_ENABLED")
	_ = v.BindEnv("Plaid.WebhooksDomain", "MONETR_PLAID_WEBHOOKS_DOMAIN")
	_ = v.BindEnv("Plaid.OAuthDomain", "MONETR_PLAID_OAUTH_DOMAIN")
	_ = v.BindEnv("PostgreSQL.Address", "MONETR_PG_ADDRESS")
	_ = v.BindEnv("PostgreSQL.Port", "MONETR_PG_PORT")
	_ = v.BindEnv("PostgreSQL.Username", "MONETR_PG_USERNAME")
	_ = v.BindEnv("PostgreSQL.Password", "MONETR_PG_PASSWORD")
	_ = v.BindEnv("PostgreSQL.Database", "MONETR_PG_DATABASE")
	_ = v.BindEnv("PostgreSQL.InsecureSkipVerify", "MONETR_PG_INSECURE_SKIP_VERIFY")
	_ = v.BindEnv("PostgreSQL.CACertificatePath", "MONETR_PG_CA_PATH")
	_ = v.BindEnv("PostgreSQL.CertificatePath", "MONETR_PG_CERT_PATH")
	_ = v.BindEnv("PostgreSQL.KeyPath", "MONETR_PG_KEY_PATH")
	_ = v.BindEnv("ReCAPTCHA.Enabled", "MONETR_CAPTCHA_ENABLED")
	_ = v.BindEnv("ReCAPTCHA.PublicKey", "MONETR_CAPTCHA_PUBLIC_KEY")
	_ = v.BindEnv("ReCAPTCHA.PrivateKey", "MONETR_CAPTCHA_PRIVATE_KEY")
	_ = v.BindEnv("ReCAPTCHA.VerifyLogin", "MONETR_CAPTCHA_VERIFY_LOGIN")
	_ = v.BindEnv("ReCAPTCHA.VerifyRegister", "MONETR_CAPTCHA_VERIFY_REGISTER")
	_ = v.BindEnv("Redis.Enabled", "MONETR_REDIS_ENABLED")
	_ = v.BindEnv("Redis.Address", "MONETR_REDIS_ADDRESS")
	_ = v.BindEnv("Redis.Port", "MONETR_REDIS_PORT")
	_ = v.BindEnv("Redis.Namespace", "MONETR_REDIS_NAMESPACE")
	_ = v.BindEnv("Sentry.Enabled", "MONETR_SENTRY_ENABLED")
	_ = v.BindEnv("Sentry.DSN", "MONETR_SENTRY_DSN")
	_ = v.BindEnv("Sentry.SampleRate", "MONETR_SENTRY_SAMPLE_RATE")
	_ = v.BindEnv("Sentry.TraceSampleRate", "MONETR_SENTRY_TRACE_SAMPLE_RATE")
	_ = v.BindEnv("Stripe.Enabled", "MONETR_STRIPE_ENABLED")
	_ = v.BindEnv("Stripe.APIKey", "MONETR_STRIPE_API_KEY")
	_ = v.BindEnv("Stripe.PublicKey", "MONETR_STRIPE_PUBLIC_KEY")
	_ = v.BindEnv("Stripe.WebhooksEnabled", "MONETR_STRIPE_WEBHOOKS_ENABLED")
	_ = v.BindEnv("Stripe.WebhooksDomain", "MONETR_STRIPE_WEBHOOKS_DOMAIN")
	_ = v.BindEnv("Stripe.WebhookSecret", "MONETR_STRIPE_WEBHOOK_SECRET")
	_ = v.BindEnv("Stripe.BillingEnabled", "MONETR_STRIPE_BILLING_ENABLED")
	_ = v.BindEnv("Vault.Enabled", "MONETR_VAULT_ENABLED")
	_ = v.BindEnv("Vault.Address", "MONETR_VAULT_ADDRESS")
	_ = v.BindEnv("Vault.Auth", "MONETR_VAULT_AUTH")
	_ = v.BindEnv("Vault.Token", "MONETR_VAULT_TOKEN")
	_ = v.BindEnv("Vault.TokenFile", "MONETR_VAULT_TOKEN_FILE")
	_ = v.BindEnv("Vault.Role", "MONETR_VAULT_ROLE")
	_ = v.BindEnv("Vault.CertificatePath", "MONETR_VAULT_TLS_CERT_PATH")
	_ = v.BindEnv("Vault.KeyPath", "MONETR_VAULT_TLS_KEY_PATH")
	_ = v.BindEnv("Vault.CACertificatePath", "MONETR_VAULT_TLS_CA_PATH")
	_ = v.BindEnv("Vault.InsecureSkipVerify", "MONETR_VAULT_INSECURE_SKIP_VERIFY")
	_ = v.BindEnv("Vault.Timeout", "MONETR_VAULT_TIMEOUT")
	_ = v.BindEnv("Vault.IdleConnTimeout", "MONETR_VAULT_IDLE_CONN_TIMEOUT")
}
