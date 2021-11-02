package vault_helper

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Address            string
	Role               string
	Auth               string
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

type VaultHelper interface {
	WriteKV(ctx context.Context, key string, value map[string]interface{}) error
	ReadKV(ctx context.Context, key string) (*api.Secret, error)
	DeleteKV(ctx context.Context, key string) error
	Close() error
}

var (
	_ VaultHelper = &vaultBase{}
)

type vaultBase struct {
	tokenTTL        sync.Once
	tokenSync       sync.RWMutex
	tokenExpiration int64
	tokenCloser     chan chan error
	host            string
	config          Config
	log             *logrus.Entry
	client          *api.Client
	usingTLS        bool
	tlsWatch        sync.Once
	lock            sync.RWMutex
	tls             *tls.Config
	closer          chan chan error
}

func NewVaultHelper(log *logrus.Entry, config Config) (VaultHelper, error) {
	host, err := url.Parse(config.Address)
	if err != nil {
		log.WithField("url", config.Address).WithError(err).Errorf("failed to parse vault URL")
		return nil, errors.Wrap(err, "failed to parse vault URL")
	}

	helper := &vaultBase{
		host:     host.Host,
		config:   config,
		log:      log,
		client:   nil,
		usingTLS: len(config.TLSCAPath) > 0,
		tls:      nil,
		closer:   nil,
	}

	vaultConfig := &api.Config{
		Address: config.Address,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				IdleConnTimeout: config.IdleConnTimeout,
			},
			Timeout: config.Timeout,
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

	if helper.usingTLS {
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

	helper.client, err = api.NewClient(vaultConfig)
	if err != nil {
		log.WithError(err).Errorf("failed to create vault client")
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	if err = helper.authenticate(); err != nil {
		return nil, err
	}

	// Start the authentication worker.
	helper.authenticationWorker()

	return helper, nil
}

func (v *vaultBase) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	v.lock.RLock()
	defer v.lock.RUnlock()

	return tls.Dial(network, addr, v.tls)
}

func (v *vaultBase) watchCertificates() {
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

func (v *vaultBase) reloadTLS() error {
	log := v.log
	log.Debugf("reloading vault TLS config")

	caCert, err := ioutil.ReadFile(v.config.TLSCAPath)
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

func (v *vaultBase) authenticationWorker() {
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
			frequency := 1 * time.Minute
			ticker := time.NewTimer(frequency)
			for {
				select {
				case <-ticker.C:
					// If the token will expire before we check it next, then refresh the token.
					if atomic.LoadInt64(&v.tokenExpiration) < time.Now().Add(frequency).Unix() {
						log.Debug("token will expire before the next check, refreshing token")
						if err := v.authenticate(); err != nil {
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

func (v *vaultBase) authenticate() error {
	v.tokenSync.Lock()
	defer v.tokenSync.Unlock()
	log := v.log.WithField("method", v.config.Auth)

	var auth *api.SecretAuth
	switch v.config.Auth {
	case "userpass":
		log.Trace("authenticating to vault")
		result, err := v.client.Logical().Write("auth/userpass/login/"+v.config.Username, map[string]interface{}{
			"password": v.config.Password,
			"role":     v.config.Role,
		})
		if err != nil {
			log.WithError(err).Errorf("failed to authenticate to vault")
			return errors.Wrap(err, "failed to authenticate to vault")
		}
		auth = result.Auth
	case "token":
		auth = &api.SecretAuth{
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
			data, err := ioutil.ReadFile(tokenPath)
			if err != nil {
				fileLog.WithError(err).Errorf("failed to read token from specified file")
				return errors.Wrap(err, "failed to read token from specified file")
			}
			token = string(data)
		}

		log.Trace("authenticating to vault")
		result, err := v.client.Logical().Write("auth/kubernetes/login", map[string]interface{}{
			"role": v.config.Role,
			"jwt":  token,
		})
		if err != nil {
			log.WithError(err).Errorf("failed to authenticate to vault")
			return errors.Wrap(err, "failed to authenticate to vault")
		}
		auth = result.Auth
	default:
		return errors.Errorf("%s authentication not implemented", v.config.Auth)
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
		// If the token does expire, store the expiration time (minus 10 seconds) to have a safe buffer.
		nextExpiration = time.Now().Unix() + int64(auth.LeaseDuration) - 10
	}

	atomic.StoreInt64(&v.tokenExpiration, nextExpiration)

	v.client.SetToken(auth.ClientToken)
	log.Trace("successfully authenticated to vault")

	return nil
}

func (v *vaultBase) WriteKV(ctx context.Context, key string, value map[string]interface{}) error {
	span := sentry.StartSpan(ctx, "Vault - WriteKV")
	defer span.Finish()
	span.Data = map[string]interface{}{
		"key": key,
	}

	log := v.log.WithField("key", key).WithContext(span.Context())

	log.Trace("writing secret")
	_, err := v.client.Logical().Write(key, map[string]interface{}{
		"data": value,
	})
	if err != nil {
		log.WithError(err).Errorf("failed to write secret to vault")
		return errors.Wrap(err, "failed to write secret")
	}

	return nil
}

func (v *vaultBase) ReadKV(ctx context.Context, key string) (*api.Secret, error) {
	span := sentry.StartSpan(ctx, "Vault - ReadKV")
	defer span.Finish()
	span.Data = map[string]interface{}{
		"key": key,
	}

	log := v.log.WithField("key", key).WithContext(span.Context())

	log.Trace("reading secret")

	secret, err := v.client.Logical().Read(key)
	if err != nil {
		log.WithError(err).Errorf("failed to read secret from vault")
		return nil, errors.Wrap(err, "failed to read secret")
	}

	return secret, nil
}

func (v *vaultBase) DeleteKV(ctx context.Context, key string) error {
	span := sentry.StartSpan(ctx, "Vault - DeleteKV")
	defer span.Finish()
	span.Data = map[string]interface{}{
		"key": key,
	}

	log := v.log.WithField("key", key).WithContext(span.Context())

	log.Trace("deleting secret")

	_, err := v.client.Logical().Delete(key)
	if err != nil {
		log.WithError(err).Errorf("failed to delete secret from vault")
		return errors.Wrap(err, "failed to delete secret")
	}

	return nil
}

func (v *vaultBase) Close() error {
	var err error
	if v.closer != nil {
		promise := make(chan error, 0)
		v.closer <- promise

		err = <-promise
		if err != nil {
			v.log.WithError(err).Errorf("failed to close TLS worker")
		}
	}

	if v.tokenCloser != nil {
		promise := make(chan error, 0)
		v.closer <- promise

		err = <-promise
		if err != nil {
			v.log.WithError(err).Errorf("failed to close token refresh worker")
		}
	}

	return err
}
