package commands

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckStatus(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/health", r.URL.Path, "should request the health endpoint")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"apiHealthy":true,"dbHealthy":true}`))
		}))
		defer server.Close()

		requestUrl, err := url.Parse(server.URL)
		require.NoError(t, err, "must parse server url")

		err = checkStatus(
			t.Context(),
			testutils.GetLog(t),
			server.Client(),
			requestUrl.JoinPath("/api/health"),
		)
		assert.NoError(t, err, "should be healthy")
	})

	t.Run("healthy over tls", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		requestUrl, err := url.Parse(server.URL)
		require.NoError(t, err, "must parse server url")

		err = checkStatus(
			t.Context(),
			testutils.GetLog(t),
			server.Client(),
			requestUrl.JoinPath("/api/health"),
		)
		assert.NoError(t, err, "should be healthy")
	})

	t.Run("unhealthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"apiHealthy":true,"dbHealthy":false}`))
		}))
		defer server.Close()

		requestUrl, err := url.Parse(server.URL)
		require.NoError(t, err, "must parse server url")

		err = checkStatus(
			t.Context(),
			testutils.GetLog(t),
			server.Client(),
			requestUrl.JoinPath("/api/health"),
		)
		assert.EqualError(t, err, "monetr is not healthy, status code: 500")
	})

	t.Run("untrusted certificate", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		requestUrl, err := url.Parse(server.URL)
		require.NoError(t, err, "must parse server url")

		// The default client does not trust the test server's certificate.
		err = checkStatus(
			t.Context(),
			testutils.GetLog(t),
			&http.Client{},
			requestUrl.JoinPath("/api/health"),
		)
		assert.ErrorContains(t, err, "failed to make status request")
	})

	t.Run("server not running", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		requestUrl, err := url.Parse(server.URL)
		require.NoError(t, err, "must parse server url")
		server.Close()

		err = checkStatus(
			t.Context(),
			testutils.GetLog(t),
			&http.Client{},
			requestUrl.JoinPath("/api/health"),
		)
		assert.ErrorContains(t, err, "failed to make status request")
	})
}
