package config_test

import (
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/stretchr/testify/assert"
)

func TestSentry_GetExternalOrigin(t *testing.T) {
	t.Run("disabled sentry is skipped", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     false,
			ExternalDSN: "https://abc123@o12345.ingest.sentry.io/67890",
		}
		assert.Equal(t, "", sentry.GetExternalOrigin())
	})

	t.Run("empty external dsn is skipped", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "",
		}
		assert.Equal(t, "", sentry.GetExternalOrigin())
	})

	t.Run("unparseable dsn is skipped", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "https://abc@o1.ingest.sentry.io/%zz",
		}
		assert.Equal(t, "", sentry.GetExternalOrigin())
	})

	t.Run("non http scheme is skipped", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "ftp://abc@files.example.com/1",
		}
		assert.Equal(t, "", sentry.GetExternalOrigin())
	})

	t.Run("missing host is skipped", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "https://abc@/1",
		}
		assert.Equal(t, "", sentry.GetExternalOrigin())
	})

	t.Run("localhost is skipped", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "http://abc@localhost:9000/1",
		}
		assert.Equal(t, "", sentry.GetExternalOrigin())
	})

	t.Run("loopback ipv4 is skipped", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "http://abc@127.0.0.1/1",
		}
		assert.Equal(t, "", sentry.GetExternalOrigin())
	})

	t.Run("hosted sentry dsn is returned without auth or path", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "https://abc123@o12345.ingest.sentry.io/67890",
		}
		assert.Equal(t, "https://o12345.ingest.sentry.io", sentry.GetExternalOrigin())
	})

	t.Run("self hosted dsn with port is preserved", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "https://abc@sentry.internal.example.com:8443/1",
		}
		assert.Equal(t, "https://sentry.internal.example.com:8443", sentry.GetExternalOrigin())
	})

	t.Run("plain http public host is returned", func(t *testing.T) {
		sentry := config.Sentry{
			Enabled:     true,
			ExternalDSN: "http://abc@sentry.example.com/1",
		}
		assert.Equal(t, "http://sentry.example.com", sentry.GetExternalOrigin())
	})
}
