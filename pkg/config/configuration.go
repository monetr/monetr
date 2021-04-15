package config

import (
	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/viper"
)

const EnvironmentPrefix = "HARDER"

type Configuration struct {
	Name           string
	UIDomainName   string
	APIDomainName  string
	AllowSignUp    bool
	EnableWebhooks bool
	CORS           CORS
	JWT            JWT
	Logging        Logging
	Plaid          Plaid
	PostgreSQL     PostgreSQL
	ReCAPTCHA      ReCAPTCHA
	Redis          Redis
	SMTP           SMTPClient
	SendGrid       SendGrid
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
	Enabled bool
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

func LoadConfiguration() Configuration {
	v := viper.GetViper()

	v.SetEnvPrefix(EnvironmentPrefix)
	v.AutomaticEnv()

	setupDefaults(v)
	setupEnv(v)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/monetr/")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	var config Configuration
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	return config
}

func setupDefaults(v *viper.Viper) {
	v.SetDefault("Name", "Harder Than It Needs To Be")
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
	v.BindEnv("Name", "HARDER_NAME")
	v.BindEnv("UIDomainName", "HARDER_UI_DOMAIN_NAME")
	v.BindEnv("APIDomainName", "HARDER_API_DOMAIN_NAME")
	v.BindEnv("AllowSignUp", "HARDER_ALLOW_SIGN_UP")
	v.BindEnv("EnableWebhooks", "HARDER_ENABLE_WEBHOOKS")
	v.BindEnv("Cors.AllowedOrigins", "HARDER_CORS_ALLOWED_ORIGINS")
	v.BindEnv("Cors.Debug", "HARDER_CORS_DEBUG")
	v.BindEnv("JWT.LoginJwtSecret", "HARDER_JWT_LOGIN_SECRET")
	v.BindEnv("JWT.RegistrationJwtSecret", "HARDER_JWT_REGISTRATION_SECRET")
	v.BindEnv("Logging.Level", "HARDER_LOG_LEVEL")
	v.BindEnv("Plaid.ClientID", "HARDER_PLAID_CLIENT_ID")
	v.BindEnv("Plaid.ClientSecret", "HARDER_PLAID_CLIENT_SECRET")
	v.BindEnv("Plaid.Environment", "HARDER_PLAID_ENVIRONMENT")
	v.BindEnv("Plaid.EnableBirthdatePrompt", "HARDER_PLAID_BIRTHDATE_PROMPT")
	v.BindEnv("PostgreSQL.Address", "HARDER_PG_ADDRESS")
	v.BindEnv("PostgreSQL.Port", "HARDER_PG_PORT")
	v.BindEnv("PostgreSQL.Username", "HARDER_PG_USERNAME")
	v.BindEnv("PostgreSQL.Password", "HARDER_PG_PASSWORD")
	v.BindEnv("PostgreSQL.Database", "HARDER_PG_DATABASE")
	v.BindEnv("ReCAPTCHA.Enabled", "HARDER_CAPTCHA_ENABLED")
	v.BindEnv("ReCAPTCHA.PublicKey", "HARDER_CAPTCHA_PUBLIC_KEY")
	v.BindEnv("ReCAPTCHA.PrivateKey", "HARDER_CAPTCHA_PRIVATE_KEY")
	v.BindEnv("ReCAPTCHA.VerifyLogin", "HARDER_CAPTCHA_VERIFY_LOGIN")
	v.BindEnv("ReCAPTCHA.VerifyRegister", "HARDER_CAPTCHA_VERIFY_REGISTER")
	v.BindEnv("Redis.Enabled", "HARDER_REDIS_ENABLED")
	v.BindEnv("Redis.Address", "HARDER_REDIS_ADDRESS")
	v.BindEnv("Redis.Port", "HARDER_REDIS_PORT")
	v.BindEnv("Redis.Namespace", "HARDER_REDIS_NAMESPACE")
}
