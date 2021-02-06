package config

import (
	"github.com/spf13/viper"
)

type Configuration struct {
	Name           string
	UIDomainName   string
	APIDomainName  string
	AllowSignUp    bool
	EnableWebhooks bool
	JWT            JWT
	PostgreSQL     PostgreSQL
	SMTP           SMTPClient
	ReCAPTCHA      ReCAPTCHA
	Plaid          Plaid
	CORS           CORS
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

func LoadConfiguration() Configuration {
	viper.SetDefault("Name", "Harder Than It Needs To Be")
	viper.SetDefault("UIDomainName", "localhost:3000")
	viper.SetDefault("APIDomainName", "localhost:4000")
	viper.SetDefault("AllowSignUp", true)
	viper.SetDefault("PostgreSQL.Port", 5432)
	viper.SetDefault("PostgreSQL.Address", "localhost")
	viper.SetDefault("PostgreSQL.Username", "postgres")
	viper.SetDefault("PostgreSQL.Database", "postgres")
	viper.SetDefault("SMTP.Enabled", false)
	viper.SetDefault("ReCAPTCHA.Enabled", false)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/harder/")
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
