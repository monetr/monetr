package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/monetr/monetr/server/util"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/spf13/viper"
)

var (
	// FilePath is set via a flag in pkg/cmd.
	FilePath []string
	// LogLevel is set via a flag in pkg/cmd.
	LogLevel string
	// Migrate database!
	Migrate bool
)

type PlaidEnvironment string

const (
	PlaidSandbox     PlaidEnvironment = "sandbox"
	PlaidDevelopment PlaidEnvironment = "development"
	PlaidProduction  PlaidEnvironment = "production"
)

type Configuration struct {
	// configFile is not an actual configuration variable, but is used to let
	// usages know what file was loaded for the configuration.
	configFile string `yaml:"-"`

	Environment   string        `yaml:"environment"`
	AllowSignUp   bool          `yaml:"allowSignUp"`
	Beta          Beta          `yaml:"beta"`
	CORS          CORS          `yaml:"cors"`
	Email         Email         `yaml:"email"`
	KeyManagement KeyManagement `yaml:"keyManagement"`
	Links         Links         `yaml:"links"`
	Logging       Logging       `yaml:"logging"`
	Plaid         Plaid         `yaml:"plaid"`
	PostgreSQL    PostgreSQL    `yaml:"postgreSql"`
	ReCAPTCHA     ReCAPTCHA     `yaml:"reCAPTCHA"`
	Redis         Redis         `yaml:"redis"`
	Security      Security      `yaml:"security"`
	Sentry        Sentry        `yaml:"sentry"`
	Server        Server        `yaml:"server"`
	Storage       Storage       `yaml:"storage"`
	Stripe        Stripe        `yaml:"stripe"`
}

func (c Configuration) GetConfigFileName() string {
	return c.configFile
}

type Storage struct {
	// Enabled controls whether or not monetr can actually store files. If this is
	// disabled then some monetr features will not be available. These features
	// include things like file imports for transactions.
	Enabled bool `yaml:"enabled"`
	// Provider specifies which storage backend monetr should use. Allowed values
	// are:
	// - `s3`
	// - `filesystem`
	// Note: If you use the filesystem backend you cannot run multiple monetr
	// servers. Even if the filesystem is shared between the instances via
	// something like NFS; it can cause unpredictable behavior. It is recommended
	// to use the S3 backed storage when you are running multiple instances of
	// monetr. If you are self-hosting monetr though as a single instance, then
	// filesystem is the recommended storage backend for ease of use.
	Provider   string             `yaml:"provider"`
	S3         *S3Storage         `yaml:"s3"`
	Filesystem *FilesystemStorage `yaml:"filesystem"`
}

type S3Storage struct {
	AccessKey         *string `yaml:"accessKey"`
	Endpoint          *string `yaml:"endpoint"`
	Region            string  `yaml:"region"`
	SecretKey         *string `yaml:"secretKey"`
	UseEnvCredentials bool    `yaml:"useEnvCredentials"`

	Bucket         string `yaml:"bucket"`
	ForcePathStyle bool   `yaml:"forcePathStyle"`
}

type FilesystemStorage struct {
	BasePath string `yaml:"basePath"`
}

type Beta struct {
	EnableBetaCodes bool `yaml:"enableBetaCodes"`
}

type Security struct {
	// PrivateKey is the path to the file containing the ED22519 private key in
	// pem format.
	PrivateKey string `yaml:"privateKey"`
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
	Migrate            bool   `yaml:"migrate"`
}

func (c Configuration) GetEmail() Email {
	return c.Email
}

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

type ReCAPTCHA struct {
	Enabled        bool   `yaml:"enabled"`
	PublicKey      string `yaml:"publicKey"`
	PrivateKey     string `yaml:"privateKey"`
	Version        int    `yaml:"version"` // Currently only version 2 is supported by the UI.
	VerifyLogin    bool   `yaml:"verifyLogin"`
	VerifyRegister bool   `yaml:"loginRegister"`
	// VerifyForgotPassword determines whether or not the user will be required to
	// verify that they are not a robot overlord.
	VerifyForgotPassword bool `yaml:"verifyPasswordReset"`
}

