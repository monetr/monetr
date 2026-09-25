package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils/testlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testAuthority struct {
	certificate *x509.Certificate
	key         *ecdsa.PrivateKey
	pem         []byte
}

func newTestAuthority(t *testing.T) *testAuthority {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "must generate ca key")

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "monetr test ca"},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(1 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err, "must create ca certificate")
	certificate, err := x509.ParseCertificate(der)
	require.NoError(t, err, "must parse ca certificate")

	return &testAuthority{
		certificate: certificate,
		key:         key,
		pem:         pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
	}
}

// issue returns a certificate and key signed by the authority as PEM, as well
// as the parsed pair.
func (a *testAuthority) issue(
	t *testing.T,
	name string,
) (certPEM, keyPEM []byte, pair tls.Certificate) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "must generate key")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: name},
		DNSNames:     []string{name},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(1 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, a.certificate, &key.PublicKey, a.key)
	require.NoError(t, err, "must create certificate")

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err, "must marshal key")

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	pair, err = tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err, "must parse key pair")

	return certPEM, keyPEM, pair
}

func writeFile(t *testing.T, path string, data []byte) {
	require.NoError(t, os.WriteFile(path, data, 0600), "must write %s", path)
}

// handshake performs a TLS handshake between the provided client config and a
// server using the provided server certificate. If clientCAs is provided then
// the server requires a client certificate signed by it. Returns the client
// certificate the server saw, if any.
func handshake(
	t *testing.T,
	client *tls.Config,
	server tls.Certificate,
	clientCAs *x509.CertPool,
) (*x509.Certificate, error) {
	// Use a real loopback connection, net.Pipe is unbuffered and will deadlock
	// when both sides try to send an alert after a failed handshake.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "must listen")
	defer listener.Close()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err, "must dial")
	defer clientConn.Close()
	serverConn, err := listener.Accept()
	require.NoError(t, err, "must accept")
	defer serverConn.Close()

	deadline := time.Now().Add(5 * time.Second)
	clientConn.SetDeadline(deadline)
	serverConn.SetDeadline(deadline)

	serverConfig := &tls.Config{
		Certificates: []tls.Certificate{server},
	}
	if clientCAs != nil {
		serverConfig.ClientAuth = tls.RequireAndVerifyClientCert
		serverConfig.ClientCAs = clientCAs
	}

	type result struct {
		peer *x509.Certificate
		err  error
	}
	serverResult := make(chan result, 1)
	go func() {
		conn := tls.Server(serverConn, serverConfig)
		err := conn.Handshake()
		var peer *x509.Certificate
		if certificates := conn.ConnectionState().PeerCertificates; len(certificates) > 0 {
			peer = certificates[0]
		}
		// Unblock the client if the server fails first.
		serverConn.Close()
		serverResult <- result{peer: peer, err: err}
	}()

	clientErr := tls.Client(clientConn, client).Handshake()
	clientConn.Close()
	serverSide := <-serverResult
	if clientErr != nil {
		return nil, clientErr
	}

	return serverSide.peer, serverSide.err
}

// serverHandshake performs a TLS handshake between a client trusting the
// provided roots and a server using the provided config. Returns the server
// certificate the client saw.
func serverHandshake(
	t *testing.T,
	server *tls.Config,
	roots *x509.CertPool,
	serverName string,
) (*x509.Certificate, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "must listen")
	defer listener.Close()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err, "must dial")
	defer clientConn.Close()
	serverConn, err := listener.Accept()
	require.NoError(t, err, "must accept")
	defer serverConn.Close()

	deadline := time.Now().Add(5 * time.Second)
	clientConn.SetDeadline(deadline)
	serverConn.SetDeadline(deadline)

	serverResult := make(chan error, 1)
	go func() {
		err := tls.Server(serverConn, server).Handshake()
		// Unblock the client if the server fails first.
		serverConn.Close()
		serverResult <- err
	}()

	conn := tls.Client(clientConn, &tls.Config{
		RootCAs:    roots,
		ServerName: serverName,
	})
	clientErr := conn.Handshake()
	var peer *x509.Certificate
	if certificates := conn.ConnectionState().PeerCertificates; len(certificates) > 0 {
		peer = certificates[0]
	}
	clientConn.Close()
	serverErr := <-serverResult
	if clientErr != nil {
		return nil, clientErr
	}

	return peer, serverErr
}

