package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/util"
	"github.com/plaid/plaid-go/v42/plaid"
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
	Features      Features      `yaml:"transactionImports"`
	KeyManagement KeyManagement `yaml:"keyManagement"`
	Links         Links         `yaml:"links"`
	Logging       Logging       `yaml:"logging"`
	LunchFlow     LunchFlow     `yaml:"lunchFlow"`
	Plaid         Plaid         `yaml:"plaid"`
	PostgreSQL    PostgreSQL    `yaml:"postgreSql"`
	ProofOfWork   ProofOfWork   `yaml:"proofOfWork"`
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

// ProofOfWork gates the unauthenticated auth endpoints (register, login, forgot
// password, resend verification): the client must solve a small SHA-256
// challenge before its request is accepted. It makes automated abuse expensive
// without slowing real users (the work runs in a background web worker).
type ProofOfWork struct {
	// Active when true. When disabled the challenge endpoint 404s and the auth
	// endpoints skip the check. Disabled by default for now, set
	// MONETR_PROOF_OF_WORK_ENABLED=true to turn it on.
	Enabled bool `yaml:"enabled"`
	// Leading zero bits the solution must have; each extra bit doubles the work.
	// 16 is roughly 1 second on a desktop, 2 to 4 on a low end phone.
	Difficulty int `yaml:"difficulty"`
	// How long a challenge is valid for. Long enough to fill out the form, short
	// enough that a stolen challenge is not useful for long.
	Lifetime time.Duration `yaml:"lifetime"`
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
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
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

// GetExternalOrigin will return the scheme and host (including port if one
// was specified) of the ExternalDSN. It will return an empty string if
// external Sentry is not enabled, the DSN is not a valid URL, does not use
// http or https, or points at localhost. This is used by the UI to decide
// whether to emit a preconnect link tag pointing at the Sentry ingestion
// host.
func (s Sentry) GetExternalOrigin() string {
	if !s.ExternalSentryEnabled() {
		return ""
	}
	u, err := url.Parse(s.ExternalDSN)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return ""
	}
	return u.Scheme + "://" + u.Host
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

// stringSliceToSetHookFunc lets mapstructure (and therefore viper) decode a
// yaml list, or a comma separated env var, into a myownsanity.Set. Without this
// viper has no idea how to put a list into our map based Set type and the whole
// config load panics. The Set has a json unmarshaler but viper uses
// mapstructure not encoding/json, so that does us no good here.
func stringSliceToSetHookFunc() mapstructure.DecodeHookFunc {
	setType := reflect.TypeOf(myownsanity.Set[string]{})
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if to != setType {
			return data, nil
		}

		// Depending on where the value came from mapstructure hands us different
		// shapes. A yaml list comes through as a []any, the env var comes through
		// as a single comma separated string.
		set := myownsanity.NewSet[string]()
		switch value := data.(type) {
		case []any:
			for _, item := range value {
				if s, ok := item.(string); ok {
					set.Add(s)
				}
			}
		case []string:
			for _, item := range value {
				set.Add(item)
			}
		case string:
			for _, item := range strings.Split(value, ",") {
				if item = strings.TrimSpace(item); item != "" {
					set.Add(item)
				}
			}
		default:
			// Not a shape we know how to turn into a set, hand it back untouched
			// so mapstructure can take its normal path and surface a real error
			// instead of us hiding a misconfiguration.
			return data, nil
		}

		return set, nil
	}
}

func LoadConfigurationEx(v *viper.Viper) (config Configuration) {
	// We have to spell out the decode hooks ourselves because passing
	// viper.DecodeHook completely replaces viper's defaults, so the first two
	// here are just the defaults (duration parsing and comma separated slices)
	// and the last one is ours that knows how to turn a list into our Set type.
	if err := v.Unmarshal(&config, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			stringSliceToSetHookFunc(),
		),
	)); err != nil {
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
	// Lunch Flow is enabled by default for self-hosted deployments!
	v.SetDefault("LunchFlow.Enabled", true)
	v.SetDefault("LunchFlow.AllowedApiUrls", []string{DefaultLunchFlowAPIURL})
	v.SetDefault("KeyManagement.Provider", "plaintext")
	v.SetDefault("Plaid.Enabled", true)
	v.SetDefault("Plaid.CountryCodes", []plaid.CountryCode{plaid.COUNTRYCODE_US})
	// Disabled by default for the first release so existing deployments opt in;
	// likely flips to true in a later release once announced.
	v.SetDefault("ProofOfWork.Enabled", false)
	v.SetDefault("ProofOfWork.Difficulty", 16)
	v.SetDefault("ProofOfWork.Lifetime", 5*time.Minute)
	v.SetDefault("PostgreSQL.Address", "localhost")
	v.SetDefault("PostgreSQL.Database", "postgres")
	v.SetDefault("PostgreSQL.Port", 5432)
	v.SetDefault("PostgreSQL.Username", "postgres")
	v.SetDefault("Redis.Port", 6379)
	v.SetDefault("Redis.Database", 0)
	v.SetDefault("Security.PrivateKey", "/etc/monetr/ed25519.key")
	v.SetDefault("Sentry.SampleRate", 1.0)
	v.SetDefault("Sentry.TraceSampleRate", 1.0)
	v.SetDefault("Server.Cookies.Name", "M-Token")
	v.SetDefault("Server.Cookies.Secure", true)
	v.SetDefault("Server.Cookies.SameSiteStrict", true)
	v.SetDefault("Server.ListenPort", 4000)
	v.SetDefault("Server.ListenAddress", "0.0.0.0")
	v.SetDefault("Server.StatsPort", 9000)
	v.SetDefault("Server.UICacheHours", 90*24)
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
	v.MustBindEnv("Email.BlockedDomains", "MONETR_EMAIL_BLOCKED_DOMAINS")
	v.MustBindEnv("Logging.Level", "MONETR_LOG_LEVEL")
	v.MustBindEnv("Logging.Format", "MONETR_LOG_FORMAT")
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
	v.MustBindEnv("ProofOfWork.Enabled", "MONETR_PROOF_OF_WORK_ENABLED")
	v.MustBindEnv("ProofOfWork.Difficulty", "MONETR_PROOF_OF_WORK_DIFFICULTY")
	v.MustBindEnv("ProofOfWork.Lifetime", "MONETR_PROOF_OF_WORK_LIFETIME")
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
