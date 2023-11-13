package storage

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test is here to make sure that the url.Parse function behaves how I expect it to behave as part of the storage
// interface.
func TestParseURI(t *testing.T) {
	t.Run("s3 uri", func(t *testing.T) {
		uri := "s3://bucket/folder/file.txt"
		url, err := url.Parse(uri)
		assert.NoError(t, err, "must be able to parse the uri")
		assert.NotNil(t, url)
		assert.Equal(t, "s3", url.Scheme)
		assert.Equal(t, "bucket", url.Host)
		assert.Equal(t, "/folder/file.txt", url.Path)
	})

	t.Run("gcs uri", func(t *testing.T) {
		uri := "gcs://bucket/folder/file.txt"
		url, err := url.Parse(uri)
		assert.NoError(t, err, "must be able to parse the uri")
		assert.NotNil(t, url)
		assert.Equal(t, "gcs", url.Scheme)
		assert.Equal(t, "bucket", url.Host)
		assert.Equal(t, "/folder/file.txt", url.Path)
	})

	t.Run("file uri", func(t *testing.T) {
		uri := "file:///root/folder/file.txt"
		url, err := url.Parse(uri)
		assert.NoError(t, err, "must be able to parse the uri")
		assert.NotNil(t, url)
		assert.Empty(t, url.Host)
		assert.Equal(t, "file", url.Scheme)
		assert.Equal(t, "/root/folder/file.txt", url.Path)
	})
}
