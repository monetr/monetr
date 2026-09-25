package certs

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// defaultDebounce is how long the watcher waits after the last filesystem event
// before reloading. Certificate rotation usually writes the certificate and the
// key as separate files, reloading on the first event would see a mismatched
// pair.
const defaultDebounce = 2 * time.Second

type Options struct {
	// CACertificatePath is the certificate authority used to verify the remote
	// server. If this is left blank then the system certificate pool is used.
	CACertificatePath string
	// CertificatePath and KeyPath are the client certificate and key presented to
	// the remote server. Both must be provided or neither.
	CertificatePath string
	KeyPath         string
	// ServerName is the hostname the remote server's certificate is verified
	// against.
	ServerName         string
	InsecureSkipVerify bool
	// PollInterval will re-read the certificates on an interval in addition to
	// watching the filesystem. Useful when filesystem events are not reliable,
	// like on some network mounts. Zero disables polling.
	PollInterval time.Duration
}

// Source holds the current certificates loaded from disk and keeps them up to
// date as the files change. The tls.Config it provides is never swapped out,
// instead the certificates are surfaced through callbacks on the config so
// every new connection picks up whatever is currently loaded.
type Source interface {
	// ClientConfig returns a TLS config for clients connecting to a server using
	// these certificates. The same config can be shared by multiple clients.
	ClientConfig() *tls.Config
	// Reload reads the certificates from disk. If the certificates fail to load
	// then the previously loaded certificates are kept.
	Reload() error
	Start() error
	Stop() error
}

var (
	_ Source = &fileSource{}
)

type material struct {
	checksum    []byte
	roots       *x509.CertPool
	certificate *tls.Certificate
}

type fileSource struct {
	log           *slog.Logger
	options       Options
	debounce      time.Duration
	once          sync.Once
	started       atomic.Bool
	cancelChannel chan chan error
	watcher       *fsnotify.Watcher
	material      atomic.Pointer[material]
}

func NewFileSource(log *slog.Logger, options Options) (Source, error) {
	if (options.CertificatePath == "") != (options.KeyPath == "") {
		return nil, errors.New("certificate path and key path must both be provided")
	}

	source := &fileSource{
		log:           log,
		options:       options,
		debounce:      defaultDebounce,
		cancelChannel: make(chan chan error),
	}

	// Load the certificates up front so bad paths are caught at startup rather
	// than on the first connection.
	if err := source.Reload(); err != nil {
		return nil, err
	}

	paths := source.directories()
	if len(paths) == 0 {
		return source, nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate watcher")
	}

	// Watch the directories instead of the files themselves. Kubernetes rotates
	// mounted secrets by swapping a symlink in the directory, which does not
	// produce events on the files.
	for _, path := range paths {
		if err = watcher.Add(path); err != nil {
			watcher.Close()
			return nil, errors.Wrap(err, "failed to add certificate path to watcher")
		}
	}
	source.watcher = watcher

	return source, nil
}

func (f *fileSource) ClientConfig() *tls.Config {
	manuallyVerifiedConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: f.options.ServerName,
		// Go does not provide a way to change RootCAs on a config that is already
		// in use. So the built in verification is disabled and the peer is verified
		// in VerifyConnection against the currently loaded roots.
		InsecureSkipVerify:   true,
		GetClientCertificate: f.getClientCertificate,
		VerifyConnection:     f.verifyConnection,
	}

	return manuallyVerifiedConfig
}

func (f *fileSource) getClientCertificate(
	_ *tls.CertificateRequestInfo,
) (*tls.Certificate, error) {
	if certificate := f.material.Load().certificate; certificate != nil {
		return certificate, nil
	}

	// No client certificate is configured, an empty certificate tells the server
	// we don't have one.
	return &tls.Certificate{}, nil
}

