package models

import (
	"context"
	"encoding/hex"
	"hash/fnv"
	"path"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

var (
	ErrInvalidContentType = errors.New("invalid content type")
)

type ContentType string

const (
	TextCSVContentType      ContentType = "text/csv"
	OpenXMLExcelContentType ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	IntuitQFXContentType    ContentType = "application/vnd.intu.QFX"
	CAMT053ContentType      ContentType = "application/vnd.camt-053" // Not a real content type, but what monetr will use.
)

var (
	contentTypeExtensions = map[ContentType][]string{
		CAMT053ContentType:      {"xml"},
		IntuitQFXContentType:    {"qfx", "ofx"},
		OpenXMLExcelContentType: {"xlsx"},
		TextCSVContentType:      {"csv"},
	}
)

func GetContentTypeIsValid(contentType string) bool {
	_, ok := contentTypeExtensions[ContentType(contentType)]
	return ok
}

type Uploadable interface {
	// FileKind can be added to a model struct, and indicates that this model can
	// have a file associated with it. This determines the prefix to use for the
	// file upload itself. Cannot be blank.
	FileKind() string
	// FileExpiration tells the uploader to mark a file to be deleted after so
	// much time. If your model defines this, it will mark the file. Otherwise
	// this can be nil.
	FileExpiration(clock clock.Clock) *time.Time
}

var (
	_ pg.BeforeInsertHook = (*File)(nil)
	_ Identifiable        = File{}
)

type File struct {
	tableName string `pg:"files"`

	FileId        ID[File]    `json:"fileId" pg:"file_id,notnull,pk"`
	AccountId     ID[Account] `json:"-" pg:"account_id,notnull,pk"`
	Account       *Account    `json:"-" pg:"rel:has-one"`
	Kind          string      `json:"kind" pg:"kind,notnull"`
	Name          string      `json:"name" pg:"name,notnull"`
	ContentType   ContentType `json:"contentType" pg:"content_type,notnull"`
	Size          uint64      `json:"size" pg:"size,notnull"`
	CreatedAt     time.Time   `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy     ID[User]    `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser *User       `json:"-" pg:"rel:has-one,fk:created_by"`
	ExpiresAt     *time.Time  `json:"expiresAt" pg:"expires_at"`
	DeletedAt     *time.Time  `json:"deletedAt" pg:"deleted_at"`
	ReconciledAt  *time.Time  `json:"-" pg:"reconciled_at"`
}

func (File) IdentityPrefix() string {
	return "file"
}

func (o *File) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.FileId.IsZero() {
		o.FileId = NewID[File]()
	}

	// Fixes weird bug in tests where the time gets truncated by the database
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now()
	}

	return ctx, nil
}

func (o *File) GetStorePath() (string, error) {
	if o.FileId.IsZero() {
		return "", errors.New("no valid file ID")
	}

	if o.AccountId.IsZero() {
		return "", errors.New("no valid account ID")
	}

	if o.Kind == "" {
		return "", errors.New("no valid file kind")
	}

	// Hex gives us a 2 character prefix that we can use in the directory in order
	// to break files up. This makes some filesystems happier and avoids the
	// potential of having a single directory with a LOT of folders in in.
	// But we also want it in its own directory per account ID, so we are breaking
	// that up as well.
	accountHex := hex.EncodeToString(fnv.New32().Sum([]byte(o.AccountId)))
	// Same thing with the file, if one account has a ton of files we want to
	// break it up a bit. This is kind of future planning where monetr will at
	// some point support things like transaction attachments and there could
	// theoretically be multiple attachments per transactions for some users.
	fileHex := hex.EncodeToString(fnv.New32().Sum([]byte(o.FileId)))
	return path.Join(
		"data",
		o.Kind,
		accountHex[0:2],
		o.AccountId.WithoutPrefix(),
		fileHex[0:2],
		o.FileId.WithoutPrefix(),
	), nil
}
