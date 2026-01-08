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
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/round"
	openbao "github.com/openbao/openbao/api/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type OpenBaoTransitConfig struct {
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

type OpenBaoTransit struct {
	tokenTTL        sync.Once
	tokenSync       sync.RWMutex
	tokenExpiration int64
	tokenCloser     chan chan error
	host            string
	config          OpenBaoTransitConfig
	log             *logrus.Entry
	client          *openbao.Client
	usingCustomTLS  bool
	tlsWatch        sync.Once
	lock            sync.RWMutex
	tls             *tls.Config
	closer          chan chan error
}

func NewOpenBaoTransit(
	ctx context.Context,
	config OpenBaoTransitConfig,
) (*OpenBaoTransit, error) {
	log := config.Log
	host, err := url.Parse(config.Address)
	if err != nil {
		log.WithField("url", config.Address).WithError(err).Errorf("failed to parse openbao URL")
		return nil, errors.Wrap(err, "failed to parse openbao URL")
	}

	helper := &OpenBaoTransit{
		host:           host.Hostname(),
		config:         config,
		log:            log,
		client:         nil,
		usingCustomTLS: len(config.TLSCAPath) > 0,
		tls:            nil,
		closer:         nil,
	}

	trip := round.NewObservabilityRoundTripper(
		&http.Transport{
			IdleConnTimeout: config.IdleConnTimeout,
		},
		func(
			ctx context.Context,
			request *http.Request,
			response *http.Response,
			err error,
		) {
			logEntry := log.WithContext(ctx).WithFields(logrus.Fields{
				"openbao": logrus.Fields{
					"method": request.Method,
					"url":    request.URL.String(),
				},
			})

			if err != nil {
				logEntry = logEntry.WithError(err)
			}
			logEntry.Trace("making request to openbao API")
		},
	)

	openbaoConfig := &openbao.Config{
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
		openbaoConfig.HttpClient.Transport = &http.Transport{
			DialTLSContext:  helper.dialTLS,
			IdleConnTimeout: config.IdleConnTimeout,
		}
	}

	helper.client, err = openbao.NewClient(openbaoConfig)
	if err != nil {
		log.WithError(err).Errorf("failed to create openbao client")
		return nil, errors.Wrap(err, "failed to create openbao client")
	}

	if err = helper.authenticate(ctx); err != nil {
		return nil, err
	}

	// Start the authentication worker.
	helper.authenticationWorker()

	return helper, nil
}

// dialTLS is a middleware function that is added to allow monetr to easily
// rotate the TLS certificates for the OpenBao server without downtime.
func (o *OpenBaoTransit) dialTLS(
	ctx context.Context,
	network, addr string,
) (net.Conn, error) {
	o.lock.RLock()
	defer o.lock.RUnlock()

	return tls.Dial(network, addr, o.tls)
}

