package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidContentType = errors.New("invalid content type")
)

type FileInfo struct {
	Name          string
	AccountId     uint64
	BankAccountId uint64
	ContentType   ContentType
}

// Storage is the interface for reading and writing files presented to monetr by
// clients. These files might be images, CSV files or QFX files or something
// else entirely. Files are stored in a random path that is returned when the
// file is written. Files can only be retrieved using their path. Files cannot
// be listed and the tree cannot be walked. Files should only be interacted with
// via this interface within monetr.
type Storage interface {
	// Store will take the buffer provided by the caller and persist it to
	// whatever storage implementation is being provided under this interface. It
	// will then return a URI that can be used to retrieve the file. This URI is
	// randomized and is not specific to any single client. An error is returned
	// if the file is not able to be stored. Depending on the implementation the
	// file may still be present in whatever storage system even if the file was
	// not successfully stored. This should be considered on a per-implementation
	// basis as it will be unique to the implementation itself.
	Store(ctx context.Context, buf io.ReadSeekCloser, info FileInfo) (uri string, err error)
	// Read will take a file URI and will read it from the underlying storage
	// system. If the URI provided is not for the storage interface under this
	// then an error will be returned. For example; if this is backed by a file
	// system but the provided URI is an S3 protocol, then this would return an
	// error for protocol mismatch. If a file can be read then an buffer will be
	// returned for that file.
	Read(ctx context.Context, uri string) (buf io.ReadCloser, contentType ContentType, err error)
}

func getStorePath(info FileInfo) (string, error) {
	name := uuid.NewString()
	extension, ok := contentTypeExtensions[info.ContentType]
	if !ok {
		return "", errors.WithStack(ErrInvalidContentType)
	}
	name = strings.ReplaceAll(name, "-", "")

	key := fmt.Sprintf(
		"%08X/%08X/%s.%s",
		info.AccountId, info.BankAccountId, name, extension,
	)
	return key, nil
}

type ContentType string

const (
	TextCSVContentType      ContentType = "text/csv"
	OpenXMLExcelContentType ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	IntuitQFXContentType    ContentType = "application/vnd.intu.QFX"
)

var (
	contentTypeExtensions = map[ContentType]string{
		TextCSVContentType:      "csv",
		OpenXMLExcelContentType: "xlsx",
		IntuitQFXContentType:    "qfx",
	}
)

func getContentTypeByPath(filePath string) (ContentType, error) {
	extension := strings.TrimPrefix(path.Ext(filePath), ".")
	for contentType, ext := range contentTypeExtensions {
		if ext == extension {
			return contentType, nil
		}
	}

	return "", errors.WithStack(ErrInvalidContentType)
}

func GetContentTypeIsValid(contentType string) bool {
	_, ok := contentTypeExtensions[ContentType(contentType)]
	return ok
}
