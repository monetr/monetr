package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/monetr/monetr/server/util"
	"github.com/plaid/plaid-go/v20/plaid"
	"github.com/spf13/viper"
)

var (
	// FilePath is set via a flag in pkg/cmd.
	FilePath []string
	// LogLevel is set via a flag in pkg/cmd.
	LogLevel string
)

type PlaidEnvironment string

const (
	PlaidSandbox     PlaidEnvironment = "sandbox"
	PlaidDevelopment PlaidEnvironment = "development"
	PlaidProduction  PlaidEnvironment = "production"
)

type Configuration struct {
	// configFile is not an actual configuration variable, but is used to let usages know what file was loaded for the
	// configuration.
	configFile string `yaml:"-"`

	Environment string `yaml:"environment"`
	// ExternalURLProtocol is used to determine what protocol should be used in things like email templates. It defaults
	// to https.
	ExternalURLProtocol string         `yaml:"externalUrlProtocol"`
	UIDomainName        string         `yaml:"uiDomainName"`
	APIDomainName       string         `yaml:"apiDomainName"`
	AllowSignUp         bool           `yaml:"allowSignUp"`
	BackgroundJobs      BackgroundJobs `yaml:"backgroundJobs"`
	Beta                Beta           `yaml:"beta"`
	CORS                CORS           `yaml:"cors"`
	Email               Email          `yaml:"email"`
	KeyManagement       KeyManagement  `yaml:"keyManagement"`
	Links               Links          `yaml:"links"`
	Logging             Logging        `yaml:"logging"`
	Plaid               Plaid          `yaml:"plaid"`
	PostgreSQL          PostgreSQL     `yaml:"postgreSql"`
	RabbitMQ            RabbitMQ       `yaml:"rabbitMQ"`
	ReCAPTCHA           ReCAPTCHA      `yaml:"reCAPTCHA"`
	Redis               Redis          `yaml:"redis"`
	Security            Security       `yaml:"security"`
	Sentry              Sentry         `yaml:"sentry"`
	Server              Server         `yaml:"server"`
	Storage             Storage        `yaml:"storage"`
	Stripe              Stripe         `yaml:"stripe"`
}

func (c Configuration) GetConfigFileName() string {
	return c.configFile
}

func (c Configuration) GetUIDomainName() string {
	return c.UIDomainName
}

func (c Configuration) GetHTTPSecureCookie() bool {
	return c.ExternalURLProtocol == "https" && c.Server.Cookies.Secure
}

func (c Configuration) GetUIURL() string {
	return fmt.Sprintf("%s://%s", c.ExternalURLProtocol, c.UIDomainName)
}

type Storage struct {
	// Enabled controls whether or not monetr can actually store files. If this is disabled then some monetr features will
	// not be available. These features include things like file imports for transactions.
	Enabled bool `yaml:"enabled"`
	// Provider specifies which storage backend monetr should use. Allowed values are:
	// - `s3`
	// - `gcs`
	// - `filesystem`
	// Note: If you use the filesystem backend you cannot run multiple monetr servers. Even if the filesystem is shared
	// between the instances via something like NFS; it can cause unpredictable behavior. It is recommended to use GCS or
	// S3 backed storage when you are running multiple instances of monetr.
	// If you are self-hosting monetr though as a single instance, then filesystem is the recommended storage backend for
	// ease of use.
	Provider   string             `yaml:"provider"`
	S3         *S3Storage         `yaml:"s3"`
	GCS        *GCSStorage        `yaml:"gcs"`
	Filesystem *FilesystemStorage `yaml:"filesystem"`
}

type S3Storage struct {
	Endpoint          *string `yaml:"endpoint"`
	Bucket            string  `yaml:"bucket"`
	UseEnvCredentials bool    `yaml:"useEnvCredentials"`
	Region            string  `yaml:"region"`
	ForcePathStyle    bool    `yaml:"forcePathStyle"`
	AccessKey         *string `yaml:"accessKeyId"`
	SecretKey         *string `yaml:"secretAccessKey"`
}

type GCSStorage struct {
	URL             *string `yaml:"url"`
	Bucket          string  `yaml:"bucket"`
	APIKey          *string `yaml:"apiKey"`
	CredentialsJSON *string `yaml:"credentialsFile"`
}

type FilesystemStorage struct {
	BasePath string `yaml:"basePath"`
}

// KeyManagement specifies the properties required to securely encrypt and
// decrypt stored secrets. It is not recommended to change providers.
type KeyManagement struct {
	Provider string `yaml:"provider"`
	// AWS provides configuration for using AWS's KMS for encrypting and decrypting secrets.
	AWS AWSKMS `yaml:"aws"`
	// Google provides configuration for using Google's KMS for encrypting and decrypting secrets.
	Google GoogleKMS `yaml:"google"`
}

