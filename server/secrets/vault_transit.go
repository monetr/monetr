package secrets

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	vault "github.com/hashicorp/vault/api"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/round"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ KeyManagement = &VaultTransit{}
)

type VaultTransitConfig struct {
	Log                *logrus.Entry
	KeyID              string
	Address            string
	Role               string
	AuthMethod         string
	Token              string
	TokenFile          string
	Username, Password string
	Timeout            time.Duration
	TLSCertificatePath string
	TLSKeyPath         string
	TLSCAPath          string
	InsecureSkipVerify bool
	IdleConnTimeout    time.Duration
}

type VaultTransit struct {
	tokenTTL        sync.Once
	tokenSync       sync.RWMutex
	tokenExpiration int64
	tokenCloser     chan chan error
	host            string
	config          VaultTransitConfig
	log             *logrus.Entry
	client          *vault.Client
	usingCustomTLS  bool
	tlsWatch        sync.Once
	lock            sync.RWMutex
	tls             *tls.Config
	closer          chan chan error
}

func NewVaultTransit(
	ctx context.Context,
	config VaultTransitConfig,
) (*VaultTransit, error) {
	log := config.Log
	host, err := url.Parse(config.Address)
	if err != nil {
		log.WithField("url", config.Address).WithError(err).Errorf("failed to parse vault URL")
		return nil, errors.Wrap(err, "failed to parse vault URL")
	}

	helper := &VaultTransit{
		host:           host.Hostname(),
		config:         config,
		log:            log,
		client:         nil,
		usingCustomTLS: len(config.TLSCAPath) > 0,
		tls:            nil,
		closer:         nil,
	}

	trip := round.NewObservabilityRoundTripper(&http.Transport{
		IdleConnTimeout: config.IdleConnTimeout,
	}, func(ctx context.Context, request *http.Request, response *http.Response, err error) {
		logEntry := log.WithContext(ctx).WithFields(logrus.Fields{
			"vault": logrus.Fields{
				"method": request.Method,
				"url":    request.URL.String(),
			},
		})

		if err != nil {
			logEntry = logEntry.WithError(err)
		}
		logEntry.Trace("making request to vault API")
	})

	vaultConfig := &vault.Config{
		Address: config.Address,
		HttpClient: &http.Client{
			Transport: trip,
			Timeout:   config.Timeout,
		},
		MaxRetries:       3,
		Timeout:          config.Timeout,
		Error:            nil,
		Backoff:          nil,
		CheckRetry:       nil,
		Limiter:          nil,
		OutputCurlString: false,
		SRVLookup:        false,
	}

	// If we have a custom certificate authority specified then we are using
	// custom TLS certificates too. We need to watch the certificate files to see
	// if they change at all.
	if helper.usingCustomTLS {
		if err = helper.reloadTLS(); err != nil {
			log.WithError(err).Errorf("failed to configure TLS")
			return nil, err
		}
		helper.watchCertificates()
		vaultConfig.HttpClient.Transport = &http.Transport{
			DialTLSContext:  helper.dialTLS,
			IdleConnTimeout: config.IdleConnTimeout,
		}
	}

	helper.client, err = vault.NewClient(vaultConfig)
	if err != nil {
		log.WithError(err).Errorf("failed to create vault client")
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	if err = helper.authenticate(ctx); err != nil {
		return nil, err
	}

	// Start the authentication worker.
	helper.authenticationWorker()

	return helper, nil
}

// dialTLS is a middleware function that is added to allow monetr to easily
// rotate the TLS certificates for the vault server without downtime.
func (v *VaultTransit) dialTLS(
	ctx context.Context,
	network, addr string,
) (net.Conn, error) {
	v.lock.RLock()
	defer v.lock.RUnlock()

	return tls.Dial(network, addr, v.tls)
}

// This function sets up a file system watcher to monitor changes in TLS
// certificate files. When changes are detected, it calls reloadTLS which
// atomically swaps the TLS config that is used to establish new connections to
// the vault server.
//
//	watchCertificates()
//	├── Uses sync.Once to ensure the watcher setup runs only once
//	│   └── tlsWatch.Do()
//	│       └── go routine for asynchronous execution
//	│           ├── Initializes the logger and file watcher
//	│           ├── Sets up a channel for closing the watcher
//	│           ├── Defines paths to be watched:
//	│           │   ├── TLSCertificatePath
//	│           │   ├── TLSKeyPath
//	│           │   └── TLSCAPath
//	│           ├── Adds paths to the watcher
//	│           └── Event loop:
//	│               ├── Handles errors from watcher.Errors
//	│               ├── Handles events from watcher.Events
//	│               │   └── Calls reloadTLS() on file changes
//	│               └── Handles closure of the watcher via v.closer channel
//
// This function is called when the client is initialized and runs until Close()
// is called on the client.
func (v *VaultTransit) watchCertificates() {
	v.tlsWatch.Do(func() {
		go func() {
			log := v.log
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				log.WithError(err).Errorf("failed to create file system watcher, vault TLS certificates cannot be auto rotated")
				return
			}

			v.closer = make(chan chan error, 1)

			paths := []string{
				v.config.TLSCertificatePath,
				v.config.TLSKeyPath,
				v.config.TLSCAPath,
			}
			for _, path := range paths {
				fileLog := log.WithField("file", path)
				fileLog.Trace("watching file for changes for vault TLS")
				if err = watcher.Add(v.config.TLSCertificatePath); err != nil {
					fileLog.WithError(err).Errorf("failed to watch file for vault TLS")
				}
			}

			for {
				select {
				case err = <-watcher.Errors:
					log.WithError(err).Warn("error watching file for vault TLS")
				case event := <-watcher.Events:
					log.WithField("file", event.Name).Trace("observed changed in vault TLS file")
					if err = v.reloadTLS(); err != nil {
						log.WithError(err).Errorf("failed to reload vault TLS")
					}
				case promise := <-v.closer:
					log.Info("closing vault helper TLS watcher")
					promise <- watcher.Close()
					return
				}
			}
		}()
	})
}

