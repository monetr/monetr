package config

import (
	"github.com/spf13/viper"
)

type Configuration struct {
	JWTSecret     string
	UIDomainName  string
	APIDomainName string
	PostgreSQL    PostgreSQL
	SMTP          SMTPClient
	ReCAPTCHA     ReCAPTCHA
	AllowSignUp   bool
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

	VerifyEmails bool
}

type ReCAPTCHA struct {
	Enabled    bool
	PublicKey  string
	PrivateKey string
	Version    int

	VerifyLogin    bool
	VerifyRegister bool
}

func LoadConfiguration() Configuration {
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
