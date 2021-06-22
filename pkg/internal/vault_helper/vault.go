package vault_helper

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"github.com/fsnotify/fsnotify"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Config struct {
	Address            string
	Role               string
	Auth               string
	Token              string
	TokenFile          string
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
	Close() error
}

var (
	_ VaultHelper = &vaultBase{}
)

type vaultBase struct {
	host     string
	config   Config
	log      *logrus.Entry
	client   *api.Client
	usingTLS bool
	tlsWatch sync.Once
	lock     sync.RWMutex
	tls      *tls.Config
	closer   chan chan error
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

func (v *vaultBase) authenticate() error {
	log := v.log.WithField("method", v.config.Auth)

	switch v.config.Auth {
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

		if result.Auth == nil {
			log.WithError(err).Fatalf("no authentication returned from vault")
			return errors.Errorf("no authentication returned from vault")
		}

		v.client.SetToken(result.Auth.ClientToken)
		log.Trace("successfully authenticated to vault")
	default:
		return errors.Errorf("%s authentication not implemented", v.config.Auth)
	}

	return nil
}

func (v *vaultBase) WriteKV(ctx context.Context, key string, value map[string]interface{}) error {
	span := sentry.StartSpan(ctx, "Vault - WriteKV")
	defer span.Finish()

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

	panic("implement me")
}

func (v *vaultBase) Close() error {
	if v.closer != nil {
		promise := make(chan error, 0)
		v.closer <- promise

		return <-promise
	}

	return nil
}