// This function reloads the TLS configuration by reading the certificate files
// and updating the TLS configuration to be used for subsequent requests.
//
//	reloadTLS()
//	├── Reads CA certificate from file
//	│   ├── Adds CA certificate to a new cert pool
//	├── Configures tls.Config with:
//	│   ├── CA certificate pool
//	│   ├── InsecureSkipVerify from config
//	│   ├── ServerName from host
//	│   └── Renegotiation setting
//	├── If key and certificate paths are provided:
//	│   ├── Loads TLS key pair
//	│   └── Adds key pair to tls.Config
//	├── Acquires lock on v.lock
//	│   └── Updates v.tls with new tls.Config
//	└── Returns nil if successful, otherwise an error
//
// This function is called when the vault client is initialized and every time a
// certificate change is detected.
func (v *VaultTransit) reloadTLS() error {
	log := v.log
	log.Debugf("reloading vault TLS config")

	caCert, err := os.ReadFile(v.config.TLSCAPath)
	if err != nil {
		log.WithField("file", v.config.TLSCAPath).Errorf("failed to read CA for vault TLS")
		return errors.Wrap(err, "failed to read CA for vault TLS")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: v.config.InsecureSkipVerify,
		RootCAs:            caCertPool,
		ServerName:         v.host,
		Renegotiation:      tls.RenegotiateFreelyAsClient,
	}

	if v.config.TLSKeyPath != "" || v.config.TLSCertificatePath != "" {
		tlsKeyPair, err := tls.LoadX509KeyPair(
			v.config.TLSCertificatePath,
			v.config.TLSKeyPath,
		)
		if err != nil {
			log.WithFields(logrus.Fields{
				"certPath": v.config.TLSCertificatePath,
				"keyPath":  v.config.TLSKeyPath,
			}).WithError(err).Errorf("failed to load TLS key pair for vault")
			return errors.Wrap(err, "failed to load TLS key pair for vault")
		}

		tlsConfig.Certificates = []tls.Certificate{
			tlsKeyPair,
		}
	}

	v.lock.Lock()
	defer v.lock.Unlock()

	v.tls = tlsConfig

	return nil
}

