package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"runtime"
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
		require.NoError(t, err, "should read file back")
		require.NotNil(t, content, "content buffer should not be nil")
		defer content.Close()

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

func TestFilesystemStoragePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not meaningful on Windows")
	}

	tempDir, err := os.MkdirTemp("", "monetr")
	require.NoError(t, err, "must create temp directory")
	log := testutils.GetLog(t)

	// Use a non-existent base path so the constructor's MkdirAll codepath runs
	// and we can validate the base directory's mode too.
	basePath := path.Join(tempDir, "storage")
	fs, err := NewFilesystemStorage(log, basePath)
	require.NoError(t, err, "must create the filesystem storage interface")

	baseInfo, err := os.Stat(basePath)
	require.NoError(t, err, "must stat base directory")
	require.True(t, baseInfo.IsDir(), "base path must be a directory")
	assert.Equal(t, os.FileMode(0700), baseInfo.Mode().Perm(), "base directory must be drwx------")

	file := models.File{
		FileId:      models.NewID[models.File](),
		AccountId:   models.NewID[models.Account](),
		Kind:        "transactions/uploads",
		Name:        "statement.csv",
		ContentType: models.TextCSVContentType,
		Size:        100,
	}
	input := []byte("sensitive financial data")
	buf := &bufferWrapper{bytes.NewReader(input)}
	require.NoError(t, fs.Store(context.Background(), buf, file), "must store file")

	storedPath, err := file.GetStorePath()
	require.NoError(t, err, "must derive store path")
	fullPath := path.Join(basePath, storedPath)

	fileInfo, err := os.Stat(fullPath)
	require.NoError(t, err, "must stat stored file")
	assert.True(t, fileInfo.Mode().IsRegular(), "stored entry must be a regular file")
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm(), "stored file must be -rw-------")

	parentInfo, err := os.Stat(path.Dir(fullPath))
	require.NoError(t, err, "must stat parent directory of stored file")
	require.True(t, parentInfo.IsDir(), "parent must be a directory")
	assert.Equal(t, os.FileMode(0700), parentInfo.Mode().Perm(), "parent directory must be drwx------")

	// Confirm the server can still read what it wrote under the tightened mode.
	rdr, err := fs.Read(context.Background(), file)
	require.NoError(t, err, "must read file back")
	defer rdr.Close()
	got, err := io.ReadAll(rdr)
	require.NoError(t, err, "must read all file content")
	assert.Equal(t, input, got, "read content must match written content")
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