// This function sets up a file system watcher to monitor changes in TLS
// certificate files. When changes are detected, it calls reloadTLS which
// atomically swaps the TLS config that is used to establish new connections to
// the OpenBao server.
//
//	watchCertificates()
//	├── Uses sync.Once to ensure the watcher setup runs only once
//	│   └── tlsWatch.Do()
//	│       └── go routine for asynchronous execution
//	│           ├── Initializes the logger and file watcherimport openbao "github.com/openbao/openbao/api/v2"
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
//	│               └── Handles closure of the watcher via o.closer channel
//
// This function is called when the client is initialized and runs until Close()
// is called on the client.
func (o *OpenBaoTransit) watchCertificates() {
	o.tlsWatch.Do(func() {
		go func() {
			log := o.log
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				log.WithError(err).Errorf("failed to create file system watcher, openbao TLS certificates cannot be auto rotated")
				return
			}

			o.closer = make(chan chan error, 1)

			paths := []string{
				o.config.TLSCertificatePath,
				o.config.TLSKeyPath,
				o.config.TLSCAPath,
			}
			for _, path := range paths {
				fileLog := log.WithField("file", path)
				fileLog.Trace("watching file for changes for openbao TLS")
				if err = watcher.Add(o.config.TLSCertificatePath); err != nil {
					fileLog.WithError(err).Errorf("failed to watch file for openbao TLS")
				}
			}

			for {
				select {
				case err = <-watcher.Errors:
					log.WithError(err).Warn("error watching file for openbao TLS")
				case event := <-watcher.Events:
					log.WithField("file", event.Name).Trace("observed changed in openbao TLS file")
					if err = o.reloadTLS(); err != nil {
						log.WithError(err).Errorf("failed to reload openbao TLS")
					}
				case promise := <-o.closer:
					log.Info("closing openbao helper TLS watcher")
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
//	├── Acquires lock on o.lock
//	│   └── Updates o.tls with new tls.Config
//	└── Returns nil if successful, otherwise an error
//
// This function is called when the OpenBao client is initialized and every time
// a certificate change is detected.
func (o *OpenBaoTransit) reloadTLS() error {
	log := o.log
	log.Debugf("reloading openbao TLS config")

	caCert, err := os.ReadFile(o.config.TLSCAPath)
	if err != nil {
		log.WithField("file", o.config.TLSCAPath).Errorf("failed to read CA for openbao TLS")
		return errors.Wrap(err, "failed to read CA for openbao TLS")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: o.config.InsecureSkipVerify,
		RootCAs:            caCertPool,
		ServerName:         o.host,
		Renegotiation:      tls.RenegotiateFreelyAsClient,
	}

	if o.config.TLSKeyPath != "" || o.config.TLSCertificatePath != "" {
		tlsKeyPair, err := tls.LoadX509KeyPair(
			o.config.TLSCertificatePath,
			o.config.TLSKeyPath,
		)
		if err != nil {
			log.WithFields(logrus.Fields{
				"certPath": o.config.TLSCertificatePath,
				"keyPath":  o.config.TLSKeyPath,
			}).WithError(err).Errorf("failed to load TLS key pair for openbao")
			return errors.Wrap(err, "failed to load TLS key pair for openbao")
		}

		tlsConfig.Certificates = []tls.Certificate{
			tlsKeyPair,
		}
	}

	o.lock.Lock()
	defer o.lock.Unlock()

	o.tls = tlsConfig

	return nil
}

// This function sets up a worker that periodically checks the expiration status
// of the OpenBao token and refreshes it if necessary.
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
//	│               └── Handles closure of the worker via o.tokenCloser channel
//
// This function is called once when the client is created and runs a background
// job until Close() is called.
func (o *OpenBaoTransit) authenticationWorker() {
	o.tokenTTL.Do(func() {
		log := o.log
		if atomic.LoadInt64(&o.tokenExpiration) == math.MaxInt64 {
			log.Info("openbao token will never expire, background token refresher will not be started")
			return
		}

		go func() {
			log.Debug("openbao token refresh worker has started, tokens will be refreshed before they expire")
			o.tokenCloser = make(chan chan error, 1)

			// Check to see if the token is going to expire every minute
			frequency := 30 * time.Second
			ticker := time.NewTicker(frequency)
			for {
				select {
				case <-ticker.C:
					// If the token will expire before we check it next, then refresh the token.
					if atomic.LoadInt64(&o.tokenExpiration) < time.Now().Add(frequency).Unix() {
						log.Debug("token will expire before the next check, refreshing token")
						err := func() error {
							ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
							defer cancel()
							return o.authenticate(ctx)
						}()
						if err != nil {
							log.WithError(err).Error("failed to refresh openbao token")
						}
					}
				case promise := <-o.tokenCloser:
					log.Info("stopping refresh token worker")
					promise <- nil
					return
				}
			}
		}()
	})
}

// This function handles the authentication process to OpenBao using different
// methods based on the configuration.
//
//	authenticate(ctx context.Context) error
//	├── Acquires lock on o.tokenSync
//	├── Switches on o.config.Auth for different authentication methods:
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
// currently used authentication for communication with the OpenBao server.
func (o *OpenBaoTransit) authenticate(ctx context.Context) error {
	o.tokenSync.Lock()
	defer o.tokenSync.Unlock()
	log := o.log.WithField("method", o.config.AuthMethod)

	var auth *openbao.SecretAuth
	switch o.config.AuthMethod {
	// TODO Implement app role authentication
	// https://openbao.org/docs/auth/approle/
	case "userpass":
		// https://openbao.org/docs/auth/userpass/
		log.Trace("authenticating to openbao")
		result, err := o.client.Logical().WriteWithContext(
			ctx,
			"auth/userpass/login/"+o.config.Username,
			map[string]any{
				"password": o.config.Password,
				"role":     o.config.Role,
			},
		)
		if err != nil {
			log.WithError(err).Errorf("failed to authenticate to openbao")
			return errors.Wrap(err, "failed to authenticate to openbao")
		}
		auth = result.Auth
	case "token":
		// https://openbao.org/docs/auth/token/
		o.client.SetToken(o.config.Token)
		_, err := o.client.Auth().Token().LookupSelfWithContext(ctx)
		if err != nil {
			log.WithError(err).Error("failed to authenticate to openbao")
			return errors.Wrap(err, "failed to authenticate to openbao")
		}
		auth = &openbao.SecretAuth{
			ClientToken:   o.config.Token,
			LeaseDuration: 0,
		}
	case "kubernetes":
		// https://openbao.org/docs/auth/kubernetes/
		log.Trace("authenticating to openbao")
		var token string
		tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"

		switch {
		case o.config.Token != "":
			log.Trace("using included token")
			token = o.config.Token
		case o.config.TokenFile != "":
			tokenPath = o.config.TokenFile
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

		log.Trace("authenticating to openbao")
		result, err := o.client.Logical().WriteWithContext(
			ctx,
			"auth/kubernetes/login",
			map[string]any{
				"role": o.config.Role,
				"jwt":  token,
			},
		)
		if err != nil {
			log.WithError(err).Error("failed to authenticate to openbao")
			return errors.Wrap(err, "failed to authenticate to openbao")
		}
		auth = result.Auth
	default:
		return errors.Errorf("%s authentication not implemented", o.config.AuthMethod)
	}

	if auth == nil {
		log.Fatalf("no authentication returned from openbao")
		return errors.Errorf("no authentication returned from openbao")
	}

	var nextExpiration int64
	if auth.LeaseDuration == 0 {
		// If the token does not expire, store a timestamp so far in the future that we won't ever re-auth
		nextExpiration = math.MaxInt64
	} else {
		// If the token does expire, store the expiration time (minus 1 minute) to have a safe buffer.
		nextExpiration = time.Now().Add(time.Duration(auth.LeaseDuration)*time.Second - 1*time.Minute).Unix()
		log.Debugf("openbao authentication will refresh by %s", time.Unix(nextExpiration, 0))
	}

	atomic.StoreInt64(&o.tokenExpiration, nextExpiration)

	o.client.SetToken(auth.ClientToken)
	log.Trace("successfully authenticated to openbao")

	return nil
}

func (o *OpenBaoTransit) Write(
	ctx context.Context,
	key string,
	value map[string]any,
) (*openbao.Secret, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]any{
		"key": key,
	}

	log := o.log.WithFields(logrus.Fields{
		"key":    key,
		"action": "write",
	}).WithContext(span.Context())
	log.Trace("handling secret with openbao")
	secret, err := o.client.Logical().WriteWithContext(
		span.Context(),
		key,
		value,
	)
	if err != nil {
		log.WithError(err).Errorf("failed to write secret to openbao")
		return nil, errors.Wrap(err, "failed to write secret")
	}

	return secret, nil
}

func (o *OpenBaoTransit) Read(
	ctx context.Context,
	key string,
) (*openbao.Secret, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]any{
		"key": key,
	}

	log := o.log.WithField("key", key).WithContext(span.Context())
	secret, err := o.client.Logical().ReadWithContext(span.Context(), key)
	if err != nil {
		log.WithError(err).Errorf("failed to read secret from openbao")
		return nil, errors.Wrap(err, "failed to read secret")
	}

	return secret, nil
}

func (o *OpenBaoTransit) Delete(ctx context.Context, key string) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]any{
		"key": key,
	}

	log := o.log.WithField("key", key).WithContext(span.Context())

	log.Trace("deleting secret")

	_, err := o.client.Logical().DeleteWithContext(span.Context(), key)
	if err != nil {
		log.WithError(err).Errorf("failed to delete secret from openbao")
		return errors.Wrap(err, "failed to delete secret")
	}

	return nil
}

// Decrypt implements KeyManagement.
func (o *OpenBaoTransit) Decrypt(
	ctx context.Context,
	keyID *string,
	version *string,
	input string,
) (result string, _ error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	secret, err := o.Write(
		span.Context(),
		fmt.Sprintf("transit/decrypt/%s", o.config.KeyID),
		map[string]any{
			"ciphertext": input,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "openbao failed to decrypt secret")
	}

	value, err := base64.StdEncoding.DecodeString(secret.Data["plaintext"].(string))
	if err != nil {
		return "", errors.Wrap(err, "failed to decode decrypted secret")
	}

	return string(value), nil
}

// Encrypt implements KeyManagement.
func (o *OpenBaoTransit) Encrypt(
	ctx context.Context,
	input string,
) (keyID *string, version *string, result string, _ error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	secret, err := o.Write(
		span.Context(),
		fmt.Sprintf("transit/encrypt/%s", o.config.KeyID),
		map[string]any{
			"plaintext": base64.StdEncoding.EncodeToString([]byte(input)),
		},
	)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "openbao failed to encrypt secret")
	}

	return &o.config.KeyID, nil, secret.Data["ciphertext"].(string), nil
}

// TODO Add a timeout to closing this, and test it
func (o *OpenBaoTransit) Close() error {
	var err error
	if o.closer != nil {
		promise := make(chan error)
		o.closer <- promise

		err = <-promise
		if err != nil {
			o.log.WithError(err).Errorf("failed to close TLS worker")
		}
	}

	if o.tokenCloser != nil {
		promise := make(chan error)
		o.closer <- promise

		err = <-promise
		if err != nil {
			o.log.WithError(err).Errorf("failed to close token refresh worker")
		}
	}

	return err
}