// This function sets up a worker that periodically checks the expiration status
// of the Vault token and refreshes it if necessary.
//
//	authenticationWorker()
//	├── Uses sync.Once to ensure the worker setup runs only once
//	│   └── tokenTTL.Do()
//	│       └── Checks if the token will never expire
//	│           └── If true, logs the info and exits
//	│       └── go routine for asynchronous execution
//	│           ├── Initializes a channel for closing the worker
//	│           ├── Sets the check frequency to 30 seconds
//	│           ├── Creates a ticker to trigger checks at the defined frequency
//	│           └── Event loop:
//	│               ├── Checks token expiration status at each tick
//	│               │   └── If token will expire before the next check, refreshes the token
//	│               │       └── Calls authenticate() to refresh the token
//	│               └── Handles closure of the worker via v.tokenCloser channel
//
// This function is called once when the client is created and runs a background
// job until Close() is called.
func (v *VaultTransit) authenticationWorker() {
	v.tokenTTL.Do(func() {
		log := v.log
		if atomic.LoadInt64(&v.tokenExpiration) == math.MaxInt64 {
			log.Info("vault token will never expire, background token refresher will not be started")
			return
		}

		go func() {
			log.Debug("vault token refresh worker has started, tokens will be refreshed before they expire")
			v.tokenCloser = make(chan chan error, 1)

			// Check to see if the token is going to expire every minute
			frequency := 30 * time.Second
			ticker := time.NewTicker(frequency)
			for {
				select {
				case <-ticker.C:
					// If the token will expire before we check it next, then refresh the token.
					if atomic.LoadInt64(&v.tokenExpiration) < time.Now().Add(frequency).Unix() {
						log.Debug("token will expire before the next check, refreshing token")
						err := func() error {
							ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
							defer cancel()
							return v.authenticate(ctx)
						}()
						if err != nil {
							log.WithError(err).Error("failed to refresh vault token")
						}
					}
				case promise := <-v.tokenCloser:
					log.Info("stopping refresh token worker")
					promise <- nil
					return
				}
			}
		}()
	})
}

// This function handles the authentication process to Vault using different
// methods based on the configuration.
//
//	authenticate(ctx context.Context) error
//	├── Acquires lock on v.tokenSync
//	├── Switches on v.config.Auth for different authentication methods:
//	│   ├── "userpass":
//	│   │   ├── Authenticates using username and password
//	│   │   ├── If successful, retrieves auth information
//	│   │   └── If failed, logs error and returns error
//	│   ├── "token":
//	│   │   └── Uses the provided token for authentication
//	│   ├── "kubernetes":
//	│   │   ├── Retrieves token from file or config
//	│   │   ├── Authenticates using the token
//	│   │   ├── If successful, retrieves auth information
//	│   │   └── If failed, logs error and returns error
//	│   └── default:
//	│       └── Logs an error and returns it if the auth method is not implemented
//	├── Logs fatal error if no authentication is returned
//	├── Calculates next token expiration time based on LeaseDuration
//	│   └── If LeaseDuration is 0, sets expiration to a far future timestamp
//	│   └── If LeaseDuration is not 0, sets expiration with a 1-minute buffer
//	├── Stores the new token expiration time atomically
//	├── Atomically updates the client's token with the new one
//	└── Returns nil if successful, otherwise an error
//
// This function is called by the authenticationWorker and simply refreshes the
// currently used authentication for communication with the vault server.
func (v *VaultTransit) authenticate(ctx context.Context) error {
	v.tokenSync.Lock()
	defer v.tokenSync.Unlock()
	log := v.log.WithField("method", v.config.AuthMethod)

	var auth *vault.SecretAuth
	switch v.config.AuthMethod {
	case "userpass":
		log.Trace("authenticating to vault")
		result, err := v.client.Logical().WriteWithContext(ctx, "auth/userpass/login/"+v.config.Username, map[string]any{
			"password": v.config.Password,
			"role":     v.config.Role,
		})
		if err != nil {
			log.WithError(err).Errorf("failed to authenticate to vault")
			return errors.Wrap(err, "failed to authenticate to vault")
		}
		auth = result.Auth
	case "token":
		auth = &vault.SecretAuth{
			ClientToken:   v.config.Token,
			LeaseDuration: 0,
		}
	case "kubernetes":
		log.Trace("authenticating to vault")
		var token string
		tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"

		switch {
		case v.config.Token != "":
			log.Trace("using included token")
			token = v.config.Token
		case v.config.TokenFile != "":
			tokenPath = v.config.TokenFile
			fallthrough
		default:
			fileLog := log.WithField("file", tokenPath)
			fileLog.Trace("reading token from specified file")
			data, err := os.ReadFile(tokenPath)
			if err != nil {
				fileLog.WithError(err).Error("failed to read token from specified file")
				return errors.Wrap(err, "failed to read token from specified file")
			}
			token = string(data)
		}

		log.Trace("authenticating to vault")
		result, err := v.client.Logical().WriteWithContext(
			ctx,
			"auth/kubernetes/login",
			map[string]any{
				"role": v.config.Role,
				"jwt":  token,
			},
		)
		if err != nil {
			log.WithError(err).Error("failed to authenticate to vault")
			return errors.Wrap(err, "failed to authenticate to vault")
		}
		auth = result.Auth
	default:
		return errors.Errorf("%s authentication not implemented", v.config.AuthMethod)
	}

	if auth == nil {
		log.Fatalf("no authentication returned from vault")
		return errors.Errorf("no authentication returned from vault")
	}

	var nextExpiration int64
	if auth.LeaseDuration == 0 {
		// If the token does not expire, store a timestamp so far in the future that we won't ever re-auth
		nextExpiration = math.MaxInt64
	} else {
		// If the token does expire, store the expiration time (minus 1 minute) to have a safe buffer.
		nextExpiration = time.Now().Add(time.Duration(auth.LeaseDuration)*time.Second - 1*time.Minute).Unix()
		log.Debugf("vault authentication will refresh by %s", time.Unix(nextExpiration, 0))
	}

	atomic.StoreInt64(&v.tokenExpiration, nextExpiration)

	v.client.SetToken(auth.ClientToken)
	log.Trace("successfully authenticated to vault")

	return nil
}

