package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// Storage is the interface for reading and writing files presented to monetr by clients. These files might be images,
// CSV files or QFX files or something else entirely. Files are stored in a random path that is returned when the file
// is written. Files can only be retrieved using their path. Files cannot be listed and the tree cannot be walked. Files
// should only be interacted with via this interface within monetr.
type Storage interface {
	// Store will take the buffer provided by the caller and persist it to whatever storage implementation is being
	// provided under this interface. It will then return a URI that can be used to retrieve the file. This URI is
	// randomized and is not specific to any single client. An error is returned if the file is not able to be stored.
	// Depending on the implementation the file may still be present in whatever storage system even if the file was not
	// successfully stored. This should be considered on a per-implementation basis as it will be unique to the
	// implementation itself.
	Store(ctx context.Context, buf io.ReadSeekCloser) (uri string, err error)
	// Read will take a file URI and will read it from the underlying storage system. If the URI provided is not for the
	// storage interface under this then an error will be returned. For example; if this is backed by a file system but
	// the provided URI is an S3 protocol, then this would return an error for protocol mismatch. If a file can be read
	// then an buffer will be returned for that file.
	Read(ctx context.Context, uri string) (buf io.ReadCloser, err error)
}

func getStorePath() string {
	chunk := uuid.NewString()
	name := uuid.NewString()
	key := fmt.Sprintf("%s/%s.[FILETYPE]", chunk, name)
	return key
}
