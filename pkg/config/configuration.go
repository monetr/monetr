package config

import (
	"fmt"
	"os"
	"time"

	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/viper"
)

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
	// DEPRECATED: This is not used anymore. It serves no function at all.
	Name string `yaml:"name"`
	// DEPRECATED: This is not used anymore. Use Server.ListenPort instead.
	ListenPort int `yaml:"listenPort"`
	// DEPRECATED: This is not used anymore. Use Server.StatsPort instead.
	StatsPort     int        `yaml:"statsPort"`
	Environment   string     `yaml:"environment"`
	UIDomainName  string     `yaml:"uiDomainName"`
	APIDomainName string     `yaml:"apiDomainName"`
	AllowSignUp   bool       `yaml:"allowSignUp"`
	Server        Server     `yaml:"server"`
	Beta          Beta       `yaml:"beta"`
	CORS          CORS       `yaml:"cors"`
	JWT           JWT        `yaml:"jwt"`
	Logging       Logging    `yaml:"logging"`
	Plaid         Plaid      `yaml:"plaid"`
	PostgreSQL    PostgreSQL `yaml:"postgreSql"`
	ReCAPTCHA     ReCAPTCHA  `yaml:"reCAPTCHA"`
	Redis         Redis      `yaml:"redis"`
	Email         Email      `yaml:"email"`
	Sentry        Sentry     `yaml:"sentry"`
	Stripe        Stripe     `yaml:"stripe"`
	Vault         Vault      `yaml:"vault"`
}

func (c Configuration) GetUIDomainName() string {
	return c.UIDomainName
}

type Server struct {
	// ListenPort defines the port that monetr will listen for HTTP requests on. This port should be forwarded such that
	// it is accessible to the desired clients. Be that on a local network, or forwarded to the public internet.
	ListenPort int `yaml:"listenPort"`
	// StatsPort is the port that our prometheus metrics are served on. This port should not be publicly accessible and
	// should only be accessible by the prometheus server scraping for metrics. It is not an endpoint that needs to be
	// secured as no sensitive client information will be served by it; but it should not be accessible publicly.
	StatsPort int `yaml:"statsPort"`
	// Cookies defines the parameters used for issuing and processing cookies from clients. Cookies are used for
	// authentication.
	Cookies Cookies `yaml:"cookies"`
}

type Cookies struct {
	// SameSiteStrict allows the host of monetr to define whether the cookie used for authentication is limited to same
	// site. This might impact use cases where the UI is on a different domain than the API. In general, it is
	// recommended that this is enabled and that the UI and API are served from the same domain.
	SameSiteStrict bool `yaml:"sameSiteStrict"`
	// Secure specifies that the authentication cookie issued and required by API endpoints is a secure cookie. This
	// defaults to true, but requires that the host of monetr use HTTPS. If you are not using HTTPS then this must be
	// disabled for API calls to succeed.
	Secure bool `yaml:"secure"`
	// Name defines the name of the cookie to use for authentication. This defaults to `M-Token` but can be customized
	// if the host wants to.
	Name string `yaml:"name"`
}

type Beta struct {
	EnableBetaCodes bool `yaml:"enableBetaCodes"`
}

type JWT struct {
	LoginJwtSecret        string `yaml:"loginJwtSecret"`
	RegistrationJwtSecret string `yaml:"registrationJwtSecret"`
}

type PostgreSQL struct {
	Address            string `yaml:"address"`
	Port               int    `yaml:"port"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Database           string `yaml:"database"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	CACertificatePath  string `yaml:"caCertificatePath"`
	KeyPath            string `yaml:"keyPath"`
	CertificatePath    string `yaml:"certificatePath"`
}

func (c Configuration) GetEmail() Email {
	return c.Email
}

type Email struct {
	// Enabled controls whether the API can send emails at all. In order to support things like forgot password links or
	// email verification this must be enabled.
	Enabled        bool              `yaml:"enabled"`
	Verification   EmailVerification `yaml:"verification"`
	ForgotPassword ForgotPassword    `yaml:"forgotPassword"`
	// Domain specifies the actual domain name used to send emails. Emails will always be sent from `no-reply@{domain}`.
	Domain string `yaml:"domain"`
	// Email is sent via SMTP. If you want to send emails it is required to include an SMTP configuration.
	SMTP SMTPClient `yaml:"smtp"`
}

