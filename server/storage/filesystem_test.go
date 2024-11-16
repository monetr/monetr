package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bufferWrapper struct {
	*bytes.Reader
}

func (b *bufferWrapper) Close() error {
	return nil
}

func TestFilesystemStorage(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "monetr")
		require.NoError(t, err, "must create temp directory")
		log := testutils.GetLog(t)

		fs, err := NewFilesystemStorage(log, tempDir)
		assert.NoError(t, err, "must not have an error creating the filesystem storage interface")

		input := []byte("i am a test string")
		buf := &bufferWrapper{bytes.NewReader(input)}

		uri, err := fs.Store(context.Background(), buf, FileInfo{
			Name:        "Test file.csv",
			AccountId:   "acct_123",
			ContentType: TextCSVContentType,
		})
		assert.NoError(t, err, "should not have an error storing a file")
		assert.NotEmpty(t, uri, "should have returned a valid uri")
		fmt.Println(uri)

		content, contentType, err := fs.Read(context.Background(), uri)
		assert.NoError(t, err, "should read file back")
		assert.Equal(t, TextCSVContentType, contentType, "should have a csv content type")
		assert.NotNil(t, content, "content buffer should not be nil")

		allContent, err := io.ReadAll(content)
		assert.NoError(t, err, "should read all file content without error")
		assert.EqualValues(t, input, allContent, "file content should match input")
	})

	t.Run("blank base path", func(t *testing.T) {
		log := testutils.GetLog(t)

		fs, err := NewFilesystemStorage(log, "")
		assert.EqualError(t, err, "base directory for filesystem storage must be an absolute path")
		assert.Nil(t, fs, "filesystem interface should not be returned if path is invalid")
	})

	t.Run("non-existant base directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "monetr")
		require.NoError(t, err, "must create temp directory")
		log := testutils.GetLog(t)

		// Add something onto the end of the temp dir path so that we know it does
		// not already exist.
		fsPath := path.Join(tempDir, "storage")
		fs, err := NewFilesystemStorage(log, fsPath)
		assert.NoError(t, err, "no error should be returne for a non-existant path")
		assert.NotNil(t, fs, "fs interface should be returned if we made a directory")
	})

	t.Run("base directory is actually a file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "monetr")
		require.NoError(t, err, "must create temp directory")
		tempFile, err := os.CreateTemp(tempDir, "temp-*")
		require.NoError(t, err, "must create a temp file")
		require.NotNil(t, tempFile, "temp file must be valid")

		log := testutils.GetLog(t)

		fs, err := NewFilesystemStorage(log, tempFile.Name())
		assert.EqualError(t, err, "filesystem base directory specified is not a directory, it is a file")
		assert.Nil(t, fs, "filesystem interface should not be returned if path is invalid")
	})
}

func TestFilesystemStorageStore(t *testing.T) {
	t.Run("store file", func(t *testing.T) {
		buf := &bufferWrapper{bytes.NewReader([]byte("I am the contents of a file"))}
		tempDir, err := os.MkdirTemp("", "monetr")
		require.NoError(t, err, "must create temp directory")
		log := testutils.GetLog(t)

		fs, err := NewFilesystemStorage(log, tempDir)
		assert.NoError(t, err, "must not have an error creating the filesystem storage interface")

		uri, err := fs.Store(context.Background(), buf, FileInfo{
			Name:        "Test file.csv",
			Kind:        "transaction/uploads",
			AccountId:   "acct_01jcvdac6dzbrzm1ht90zyt65r",
			ContentType: TextCSVContentType,
		})
		assert.NoError(t, err, "should be able to store file successfully")
		assert.NotEmpty(t, uri, "URI returned should not be blank")

		parsed, err := url.Parse(uri)
		assert.NoError(t, err, "store must return a valid URI")
		assert.Equal(t, "file", parsed.Scheme, "URI scheme should be file")
		assert.Empty(t, parsed.Host, "URI host should be empty for filesystem")
		assert.NotEmpty(t, parsed.Path, "URI path should not be empty")
	})
}