func (f *fileSource) verifyConnection(state tls.ConnectionState) error {
	if f.options.InsecureSkipVerify {
		return nil
	}

	if len(state.PeerCertificates) == 0 {
		return errors.New("server did not present a certificate")
	}

	serverName := state.ServerName
	if serverName == "" {
		serverName = f.options.ServerName
	}
	if serverName == "" {
		return errors.New("server name is required to verify the server certificate")
	}

	intermediates := x509.NewCertPool()
	for _, certificate := range state.PeerCertificates[1:] {
		intermediates.AddCert(certificate)
	}

	_, err := state.PeerCertificates[0].Verify(x509.VerifyOptions{
		// If roots is nil then the system pool is used.
		Roots:         f.material.Load().roots,
		Intermediates: intermediates,
		DNSName:       serverName,
	})
	return errors.Wrap(err, "failed to verify server certificate")
}

func (f *fileSource) Reload() error {
	var checksum []byte
	next := &material{}

	if f.options.CACertificatePath != "" {
		caCert, err := os.ReadFile(f.options.CACertificatePath)
		if err != nil {
			return errors.Wrap(err, "failed to load ca certificate")
		}

		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(caCert) {
			return errors.New("failed to parse ca certificate")
		}
		next.roots = roots
		checksum = append(checksum, caCert...)
	}

	if f.options.KeyPath != "" {
		certificate, err := os.ReadFile(f.options.CertificatePath)
		if err != nil {
			return errors.Wrap(err, "failed to load client certificate")
		}

		key, err := os.ReadFile(f.options.KeyPath)
		if err != nil {
			return errors.Wrap(err, "failed to load client key")
		}

		pair, err := tls.X509KeyPair(certificate, key)
		if err != nil {
			return errors.Wrap(err, "failed to parse client certificate")
		}
		next.certificate = &pair
		checksum = append(checksum, certificate...)
		checksum = append(checksum, key...)
	}
	next.checksum = checksum

	// Only log when something actually changed, polling will reload the same
	// files over and over.
	if current := f.material.Load(); current != nil {
		if bytes.Equal(current.checksum, next.checksum) {
			return nil
		}
		f.log.Info("reloaded TLS certificates")
	}

	f.material.Store(next)

	return nil
}

func (f *fileSource) Start() error {
	f.once.Do(func() {
		f.started.Store(true)
		go f.backgroundWorker()
	})

	return nil
}

func (f *fileSource) backgroundWorker() {
	var cancelChannel chan error
	var err error
	defer func() {
		if cancelChannel != nil {
			cancelChannel <- err
		}
	}()

	// A nil channel blocks forever, so any of these that are not configured are
	// just never selected.
	var events chan fsnotify.Event
	var errs chan error
	if f.watcher != nil {
		events = f.watcher.Events
		errs = f.watcher.Errors
	}

	var poll <-chan time.Time
	if f.options.PollInterval > 0 {
		ticker := time.NewTicker(f.options.PollInterval)
		defer ticker.Stop()
		poll = ticker.C
	}

	debounce := time.NewTimer(f.debounce)
	debounce.Stop()
	defer debounce.Stop()

	for {
		select {
		case watchErr := <-errs:
			f.log.Warn(
				"error watching TLS certificates",
				"err", watchErr,
			)
		case <-events:
			// Wait for events to settle before reloading, each new event pushes the
			// reload back.
			debounce.Reset(f.debounce)
		case <-debounce.C:
			f.reload()
		case <-poll:
			f.reload()
		case cancelChannel = <-f.cancelChannel:
			if f.watcher != nil {
				err = f.watcher.Close()
			}
			return
		}
	}
}

func (f *fileSource) reload() {
	if err := f.Reload(); err != nil {
		f.log.Warn(
			"failed to reload TLS certificates, keeping previous certificates",
			"err", err,
		)
	}
}

func (f *fileSource) Stop() error {
	if !f.started.Load() {
		if f.watcher != nil {
			return f.watcher.Close()
		}
		return nil
	}

	callback := make(chan error)
	f.cancelChannel <- callback

	return <-callback
}

func (f *fileSource) directories() []string {
	paths := make([]string, 0, 3)
	for _, path := range []string{
		f.options.CACertificatePath,
		f.options.CertificatePath,
		f.options.KeyPath,
	} {
		if path == "" {
			continue
		}

		directory := filepath.Dir(path)
		if !slices.Contains(paths, directory) {
			paths = append(paths, directory)
		}
	}

	return paths
}
