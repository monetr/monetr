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

		fs := &filesystemStorage{
			log:           testutils.GetLog(t),
			baseDirectory: tempDir,
		}

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
}
