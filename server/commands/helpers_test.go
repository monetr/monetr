package commands

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCertificatesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not meaningful on Windows")
	}

	tempDir, err := os.MkdirTemp("", "monetr-certs")
	require.NoError(t, err, "must create temp directory")
	log := testutils.GetLog(t)

	// Use a non-existent subdirectory so MkdirAll has to create it. This
	// exercises both the directory-creation perm and the file-write perm.
	keyDir := path.Join(tempDir, "security")
	keyPath := path.Join(keyDir, "ed25519.pem")

	configuration := config.Configuration{
		Security: config.Security{
			PrivateKey: keyPath,
		},
	}

	publicKey, privateKey, err := loadCertificates(configuration, log, true)
	require.NoError(t, err, "must generate and persist key")
	require.NotNil(t, publicKey, "public key must be returned")
	require.NotNil(t, privateKey, "private key must be returned")

	dirInfo, err := os.Stat(keyDir)
	require.NoError(t, err, "must stat key directory")
	require.True(t, dirInfo.IsDir(), "key directory must be a directory")
	assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm(), "key directory must be drwx------")

	keyInfo, err := os.Stat(keyPath)
	require.NoError(t, err, "must stat private key file")
	assert.True(t, keyInfo.Mode().IsRegular(), "private key must be a regular file")
	assert.Equal(t, os.FileMode(0600), keyInfo.Mode().Perm(), "private key file must be -rw-------")

	// Round-trip: re-load from disk and verify the same key comes back. This
	// confirms the tightened mode does not block the server's own access path.
	loadedPublic, loadedPrivate, err := loadCertificates(configuration, log, false)
	require.NoError(t, err, "must reload existing key from disk")
	assert.True(t, loadedPublic.Equal(publicKey), "public key from disk must match generated key")
	assert.True(t, loadedPrivate.Equal(privateKey), "private key from disk must match generated key")
}