func (r ReCAPTCHA) ShouldVerifyLogin() bool {
	return r.Enabled && r.VerifyLogin
}

func (r ReCAPTCHA) ShouldVerifyRegistration() bool {
	return r.Enabled && r.VerifyRegister
}

func (r ReCAPTCHA) ShouldVerifyForgotPassword() bool {
	return r.Enabled && r.VerifyForgotPassword
}

type Plaid struct {
	Enabled      bool              `yaml:"enabled"`
	ClientID     string            `yaml:"clientId"`
	ClientSecret string            `yaml:"clientSecret"`
	Environment  plaid.Environment `yaml:"environment"`
	// EnableReturningUserExperience changes the required data for sign up. If
	// this is enabled then the user must provide their full legal name as well
	// as their phone number.
	// If enabled; email address and phone number verification is REQUIRED.
	EnableReturningUserExperience bool   `yaml:"enableReturningUserExperience"`
	WebhooksEnabled               bool   `yaml:"webhooksEnabled"`
	WebhooksDomain                string `yaml:"webhooksDomain"`
	// OAuthDomain is used to specify the domain name that the user will be
	// brought to upon returning to monetr after authenticating to a bank that
	// requires OAuth. This will typically be a UI domain name and should not
	// include a protocol or a path. The protocol is auto inserted as `https` as
	// it is the only protocol supported. The path is currently hard coded until a
	// need for different paths arises?
	OAuthDomain string `yaml:"oauthDomain"`
	// Specify the country codes that monetr can connect to using Plaid. Some
	// countries require special access from Plaid and cannot simply be added to
	// enable the functionality.
	CountryCodes []plaid.CountryCode `yaml:"countryCodes"`
}

func (p Plaid) GetEnabled() bool {
	return p.Enabled && p.ClientID != "" && p.ClientSecret != ""
}

func (p Plaid) GetWebhooksURL() string {
	return fmt.Sprintf("https://%s/api/plaid/webhook", p.WebhooksDomain)
}

type CORS struct {
	AllowedOrigins []string `yaml:"allowedOrigins"`
	Debug          bool     `yaml:"debug"`
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
	Enabled     bool   `yaml:"enabled"`
	DSN         string `yaml:"dsn"`
	ExternalDSN string `yaml:"externalDSN"`
	// SecurityHeaderEndpoint tells monetr that CSP policy information can be
	// passed on to Sentry. If this value is provided, this will be included in
	// the CSP header.
	SecurityHeaderEndpoint string  `yaml:"securityHeaderEndpoint"`
	SampleRate             float64 `yaml:"sampleRate"`
	TraceSampleRate        float64 `yaml:"traceSampleRate"`
}

func (s Sentry) ExternalSentryEnabled() bool {
	return s.Enabled && s.ExternalDSN != ""
}

type Stripe struct {
	Enabled         bool   `yaml:"enabled"`
	APIKey          string `yaml:"apiKey"`
	PublicKey       string `yaml:"publicKey"`
	WebhooksEnabled bool   `yaml:"webhooksEnabled"`
	WebhookSecret   string `yaml:"webhookSecret"`
	InitialPlan     *Plan  `yaml:"initialPlan"`
	TaxesEnabled    bool   `yaml:"taxesEnabled"`
	FreeTrialDays   int    `yaml:"freeTrialDays"`
}

// IsBillingEnabled will return true if both Stripe and Billing are enabled. It
// will return false any other time.
func (s Stripe) IsBillingEnabled() bool {
	return s.Enabled
}

