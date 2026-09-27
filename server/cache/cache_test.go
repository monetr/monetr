package cache_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issueTestCertificate creates a certificate for localhost signed by the
// provided authority, or self signed if no authority is provided. Returns the
// certificate and key as PEM.
func issueTestCertificate(
	t *testing.T,
	authority *x509.Certificate,
	authorityKey *ecdsa.PrivateKey,
) (certificate *x509.Certificate, key *ecdsa.PrivateKey, certPEM, keyPEM []byte) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "must generate key")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "localhost"},
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(1 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}
	if authority == nil {
		template.IsCA = true
		template.BasicConstraintsValid = true
		template.KeyUsage |= x509.KeyUsageCertSign
		authority, authorityKey = template, key
	}

	der, err := x509.CreateCertificate(rand.Reader, template, authority, &key.PublicKey, authorityKey)
	require.NoError(t, err, "must create certificate")
	certificate, err = x509.ParseCertificate(der)
	require.NoError(t, err, "must parse certificate")
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err, "must marshal key")

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certificate, key, certPEM, keyPEM
}

func TestNewRedisCache(t *testing.T) {
	t.Run("mutual tls", func(t *testing.T) {
		authority, authorityKey, authorityPEM, _ := issueTestCertificate(t, nil, nil)
		_, _, serverCertPEM, serverKeyPEM := issueTestCertificate(t, authority, authorityKey)
		_, _, clientCertPEM, clientKeyPEM := issueTestCertificate(t, authority, authorityKey)

		serverPair, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
		require.NoError(t, err, "must parse server key pair")
		authorities := x509.NewCertPool()
		authorities.AddCert(authority)

		// Require the client to present a certificate signed by our authority so
		// we know that monetr is actually using the client certificate too.
		mini, err := miniredis.RunTLS(&tls.Config{
			Certificates: []tls.Certificate{serverPair},
			ClientCAs:    authorities,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		})
		require.NoError(t, err, "must start miniredis with tls")
		t.Cleanup(mini.Close)

		directory := t.TempDir()
		caPath := filepath.Join(directory, "ca.cert")
		certPath := filepath.Join(directory, "tls.cert")
		keyPath := filepath.Join(directory, "tls.key")
		require.NoError(t, os.WriteFile(caPath, authorityPEM, 0600), "must write ca")
		require.NoError(t, os.WriteFile(certPath, clientCertPEM, 0600), "must write client cert")
		require.NoError(t, os.WriteFile(keyPath, clientKeyPEM, 0600), "must write client key")

		controller, err := cache.NewRedisCache(testutils.GetLog(t), config.Redis{
			Enabled:           true,
			Address:           "localhost",
			Port:              mini.Server().Addr().Port,
			TLS:               true,
			CACertificatePath: caPath,
			CertificatePath:   certPath,
			KeyPath:           keyPath,
		})
		require.NoError(t, err, "must create redis cache over tls")
		t.Cleanup(func() {
			assert.NoError(t, controller.Close(), "must close redis cache")
		})

		connection := controller.Pool().Get()
		defer connection.Close()
		_, err = connection.Do("SET", "key", "value")
		require.NoError(t, err, "must set key over tls")
		value, err := mini.Get("key")
		require.NoError(t, err, "must read key from miniredis")
		assert.Equal(t, "value", value, "value written over tls must be in miniredis")
	})
}
