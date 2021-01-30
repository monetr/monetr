package config

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