func TestFileSource(t *testing.T) {
	t.Run("verifies server against ca", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		_, _, server := ca.issue(t, "postgres.monetr.local")
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			ServerName:        "postgres.monetr.local",
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		_, err = handshake(t, source.ClientConfig(), server, nil)
		assert.NoError(t, err, "handshake should succeed")
	})

	t.Run("rejects server from another ca", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		other := newTestAuthority(t)
		_, _, server := other.issue(t, "postgres.monetr.local")
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			ServerName:        "postgres.monetr.local",
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		_, err = handshake(t, source.ClientConfig(), server, nil)
		assert.ErrorContains(t, err, "failed to verify server certificate")
	})

	t.Run("rejects server with wrong hostname", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		_, _, server := ca.issue(t, "evil.monetr.local")
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			ServerName:        "postgres.monetr.local",
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		_, err = handshake(t, source.ClientConfig(), server, nil)
		assert.ErrorContains(t, err, "failed to verify server certificate")
	})

	t.Run("insecure skip verify", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		other := newTestAuthority(t)
		_, _, server := other.issue(t, "postgres.monetr.local")
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath:  filepath.Join(directory, "ca.crt"),
			ServerName:         "postgres.monetr.local",
			InsecureSkipVerify: true,
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		_, err = handshake(t, source.ClientConfig(), server, nil)
		assert.NoError(t, err, "handshake should succeed when skipping verification")
	})

	t.Run("presents client certificate", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		_, _, server := ca.issue(t, "postgres.monetr.local")
		clientCert, clientKey, _ := ca.issue(t, "monetr")
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)
		writeFile(t, filepath.Join(directory, "tls.crt"), clientCert)
		writeFile(t, filepath.Join(directory, "tls.key"), clientKey)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			CertificatePath:   filepath.Join(directory, "tls.crt"),
			KeyPath:           filepath.Join(directory, "tls.key"),
			ServerName:        "postgres.monetr.local",
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		clientCAs := x509.NewCertPool()
		clientCAs.AddCert(ca.certificate)
		peer, err := handshake(t, source.ClientConfig(), server, clientCAs)
		require.NoError(t, err, "handshake should succeed")
		require.NotNil(t, peer, "server should have seen a client certificate")
		assert.Equal(t, "monetr", peer.Subject.CommonName)
	})

	t.Run("missing files", func(t *testing.T) {
		directory := t.TempDir()
		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			ServerName:        "postgres.monetr.local",
		})
		assert.ErrorContains(t, err, "failed to load ca certificate")
		assert.Nil(t, source)
	})

	t.Run("certificate without key", func(t *testing.T) {
		source, err := NewFileSource(testlog.GetLog(t), Options{
			CertificatePath: "/tmp/tls.crt",
		})
		assert.ErrorContains(t, err, "certificate path and key path must both be provided")
		assert.Nil(t, source)
	})

	t.Run("reload swaps certificates on the same config", func(t *testing.T) {
		directory := t.TempDir()
		oldCA := newTestAuthority(t)
		newCA := newTestAuthority(t)
		_, _, server := newCA.issue(t, "postgres.monetr.local")
		writeFile(t, filepath.Join(directory, "ca.crt"), oldCA.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			ServerName:        "postgres.monetr.local",
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		config := source.ClientConfig()
		_, err = handshake(t, config, server, nil)
		assert.Error(t, err, "should not trust the new ca yet")

		writeFile(t, filepath.Join(directory, "ca.crt"), newCA.pem)
		require.NoError(t, source.Reload(), "must reload")

		_, err = handshake(t, config, server, nil)
		assert.NoError(t, err, "existing config should trust the new ca after reload")
	})

	t.Run("failed reload keeps previous certificates", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		_, _, server := ca.issue(t, "postgres.monetr.local")
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			ServerName:        "postgres.monetr.local",
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		writeFile(t, filepath.Join(directory, "ca.crt"), []byte("not a certificate"))
		assert.ErrorContains(t, source.Reload(), "failed to parse ca certificate")

		_, err = handshake(t, source.ClientConfig(), server, nil)
		assert.NoError(t, err, "should still trust the previous ca")
	})

	t.Run("watches for changes", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		_, _, server := ca.issue(t, "postgres.monetr.local")
		oldCert, oldKey, _ := ca.issue(t, "old")
		newCert, newKey, _ := ca.issue(t, "new")
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)
		writeFile(t, filepath.Join(directory, "tls.crt"), oldCert)
		writeFile(t, filepath.Join(directory, "tls.key"), oldKey)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			CertificatePath:   filepath.Join(directory, "tls.crt"),
			KeyPath:           filepath.Join(directory, "tls.key"),
			ServerName:        "postgres.monetr.local",
		})
		require.NoError(t, err, "must create source")
		source.(*fileSource).debounce = 100 * time.Millisecond
		require.NoError(t, source.Start(), "must start")

		clientCAs := x509.NewCertPool()
		clientCAs.AddCert(ca.certificate)
		config := source.ClientConfig()

		peer, err := handshake(t, config, server, clientCAs)
		require.NoError(t, err, "handshake should succeed")
		assert.Equal(t, "old", peer.Subject.CommonName)

		// Write the certificate and key separately, the reload should wait for
		// both instead of failing on a mismatched pair.
		writeFile(t, filepath.Join(directory, "tls.crt"), newCert)
		time.Sleep(20 * time.Millisecond)
		writeFile(t, filepath.Join(directory, "tls.key"), newKey)

		assert.Eventually(t, func() bool {
			peer, err := handshake(t, config, server, clientCAs)
			return err == nil && peer.Subject.CommonName == "new"
		}, 5*time.Second, 50*time.Millisecond, "should present the new client certificate")

		assert.NoError(t, source.Stop(), "must stop")
	})

	t.Run("polls for changes", func(t *testing.T) {
		directory := t.TempDir()
		oldCA := newTestAuthority(t)
		newCA := newTestAuthority(t)
		_, _, server := newCA.issue(t, "postgres.monetr.local")
		writeFile(t, filepath.Join(directory, "ca.crt"), oldCA.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			ServerName:        "postgres.monetr.local",
			PollInterval:      50 * time.Millisecond,
		})
		require.NoError(t, err, "must create source")
		// Push the debounce out so only polling could pick up the change.
		source.(*fileSource).debounce = time.Hour
		require.NoError(t, source.Start(), "must start")

		writeFile(t, filepath.Join(directory, "ca.crt"), newCA.pem)

		config := source.ClientConfig()
		assert.Eventually(t, func() bool {
			_, err := handshake(t, config, server, nil)
			return err == nil
		}, 5*time.Second, 50*time.Millisecond, "should trust the new ca")

		assert.NoError(t, source.Stop(), "must stop")
	})

	t.Run("serves certificate", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		serverCert, serverKey, _ := ca.issue(t, "my.monetr.local")
		writeFile(t, filepath.Join(directory, "tls.crt"), serverCert)
		writeFile(t, filepath.Join(directory, "tls.key"), serverKey)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CertificatePath: filepath.Join(directory, "tls.crt"),
			KeyPath:         filepath.Join(directory, "tls.key"),
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		roots := x509.NewCertPool()
		roots.AddCert(ca.certificate)
		peer, err := serverHandshake(t, source.ServerConfig(), roots, "my.monetr.local")
		require.NoError(t, err, "handshake should succeed")
		assert.Equal(t, "my.monetr.local", peer.Subject.CommonName)
	})

	t.Run("server without certificate", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
		})
		require.NoError(t, err, "must create source")
		defer source.Stop()

		_, err = serverHandshake(t, source.ServerConfig(), nil, "my.monetr.local")
		assert.Error(t, err, "handshake should fail without a server certificate")
	})

	t.Run("serves rotated certificate", func(t *testing.T) {
		directory := t.TempDir()
		oldCA := newTestAuthority(t)
		newCA := newTestAuthority(t)
		oldCert, oldKey, _ := oldCA.issue(t, "my.monetr.local")
		newCert, newKey, _ := newCA.issue(t, "my.monetr.local")
		writeFile(t, filepath.Join(directory, "tls.crt"), oldCert)
		writeFile(t, filepath.Join(directory, "tls.key"), oldKey)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CertificatePath: filepath.Join(directory, "tls.crt"),
			KeyPath:         filepath.Join(directory, "tls.key"),
		})
		require.NoError(t, err, "must create source")
		source.(*fileSource).debounce = 100 * time.Millisecond
		require.NoError(t, source.Start(), "must start")

		oldRoots := x509.NewCertPool()
		oldRoots.AddCert(oldCA.certificate)
		newRoots := x509.NewCertPool()
		newRoots.AddCert(newCA.certificate)
		config := source.ServerConfig()

		_, err = serverHandshake(t, config, oldRoots, "my.monetr.local")
		require.NoError(t, err, "handshake should succeed with the old certificate")

		// Write the certificate and key separately, the reload should wait for
		// both instead of failing on a mismatched pair.
		writeFile(t, filepath.Join(directory, "tls.crt"), newCert)
		time.Sleep(20 * time.Millisecond)
		writeFile(t, filepath.Join(directory, "tls.key"), newKey)

		// The certificates are issued by different authorities so the client can
		// only verify the new certificate once it is being served.
		assert.Eventually(t, func() bool {
			_, err := serverHandshake(t, config, newRoots, "my.monetr.local")
			return err == nil
		}, 5*time.Second, 50*time.Millisecond, "should serve the new certificate")

		assert.NoError(t, source.Stop(), "must stop")
	})

	t.Run("stop without start", func(t *testing.T) {
		directory := t.TempDir()
		ca := newTestAuthority(t)
		writeFile(t, filepath.Join(directory, "ca.crt"), ca.pem)

		source, err := NewFileSource(testlog.GetLog(t), Options{
			CACertificatePath: filepath.Join(directory, "ca.crt"),
			ServerName:        "postgres.monetr.local",
		})
		require.NoError(t, err, "must create source")
		assert.NoError(t, source.Stop(), "should not block when never started")
	})
}
