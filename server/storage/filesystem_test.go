package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"testing"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
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

		file := models.File{
			FileId:      models.ID[models.File]("file_01kheqbr884scw3wt8v70g95yk"),
			AccountId:   models.ID[models.Account]("acct_01kheqbr884scw3wt8vacxg1sd"),
			Name:        "Test file.csv",
			ContentType: models.TextCSVContentType,
			Kind:        "transactions/uploads",
			Size:        100,
		}
		err = fs.Store(context.Background(), buf, file)
		assert.NoError(t, err, "should not have an error storing a file")

		content, err := fs.Read(context.Background(), file)
		defer content.Close()
		assert.NoError(t, err, "should read file back")
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

		file := models.File{
			FileId:      models.NewID[models.File](),
			AccountId:   models.NewID[models.Account](),
			Kind:        "transactions/upload",
			Name:        "transactions (10).ofx",
			ContentType: models.IntuitQFXContentType,
			Size:        100,
		}
		err = fs.Store(context.Background(), buf, file)
		assert.NoError(t, err, "should be able to store file successfully")
	})
}
