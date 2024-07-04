package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
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
		log := testutils.GetLog(t)

		fs, err := NewFilesystemStorage(log, "/foo/bar")
		assert.EqualError(t, err, "must provide a valid base directory for filesystem storage: stat /foo/bar: no such file or directory")
		assert.Nil(t, fs, "filesystem interface should not be returned if path is invalid")
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