func (v *VaultTransit) Write(
	ctx context.Context,
	key string,
	value map[string]any,
) (*vault.Secret, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]any{
		"key": key,
	}

	log := v.log.WithFields(logrus.Fields{
		"key":    key,
		"action": "write",
	}).WithContext(span.Context())
	log.Trace("handling secret with vault")
	secret, err := v.client.Logical().WriteWithContext(
		span.Context(),
		key,
		value,
	)
	if err != nil {
		log.WithError(err).Errorf("failed to write secret to vault")
		return nil, errors.Wrap(err, "failed to write secret")
	}

	return secret, nil
}

func (v *VaultTransit) Read(
	ctx context.Context,
	key string,
) (*vault.Secret, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]any{
		"key": key,
	}

	log := v.log.WithField("key", key).WithContext(span.Context())
	secret, err := v.client.Logical().ReadWithContext(span.Context(), key)
	if err != nil {
		log.WithError(err).Errorf("failed to read secret from vault")
		return nil, errors.Wrap(err, "failed to read secret")
	}

	return secret, nil
}

func (v *VaultTransit) Delete(ctx context.Context, key string) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]any{
		"key": key,
	}

	log := v.log.WithField("key", key).WithContext(span.Context())

	log.Trace("deleting secret")

	_, err := v.client.Logical().DeleteWithContext(span.Context(), key)
	if err != nil {
		log.WithError(err).Errorf("failed to delete secret from vault")
		return errors.Wrap(err, "failed to delete secret")
	}

	return nil
}

// Decrypt implements KeyManagement.
func (v *VaultTransit) Decrypt(
	ctx context.Context,
	keyID *string,
	version *string,
	input string,
) (result string, _ error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	secret, err := v.Write(
		span.Context(),
		fmt.Sprintf("transit/decrypt/%s", v.config.KeyID),
		map[string]any{
			"ciphertext": input,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "vault failed to decrypt secret")
	}

	value, err := base64.StdEncoding.DecodeString(secret.Data["plaintext"].(string))
	if err != nil {
		return "", errors.Wrap(err, "failed to decode decrypted secret")
	}

	return string(value), nil
}

// Encrypt implements KeyManagement.
func (v *VaultTransit) Encrypt(
	ctx context.Context,
	input string,
) (keyID *string, version *string, result string, _ error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	secret, err := v.Write(
		span.Context(),
		fmt.Sprintf("transit/encrypt/%s", v.config.KeyID),
		map[string]any{
			"plaintext": base64.StdEncoding.EncodeToString([]byte(input)),
		},
	)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "vault failed to encrypt secret")
	}

	return &v.config.KeyID, nil, secret.Data["ciphertext"].(string), nil
}

// TODO Add a timeout to closing this, and test it
func (v *VaultTransit) Close() error {
	var err error
	if v.closer != nil {
		promise := make(chan error)
		v.closer <- promise

		err = <-promise
		if err != nil {
			v.log.WithError(err).Errorf("failed to close TLS worker")
		}
	}

	if v.tokenCloser != nil {
		promise := make(chan error)
		v.closer <- promise

		err = <-promise
		if err != nil {
			v.log.WithError(err).Errorf("failed to close token refresh worker")
		}
	}

	return err
}