func getViper(configFilePath []string) *viper.Viper {
	v := viper.GetViper()
	v.SetConfigType("yaml")

	setupDefaults(v)
	setupEnv(v)

	switch len(FilePath) {
	case 0:
		{ // If we can determine the user's home directory, then look there.
			homeDir, err := os.UserHomeDir()
			if err == nil {
				v.AddConfigPath(path.Join(homeDir, "/.monetr/config.yaml"))
			}
		}

		v.AddConfigPath("/etc/monetr/config.yaml")
		v.AddConfigPath("config.yaml")
	default:
		for i := range configFilePath {
			path := configFilePath[i]
			v.SetConfigFile(path)
			if i == 0 {
				if err := v.ReadInConfig(); err != nil {
					log.Fatalf("failed to read config [%s]: %+v", path, err)
				}
			} else {
				if err := v.MergeInConfig(); err != nil {
					log.Fatalf("failed to read config [%s]: %+v", path, err)
				}
			}
		}
	}

	return v
}

func LoadConfiguration() Configuration {
	return LoadConfigurationFromFile(FilePath)
}

func LoadConfigurationFromFile(configFilePath []string) Configuration {
	v := getViper(configFilePath)

	return LoadConfigurationEx(v)
}

func LoadConfigurationEx(v *viper.Viper) (config Configuration) {
	if err := v.Unmarshal(&config); err != nil {
		panic(err)
	}

	config.configFile = v.ConfigFileUsed()

	privateKey, err := util.ExpandHomePath(config.Security.PrivateKey)
	if err != nil {
		panic(err)
	}
	config.Security.PrivateKey = privateKey

	return config
}

func setupDefaults(v *viper.Viper) {
	v.SetDefault("Environment", "development")
	v.SetDefault("AllowSignUp", true)
	v.SetDefault("Email.ForgotPassword.TokenLifetime", 10*time.Minute)
	v.SetDefault("Email.Verification.TokenLifetime", 10*time.Minute)
	v.SetDefault("Logging.Format", "text")
	v.SetDefault("Logging.Level", LogLevel) // Info
	v.SetDefault("Logging.StackDriver.Enabled", false)
	v.SetDefault("KeyManagement.Provider", "plaintext")
	v.SetDefault("Plaid.Enabled", true)
	v.SetDefault("Plaid.CountryCodes", []plaid.CountryCode{plaid.COUNTRYCODE_US})
	v.SetDefault("PostgreSQL.Address", "localhost")
	v.SetDefault("PostgreSQL.Database", "postgres")
	v.SetDefault("PostgreSQL.Port", 5432)
	v.SetDefault("PostgreSQL.Username", "postgres")
	v.SetDefault("Redis.Port", 6379)
	v.SetDefault("Redis.Database", 0)
	v.SetDefault("ReCAPTCHA.Enabled", false)
	v.SetDefault("ReCAPTCHA.VerifyLogin", true)
	v.SetDefault("ReCAPTCHA.VerifyRegister", true)
	v.SetDefault("ReCAPTCHA.VerifyForgotPassword", true)
	v.SetDefault("Security.PrivateKey", "/etc/monetr/ed25519.key")
	v.SetDefault("Sentry.SampleRate", 1.0)
	v.SetDefault("Sentry.TraceSampleRate", 1.0)
	v.SetDefault("Server.Cookies.Name", "M-Token")
	v.SetDefault("Server.Cookies.Secure", true)
	v.SetDefault("Server.Cookies.SameSiteStrict", true)
	v.SetDefault("Server.ListenPort", 4000)
	v.SetDefault("Server.ListenAddress", "0.0.0.0")
	v.SetDefault("Server.StatsPort", 9000)
	v.SetDefault("Server.UICacheHours", 14*24)
	v.SetDefault("Storage.Enabled", false)
	v.SetDefault("Storage.Provider", "filesystem")
	v.SetDefault("Storage.Filesystem.BasePath", "/etc/monetr/storage")
	v.SetDefault("Stripe.FreeTrialDays", 30)
}