type AWSKMS struct {
	Region    string  `yaml:"region"`
	AccessKey string  `yaml:"accessKey"`
	SecretKey string  `yaml:"secretKey"`
	KeyID     string  `yaml:"keyID"`
	Endpoint  *string `yaml:"endpoint"`
}

type GoogleKMS struct {
	CredentialsJSON *string `yaml:"credentialsJSON"`
	ResourceName    string  `yaml:"resourceName"`
}

type Server struct {
	// ListenPort defines the port that monetr will listen for HTTP requests on. This port should be forwarded such that
	// it is accessible to the desired clients. Be that on a local network, or forwarded to the public internet.
	ListenPort int `yaml:"listenPort"`
	// ListenAddress defines the IP address that monetr should listen on for HTTP requests.
	ListenAddress string `yaml:"listenAddress"`
	// StatsPort is the port that our prometheus metrics are served on. This port should not be publicly accessible and
	// should only be accessible by the prometheus server scraping for metrics. It is not an endpoint that needs to be
	// secured as no sensitive client information will be served by it; but it should not be accessible publicly.
	StatsPort int `yaml:"statsPort"`
	// Cookies defines the parameters used for issuing and processing cookies from clients. Cookies are used for
	// authentication.
	Cookies Cookies `yaml:"cookies"`
	// UICacheHours is the number of hours that UI files should be cached by the client. This is done by including an
	// Expires and Cache-Control header in the response for all UI related requests. If this is 0 then the headers will
	// not be included. Defaults to 12 hours.
	UICacheHours int `yaml:"uiCacheHours"`
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

type Security struct {
	// PublicKey is a path to the file containing the ED22519 public key in pem format.
	PublicKey string `yaml:"publicKey"`
	// PrivateKey is the path to the file containing the ED22519 private key in pem format.
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
	// VerifyForgotPassword determines whether or not the user will be required to verify that they are not a robot
	// overlord.
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
	FreeTrialDays  int    `yaml:"freeTrialDays"`
}

// IsBillingEnabled will return true if both Stripe and Billing are enabled. It will return false any other time.
func (s Stripe) IsBillingEnabled() bool {
	return s.Enabled && s.BillingEnabled
}

type RabbitMQ struct {
	Enabled  bool   `yaml:"enabled"`
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type BackgroundJobEngine string

const (
	// BackgroundJobEngineInMemory is ideal for self-hosted or development
	// deployments. It is not recommended at all for any type of deployment that
	// demands reliability or high-availability. When using this option you
	// **cannot** deploy more than a single replica of monetr. This is because
	// jobs will not be shared between replicas, and you can end up with problems
	// where scheduled jobs are performed multiple times by each replica. This
	// option must be manually specified and is only the default configuration
	// when Redis and RabbitMQ are disabled. **NOT YET IMPLEMENTED**
	BackgroundJobEngineInMemory BackgroundJobEngine = "memory"
	// BackgroundJobEngineGoCraftWork will soon be deprecated as monetr moves
	// towards using PostgreSQL as its primary job queue. It should not be used
	// anymore and will be removed before the V1.0 release.
	// TODO Remove before the V1.0 release.
	BackgroundJobEngineGoCraftWork BackgroundJobEngine = "gocraft"
	// BackgroundJobEnginePostgreSQL is the preferred option for job queues, as it
	// provides some consistency guarantees that Redis cannot. Specifically making
	// sure that a job is not processed more than once.
	BackgroundJobEnginePostgreSQL BackgroundJobEngine = "postgresql"
)

type BackgroundJobScheduler string

const (
	// BackgroundJobSchedulerExternal requires that some process outside the monetr instance add scheduled jobs to the
	// queue in order to be processed. This is required for deployments using the RabbitMQ background job engine.
	BackgroundJobSchedulerExternal BackgroundJobScheduler = "external"
	// BackgroundJobSchedulerInternal adds jobs to the queue within the monetr process and does not require any
	// additional instances or external processes for jobs to be performed. This is only available for in-memory or
	// gocraft/work job engine deployments. This also requires that the cron schedules for each queue be specified.
	BackgroundJobSchedulerInternal BackgroundJobScheduler = "internal"
)

type BackgroundJobs struct {
	Engine      BackgroundJobEngine    `yaml:"engine"`
	Scheduler   BackgroundJobScheduler `yaml:"scheduler"`
	JobSchedule map[string]string      `yaml:"jobSchedule"`
}

func getViper(configFilePath []string) *viper.Viper {
	v := viper.GetViper()
	v.SetConfigType("yaml")

	setupDefaults(v)
	setupEnv(v)

	switch len(FilePath) {
	case 0:
		{ // If we can determine the user's home directory, then look there + /.sentry for the config
			homeDir, err := os.UserHomeDir()
			if err == nil {
				v.AddConfigPath(homeDir + "/.monetr/config.yaml")
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

	{ // Background job processing defaults.
		if config.BackgroundJobs.JobSchedule == nil {
			config.BackgroundJobs.JobSchedule = map[string]string{}
		}

		switch config.BackgroundJobs.Engine {
		case BackgroundJobEngineInMemory:
			log.Fatal("in-memory job scheduling is not yet implemented")
		case BackgroundJobEngineGoCraftWork:
			if config.BackgroundJobs.Scheduler == "" {
				config.BackgroundJobs.Scheduler = BackgroundJobSchedulerInternal
			}

			// Make sure that the scheduler specified in the configuration is valid.
			switch config.BackgroundJobs.Scheduler {
			case BackgroundJobSchedulerInternal:
			// The scheduler is configured correctly for the gocraft/work engine.
			default:
				log.Fatal("invalid scheduler provided to configuration for background jobs")
				return
			}
		}
	}

	privateKey, err := util.ExpandHomePath(config.Security.PrivateKey)
	if err != nil {
		panic(err)
	}
	config.Security.PrivateKey = privateKey

	return config
}

func setupDefaults(v *viper.Viper) {
	v.SetDefault("APIDomainName", "0.0.0.0:4000")
	v.SetDefault("AllowSignUp", true)
	v.SetDefault("ExternalURLProtocol", "https")
	v.SetDefault("BackgroundJobs.Engine", BackgroundJobEngineGoCraftWork)
	v.SetDefault("BackgroundJobs.Scheduler", BackgroundJobSchedulerInternal)
	v.SetDefault("Email.ForgotPassword.TokenLifetime", 10*time.Minute)
	v.SetDefault("Email.Verification.TokenLifetime", 10*time.Minute)
	v.SetDefault("Environment", "development")
	v.SetDefault("Logging.Format", "text")
	v.SetDefault("Logging.Level", LogLevel) // Info
	v.SetDefault("Logging.StackDriver.Enabled", false)
	v.SetDefault("KeyManagement.Provider", "plaintext")
	v.SetDefault("KeyManagement.AWS", nil)
	v.SetDefault("KeyManagement.Google", nil)
	v.SetDefault("Plaid.Enabled", true)
	v.SetDefault("PostgreSQL.Address", "localhost")
	v.SetDefault("PostgreSQL.Database", "postgres")
	v.SetDefault("PostgreSQL.Port", 5432)
	v.SetDefault("PostgreSQL.Username", "postgres")
	v.SetDefault("Redis.Port", 6379)
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
	v.SetDefault("Server.UICacheHours", 12)
	v.SetDefault("Stripe.FreeTrialDays", 30)
	v.SetDefault("UIDomainName", "0.0.0.0:4000")
}

func setupEnv(v *viper.Viper) {
	_ = v.BindEnv("Environment", "MONETR_ENVIRONMENT")
	_ = v.BindEnv("UIDomainName", "MONETR_UI_DOMAIN_NAME")
	_ = v.BindEnv("APIDomainName", "MONETR_API_DOMAIN_NAME")
	_ = v.BindEnv("AllowSignUp", "MONETR_ALLOW_SIGN_UP")
	_ = v.BindEnv("ExternalURLProtocol", "MONETR_EXTERNAL_PROTOCOL")
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
	_ = v.BindEnv("Logging.Level", "MONETR_LOG_LEVEL")
	_ = v.BindEnv("Logging.Format", "MONETR_LOG_FORMAT")
	_ = v.BindEnv("Logging.StackDriver.Enabled", "MONETR_LOG_STACKDRIVER_ENABLED")
	_ = v.BindEnv("KeyManagement.Provider", "MONETR_KMS_PROVIDER")
	_ = v.BindEnv("KeyManagement.AWS.AccessKey", "AWS_ACCESS_KEY_ID")
	_ = v.BindEnv("KeyManagement.AWS.SecretKey", "AWS_ACCESS_KEY")
	_ = v.BindEnv("KeyManagement.Google.ResourceName", "MONETR_KMS_RESOURCE_NAME")
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
	_ = v.BindEnv("Stripe.InitialPlan.StripePriceId", "MONETR_STRIPE_DEFAULT_PRICE_ID")
}
