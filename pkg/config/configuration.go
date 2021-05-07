package config

import (
	"fmt"
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
	Name             string
	Environment      string
	UIDomainName     string
	APIDomainName    string
	AllowSignUp      bool
	Beta             Beta
	CORS             CORS
	JWT              JWT
	Logging          Logging
	Plaid            Plaid
	PostgreSQL       PostgreSQL
	ReCAPTCHA        ReCAPTCHA
	Redis            Redis
	SMTP             SMTPClient
	SendGrid         SendGrid
	Sentry           Sentry
	Stripe           Stripe
	MultiPlaidConfig map[PlaidEnvironment]PlaidSingleConfig
}

type Beta struct {
	EnableBetaCodes bool
}

type PlaidSingleConfig struct {
	ClientID     string
	ClientSecret string
}

type JWT struct {
	LoginJwtSecret        string
	RegistrationJwtSecret string
}

type PostgreSQL struct {
	Address  string
	Port     int
	Username string
	Password string
	Database string
}

type SMTPClient struct {
	Enabled  bool
	Identity string
	Username string
	Password string
	Host     string
	Port     int

	VerifyEmails bool
}

type SendGrid struct {
	Enabled   bool
	APIKey    string
	Templates map[SendGridTemplate]string
}

type ReCAPTCHA struct {
	Enabled    bool
	PublicKey  string
	PrivateKey string
	Version    int // Currently only version 2 is supported by the UI.

	VerifyLogin    bool
	VerifyRegister bool
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
	Level string
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
	v.SetDefault("Environment", "development")
	v.SetDefault("UIDomainName", "localhost:3000")
	v.SetDefault("APIDomainName", "localhost:4000")
	v.SetDefault("AllowSignUp", true)
	v.SetDefault("PostgreSQL.Port", 5432)
	v.SetDefault("PostgreSQL.Address", "localhost")
	v.SetDefault("PostgreSQL.Username", "postgres")
	v.SetDefault("PostgreSQL.Database", "postgres")
	v.SetDefault("SMTP.Enabled", false)
	v.SetDefault("ReCAPTCHA.Enabled", false)
	v.SetDefault("Logging.Level", "info")
}

func setupEnv(v *viper.Viper) {
	v.BindEnv("Name", "MONETR_NAME")
	v.BindEnv("Environment", "MONETR_ENVIRONMENT")
	v.BindEnv("UIDomainName", "MONETR_UI_DOMAIN_NAME")
	v.BindEnv("APIDomainName", "MONETR_API_DOMAIN_NAME")
	v.BindEnv("AllowSignUp", "MONETR_ALLOW_SIGN_UP")
	v.BindEnv("EnableWebhooks", "MONETR_ENABLE_WEBHOOKS")
	v.BindEnv("Beta.EnableBetaCodes", "MONETR_ENABLE_BETA_CODES")
	v.BindEnv("Cors.AllowedOrigins", "MONETR_CORS_ALLOWED_ORIGINS")
	v.BindEnv("Cors.Debug", "MONETR_CORS_DEBUG")
	v.BindEnv("JWT.LoginJwtSecret", "MONETR_JWT_LOGIN_SECRET")
	v.BindEnv("JWT.RegistrationJwtSecret", "MONETR_JWT_REGISTRATION_SECRET")
	v.BindEnv("Logging.Level", "MONETR_LOG_LEVEL")
	v.BindEnv("Plaid.ClientID", "MONETR_PLAID_CLIENT_ID")
	v.BindEnv("Plaid.ClientSecret", "MONETR_PLAID_CLIENT_SECRET")
	v.BindEnv("Plaid.Environment", "MONETR_PLAID_ENVIRONMENT")
	v.BindEnv("Plaid.EnableBirthdatePrompt", "MONETR_PLAID_BIRTHDATE_PROMPT")
	v.BindEnv("Plaid.EnableReturningUserExperience", "MONETR_PLAID_RETURNING_EXPERIENCE")
	v.BindEnv("Plaid.WebhooksEnabled", "MONETR_PLAID_WEBHOOKS_ENABLED")
	v.BindEnv("Plaid.WebhooksDomain", "MONETR_PLAID_WEBHOOKS_DOMAIN")
	v.BindEnv("PostgreSQL.Address", "MONETR_PG_ADDRESS")
	v.BindEnv("PostgreSQL.Port", "MONETR_PG_PORT")
	v.BindEnv("PostgreSQL.Username", "MONETR_PG_USERNAME")
	v.BindEnv("PostgreSQL.Password", "MONETR_PG_PASSWORD")
	v.BindEnv("PostgreSQL.Database", "MONETR_PG_DATABASE")
	v.BindEnv("ReCAPTCHA.Enabled", "MONETR_CAPTCHA_ENABLED")
	v.BindEnv("ReCAPTCHA.PublicKey", "MONETR_CAPTCHA_PUBLIC_KEY")
	v.BindEnv("ReCAPTCHA.PrivateKey", "MONETR_CAPTCHA_PRIVATE_KEY")
	v.BindEnv("ReCAPTCHA.VerifyLogin", "MONETR_CAPTCHA_VERIFY_LOGIN")
	v.BindEnv("ReCAPTCHA.VerifyRegister", "MONETR_CAPTCHA_VERIFY_REGISTER")
	v.BindEnv("Redis.Enabled", "MONETR_REDIS_ENABLED")
	v.BindEnv("Redis.Address", "MONETR_REDIS_ADDRESS")
	v.BindEnv("Redis.Port", "MONETR_REDIS_PORT")
	v.BindEnv("Redis.Namespace", "MONETR_REDIS_NAMESPACE")
	v.BindEnv("Sentry.Enabled", "MONETR_SENTRY_ENABLED")
	v.BindEnv("Sentry.DSN", "MONETR_SENTRY_DSN")
	v.BindEnv("Sentry.SampleRate", "MONETR_SENTRY_SAMPLE_RATE")
	v.BindEnv("Sentry.TraceSampleRate", "MONETR_SENTRY_TRACE_SAMPLE_RATE")
	v.BindEnv("Stripe.Enabled", "MONETR_STRIPE_ENABLED")
	v.BindEnv("Stripe.APIKey", "MONETR_STRIPE_API_KEY")
	v.BindEnv("Stripe.PublicKey", "MONETR_STRIPE_PUBLIC_KEY")
	v.BindEnv("Stripe.WebhooksEnabled", "MONETR_STRIPE_WEBHOOKS_ENABLED")
	v.BindEnv("Stripe.WebhooksDomain", "MONETR_STRIPE_WEBHOOKS_DOMAIN")
	v.BindEnv("Stripe.WebhookSecret", "MONETR_STRIPE_WEBHOOK_SECRET")
}