type EmailVerification struct {
	// If you want to verify email addresses when a new user signs up then this should be enabled. This will require a
	// user to verify that they own (or at least have proper access to) the email address that they used when they
	// signed up.
	Enabled bool `yaml:"enabled"`
	// Specify the amount of time that an email verification link is valid.
	TokenLifetime time.Duration `yaml:"tokenLifetime"`
	// The secret used to generate verification tokens and validate them.
	TokenSecret string `yaml:"tokenSecret"`
}

type ForgotPassword struct {
	// If you want to allow people to reset their passwords then we need to be able to send them a password reset link.
	Enabled bool `yaml:"enabled"`
	// Specify the amount of time that a password reset link will be valid.
	TokenLifetime time.Duration `yaml:"tokenLifetime"`
	// Specify a secret used to generate the password reset links as well as validate them.
	TokenSecret string `yaml:"tokenSecret"`
}

type SMTPClient struct {
	Identity string `yaml:"identity"`
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

type ReCAPTCHA struct {
	Enabled        bool   `yaml:"enabled"`
	PublicKey      string `yaml:"publicKey"`
	PrivateKey     string `yaml:"privateKey"`
	Version        int    `yaml:"version"` // Currently only version 2 is supported by the UI.
	VerifyLogin    bool   `yaml:"verifyLogin"`
	VerifyRegister bool   `yaml:"loginRegister"`
}

func (r ReCAPTCHA) ShouldVerifyLogin() bool {
	return r.Enabled && r.VerifyLogin
}

func (r ReCAPTCHA) ShouldVerifyRegistration() bool {
	return r.Enabled && r.VerifyRegister
}

type Plaid struct {
	ClientID     string            `yaml:"clientId"`
	ClientSecret string            `yaml:"clientSecret"`
	Environment  plaid.Environment `yaml:"environment"`
	// This does not seem to be a scope within the documentation. Per the
	// documentation "balance is not a valid product" and is enabled
	// automatically. It is not clear if that includes this beta feature though.
	EnableBalanceTransfers bool `yaml:"enableBalanceTransfers"`

	// EnableReturningUserExperience changes the required data for sign up. If
	// this is enabled then the user must provide their full legal name as well
	// as their phone number.
	// If enabled; email address and phone number verification is REQUIRED.
	EnableReturningUserExperience bool `yaml:"enableReturningUserExperience"`

	// EnableBirthdatePrompt will allow users to provide their birthday during
	// sign up or afterwards in their user settings. This is used by plaid for
	// future products. At the time of writing this it does not do anything.
	EnableBirthdatePrompt bool `yaml:"enableBirthdatePrompt"`

	WebhooksEnabled bool   `yaml:"webhooksEnabled"`
	WebhooksDomain  string `yaml:"webhooksDomain"`
	// OAuthDomain is used to specify the domain name that the user will be brought to upon returning to monetr after
	// authenticating to a bank that requires OAuth. This will typically be a UI domain name and should not include a
	// protocol or a path. The protocol is auto inserted as `https` as it is the only protocol supported. The path is
	// currently hard coded until a need for different paths arises?
	OAuthDomain string `yaml:"oauthDomain"`
	// MaxNumberOfLinks defines the max number of active Plaid links a single account can have. If this is set to 0 then
	// there is no limit.
	MaxNumberOfLinks int `yaml:"maxNumberOfLinks"`
}

func (p Plaid) GetWebhooksURL() string {
	return fmt.Sprintf("https://%s/api/plaid/webhook", p.WebhooksDomain)
}

type CORS struct {
	AllowedOrigins []string `yaml:"allowedOrigins"`
	Debug          bool     `yaml:"debug"`
}

// Redis defines the config used to connect to a redis for our worker pool. If these are left blank or default then we
// will instead use a mock redis pool that is internal only. This is fine for single instance deployments, but anytime
// more than one instance of the API is running a redis instance will be required.
type Redis struct {
	Enabled   bool   `yaml:"enabled"`
	Address   string `yaml:"address"`
	Port      int    `yaml:"port"`
	Namespace string `yaml:"namespace"`
}

type Logging struct {
	Level       string             `yaml:"level"`
	Format      string             `yaml:"format"`
	StackDriver StackDriverLogging `yaml:"stackDriver"`
}

type StackDriverLogging struct {
	Enabled bool `yaml:"enabled"`
}

type Sentry struct {
	Enabled         bool    `yaml:"enabled"`
	DSN             string  `yaml:"dsn"`
	ExternalDSN     string  `yaml:"externalDSN"`
	SampleRate      float64 `yaml:"sampleRate"`
	TraceSampleRate float64 `yaml:"traceSampleRate"`
}

func (s Sentry) ExternalSentryEnabled() bool {
	return s.Enabled && s.ExternalDSN != ""
}

type Stripe struct {
	Enabled         bool   `yaml:"enabled"`
	APIKey          string `yaml:"apiKey"`
	PublicKey       string `yaml:"publicKey"`
	WebhooksEnabled bool   `yaml:"webhooksEnabled"`
	// DEPRECATED: This does not matter to the application. This must be set inside Stripe.
	WebhooksDomain string `yaml:"webhooksDomain"`
	WebhookSecret  string `yaml:"webhookSecret"`
	InitialPlan    *Plan  `yaml:"initialPlan"`
	Plans          []Plan `yaml:"plans"`
	BillingEnabled bool   `yaml:"billingEnabled"`
	TaxesEnabled   bool   `yaml:"taxesEnabled"`
}

// IsBillingEnabled will return true if both Stripe and Billing are enabled. It will return false any other time.
func (s Stripe) IsBillingEnabled() bool {
	return s.Enabled && s.BillingEnabled
}

type Vault struct {
	Enabled            bool          `yaml:"enabled"`
	Address            string        `yaml:"address"`
	Auth               string        `yaml:"auth"`
	Token              string        `yaml:"token"`
	TokenFile          string        `yaml:"tokenFile"`
	Username           string        `yaml:"username"`
	Password           string        `yaml:"password"`
	Role               string        `yaml:"role"`
	CertificatePath    string        `yaml:"certificatePath"`
	KeyPath            string        `yaml:"keyPath"`
	CACertificatePath  string        `yaml:"caCertificatePath"`
	InsecureSkipVerify bool          `yaml:"insecureSkipVerify"`
	Timeout            time.Duration `yaml:"timeout"`
	IdleConnTimeout    time.Duration `yaml:"idleConnTimeout"`
}

func getViper(configFilePath *string) *viper.Viper {
	v := viper.GetViper()

	if configFilePath != nil {
		v.SetConfigName(*configFilePath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")

		{ // If we can determine the user's home directory, then look there + /.sentry for the config
			homeDir, err := os.UserHomeDir()
			if err == nil {
				v.AddConfigPath(homeDir + "/.monetr")
			}
		}

		v.AddConfigPath("/etc/monetr/")
		v.AddConfigPath(".")
	}

	setupDefaults(v)
	setupEnv(v)

	return v
}

func LoadConfiguration(configFilePath *string) Configuration {
	v := getViper(configFilePath)

	return LoadConfigurationEx(v)
}

func LoadConfigurationEx(v *viper.Viper) Configuration {
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("failed to read in config from file: %+v\n", err)
	}

	var config Configuration
	if err := v.Unmarshal(&config); err != nil {
		panic(err)
	}

	return config
}

func GenerateConfigFile(configFilePath *string, outputFilePath string) error {
	var v *viper.Viper
	if configFilePath == nil {
		v = viper.GetViper()
		setupDefaults(v)
	} else {
		v = getViper(configFilePath)
	}

	return v.SafeWriteConfigAs(outputFilePath)
}

func setupDefaults(v *viper.Viper) {
	v.SetDefault("APIDomainName", "localhost:4000")
	v.SetDefault("AllowSignUp", true)
	v.SetDefault("Email.ForgotPassword.TokenLifetime", 10*time.Minute)
	v.SetDefault("Email.Verification.TokenLifetime", 10*time.Minute)
	v.SetDefault("Environment", "development")
	v.SetDefault("Logging.Format", "text")
	v.SetDefault("Logging.Level", "info")
	v.SetDefault("Logging.StackDriver.Enabled", false)
	v.SetDefault("PostgreSQL.Address", "localhost")
	v.SetDefault("PostgreSQL.Database", "postgres")
	v.SetDefault("PostgreSQL.Port", 5432)
	v.SetDefault("PostgreSQL.Username", "postgres")
	v.SetDefault("ReCAPTCHA.Enabled", false)
	v.SetDefault("Server.Cookie.Name", "M-Token")
	v.SetDefault("Server.ListenPort", 4000)
	v.SetDefault("Server.StatsPort", 9000)
	v.SetDefault("UIDomainName", "localhost:4000")
	v.SetDefault("Vault.Auth", "kubernetes")
	v.SetDefault("Vault.IdleConnTimeout", 9*time.Minute)
	v.SetDefault("Vault.Timeout", 10*time.Second)
	v.SetDefault("Vault.TokenFile", "/var/run/secrets/kubernetes.io/serviceaccount/token")
}

func setupEnv(v *viper.Viper) {
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
	_ = v.BindEnv("Email.ForgotPassword.Enabled", "MONETR_EMAIL_FORGOT_PASSWORD_ENABLED")
	_ = v.BindEnv("Email.ForgotPassword.TokenLifetime", "MONETR_EMAIL_FORGOT_PASSWORD_TOKEN_LIFETIME")
	_ = v.BindEnv("Email.ForgotPassword.TokenSecret", "MONETR_EMAIL_FORGOT_PASSWORD_TOKEN_SECRET")
	_ = v.BindEnv("Email.SMTP.Identity", "MONETR_EMAIL_SMTP_IDENTITY")
	_ = v.BindEnv("Email.SMTP.Username", "MONETR_EMAIL_SMTP_USERNAME")
	_ = v.BindEnv("Email.SMTP.Password", "MONETR_EMAIL_SMTP_PASSWORD")
	_ = v.BindEnv("Email.SMTP.Host", "MONETR_EMAIL_SMTP_HOST")
	_ = v.BindEnv("Email.SMTP.Port", "MONETR_EMAIL_SMTP_PORT")
	_ = v.BindEnv("JWT.LoginJwtSecret", "MONETR_JWT_LOGIN_SECRET")
	_ = v.BindEnv("JWT.RegistrationJwtSecret", "MONETR_JWT_REGISTRATION_SECRET")
	_ = v.BindEnv("Logging.Level", "MONETR_LOG_LEVEL")
	_ = v.BindEnv("Logging.Format", "MONETR_LOG_FORMAT")
	_ = v.BindEnv("Logging.StackDriver.Enabled", "MONETR_LOG_STACKDRIVER_ENABLED")
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
	_ = v.BindEnv("Sentry.ExternalDSN", "MONETR_SENTRY_EXTERNAL_DSN")
	_ = v.BindEnv("Sentry.SampleRate", "MONETR_SENTRY_SAMPLE_RATE")
	_ = v.BindEnv("Sentry.TraceSampleRate", "MONETR_SENTRY_TRACE_SAMPLE_RATE")
	_ = v.BindEnv("Stripe.Enabled", "MONETR_STRIPE_ENABLED")
	_ = v.BindEnv("Stripe.APIKey", "MONETR_STRIPE_API_KEY")
	_ = v.BindEnv("Stripe.PublicKey", "MONETR_STRIPE_PUBLIC_KEY")
	_ = v.BindEnv("Stripe.WebhooksEnabled", "MONETR_STRIPE_WEBHOOKS_ENABLED")
	_ = v.BindEnv("Stripe.WebhooksDomain", "MONETR_STRIPE_WEBHOOKS_DOMAIN")
	_ = v.BindEnv("Stripe.WebhookSecret", "MONETR_STRIPE_WEBHOOK_SECRET")
	_ = v.BindEnv("Stripe.BillingEnabled", "MONETR_STRIPE_BILLING_ENABLED")
	_ = v.BindEnv("Stripe.TaxesEnabled", "MONETR_STRIPE_TAXES_ENABLED")
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