func setupEnv(v *viper.Viper) {
	v.MustBindEnv("Environment", "MONETR_ENVIRONMENT")
	v.MustBindEnv("AllowSignUp", "MONETR_ALLOW_SIGN_UP")
	v.MustBindEnv("Beta.EnableBetaCodes", "MONETR_ENABLE_BETA_CODES")
	v.MustBindEnv("Cors.AllowedOrigins", "MONETR_CORS_ALLOWED_ORIGINS")
	v.MustBindEnv("Cors.Debug", "MONETR_CORS_DEBUG")
	v.MustBindEnv("Email.Enabled", "MONETR_EMAIL_ENABLED")
	v.MustBindEnv("Email.Domain", "MONETR_EMAIL_DOMAIN")
	v.MustBindEnv("Email.Verification.Enabled", "MONETR_EMAIL_VERIFICATION_ENABLED")
	v.MustBindEnv("Email.Verification.TokenLifetime", "MONETR_EMAIL_VERIFICATION_TOKEN_LIFETIME")
	v.MustBindEnv("Email.ForgotPassword.Enabled", "MONETR_EMAIL_FORGOT_PASSWORD_ENABLED")
	v.MustBindEnv("Email.ForgotPassword.TokenLifetime", "MONETR_EMAIL_FORGOT_PASSWORD_TOKEN_LIFETIME")
	v.MustBindEnv("Email.SMTP.Username", "MONETR_EMAIL_SMTP_USERNAME")
	v.MustBindEnv("Email.SMTP.Password", "MONETR_EMAIL_SMTP_PASSWORD")
	v.MustBindEnv("Email.SMTP.Host", "MONETR_EMAIL_SMTP_HOST")
	v.MustBindEnv("Email.SMTP.Port", "MONETR_EMAIL_SMTP_PORT")
	v.MustBindEnv("Logging.Level", "MONETR_LOG_LEVEL")
	v.MustBindEnv("Logging.Format", "MONETR_LOG_FORMAT")
	v.MustBindEnv("Logging.StackDriver.Enabled", "MONETR_LOG_STACKDRIVER_ENABLED")
	v.MustBindEnv("KeyManagement.Provider", "MONETR_KMS_PROVIDER")
	v.MustBindEnv("KeyManagement.AWS.AccessKey", "AWS_ACCESS_KEY_ID")
	v.MustBindEnv("KeyManagement.AWS.SecretKey", "AWS_ACCESS_KEY")
	v.MustBindEnv("KeyManagement.OpenBao.AuthMethod", "MONETR_OPENBAO_AUTH_METHOD")
	v.MustBindEnv("KeyManagement.OpenBao.Token", "MONETR_OPENBAO_TOKEN")
	v.MustBindEnv("KeyManagement.OpenBao.Endpoint", "MONETR_OPENBAO_ENDPOINT")
	v.MustBindEnv("KeyManagement.Vault.AuthMethod", "MONETR_VAULT_AUTH_METHOD")
	v.MustBindEnv("KeyManagement.Vault.Token", "MONETR_VAULT_TOKEN")
	v.MustBindEnv("KeyManagement.Vault.Endpoint", "MONETR_VAULT_ENDPOINT")
	v.MustBindEnv("Plaid.ClientID", "MONETR_PLAID_CLIENT_ID")
	v.MustBindEnv("Plaid.ClientSecret", "MONETR_PLAID_CLIENT_SECRET")
	v.MustBindEnv("Plaid.Environment", "MONETR_PLAID_ENVIRONMENT")
	v.MustBindEnv("Plaid.EnableBirthdatePrompt", "MONETR_PLAID_BIRTHDATE_PROMPT")
	v.MustBindEnv("Plaid.EnableReturningUserExperience", "MONETR_PLAID_RETURNING_EXPERIENCE")
	v.MustBindEnv("Plaid.WebhooksEnabled", "MONETR_PLAID_WEBHOOKS_ENABLED")
	v.MustBindEnv("Plaid.WebhooksDomain", "MONETR_PLAID_WEBHOOKS_DOMAIN")
	v.MustBindEnv("Plaid.OAuthDomain", "MONETR_PLAID_OAUTH_DOMAIN")
	v.MustBindEnv("PostgreSQL.Address", "MONETR_PG_ADDRESS")
	v.MustBindEnv("PostgreSQL.Port", "MONETR_PG_PORT")
	v.MustBindEnv("PostgreSQL.Username", "MONETR_PG_USERNAME")
	v.MustBindEnv("PostgreSQL.Password", "MONETR_PG_PASSWORD")
	v.MustBindEnv("PostgreSQL.Database", "MONETR_PG_DATABASE")
	v.MustBindEnv("PostgreSQL.InsecureSkipVerify", "MONETR_PG_INSECURE_SKIP_VERIFY")
	v.MustBindEnv("PostgreSQL.CACertificatePath", "MONETR_PG_CA_PATH")
	v.MustBindEnv("PostgreSQL.CertificatePath", "MONETR_PG_CERT_PATH")
	v.MustBindEnv("PostgreSQL.KeyPath", "MONETR_PG_KEY_PATH")
	v.MustBindEnv("ReCAPTCHA.Enabled", "MONETR_CAPTCHA_ENABLED")
	v.MustBindEnv("ReCAPTCHA.PublicKey", "MONETR_CAPTCHA_PUBLIC_KEY")
	v.MustBindEnv("ReCAPTCHA.PrivateKey", "MONETR_CAPTCHA_PRIVATE_KEY")
	v.MustBindEnv("ReCAPTCHA.VerifyLogin", "MONETR_CAPTCHA_VERIFY_LOGIN")
	v.MustBindEnv("ReCAPTCHA.VerifyRegister", "MONETR_CAPTCHA_VERIFY_REGISTER")
	v.MustBindEnv("Redis.Enabled", "MONETR_REDIS_ENABLED")
	v.MustBindEnv("Redis.Address", "MONETR_REDIS_ADDRESS")
	v.MustBindEnv("Redis.Port", "MONETR_REDIS_PORT")
	v.MustBindEnv("Redis.Database", "MONETR_REDIS_DATABASE")
	v.MustBindEnv("Redis.Username", "MONETR_REDIS_USERNAME")
	v.MustBindEnv("Redis.Password", "MONETR_REDIS_PASSWORD")
	v.MustBindEnv("Sentry.Enabled", "MONETR_SENTRY_ENABLED")
	v.MustBindEnv("Sentry.DSN", "MONETR_SENTRY_DSN")
	v.MustBindEnv("Sentry.ExternalDSN", "MONETR_SENTRY_EXTERNAL_DSN")
	v.MustBindEnv("Sentry.SampleRate", "MONETR_SENTRY_SAMPLE_RATE")
	v.MustBindEnv("Sentry.TraceSampleRate", "MONETR_SENTRY_TRACE_SAMPLE_RATE")
	v.MustBindEnv("Sentry.SecurityHeaderEndpoint", "MONETR_SENTRY_CSP_ENDPOINT")
	v.MustBindEnv("Server.ExternalURL", "MONETR_SERVER_EXTERNAL_URL")
	v.MustBindEnv("Storage.Enabled", "MONETR_STORAGE_ENABLED")
	v.MustBindEnv("Storage.Provider", "MONETR_STORAGE_PROVIDER")
	v.MustBindEnv("Stripe.Enabled", "MONETR_STRIPE_ENABLED")
	v.MustBindEnv("Stripe.APIKey", "MONETR_STRIPE_API_KEY")
	v.MustBindEnv("Stripe.PublicKey", "MONETR_STRIPE_PUBLIC_KEY")
	v.MustBindEnv("Stripe.WebhooksEnabled", "MONETR_STRIPE_WEBHOOKS_ENABLED")
	v.MustBindEnv("Stripe.WebhooksDomain", "MONETR_STRIPE_WEBHOOKS_DOMAIN")
	v.MustBindEnv("Stripe.WebhookSecret", "MONETR_STRIPE_WEBHOOK_SECRET")
	v.MustBindEnv("Stripe.BillingEnabled", "MONETR_STRIPE_BILLING_ENABLED")
	v.MustBindEnv("Stripe.TaxesEnabled", "MONETR_STRIPE_TAXES_ENABLED")
	v.MustBindEnv("Stripe.InitialPlan.StripePriceId", "MONETR_STRIPE_DEFAULT_PRICE_ID")
	v.AutomaticEnv()
}
