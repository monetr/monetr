package models

import (
	"context"
	"encoding/hex"
	"hash/fnv"
	"path"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
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
		IntuitQFXContentType:    {"qfx", "ofx", "qbo"},
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
	_ bun.BeforeAppendModelHook = (*File)(nil)
	_ Identifiable              = File{}
)

type File struct {
	bun.BaseModel `bun:"table:files,alias:file"`

	FileId        ID[File]    `json:"fileId" bun:"file_id,notnull,pk"`
	AccountId     ID[Account] `json:"-" bun:"account_id,notnull,pk"`
	Account       *Account    `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	Kind          string      `json:"kind" bun:"kind,notnull,nullzero"`
	Name          string      `json:"name" bun:"name,notnull,nullzero"`
	ContentType   ContentType `json:"contentType" bun:"content_type,notnull,nullzero"`
	Size          uint64      `json:"size" bun:"size,notnull,nullzero"`
	CreatedAt     time.Time   `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy     ID[User]    `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser *User       `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
	ExpiresAt     *time.Time  `json:"expiresAt" bun:"expires_at"`
	DeletedAt     *time.Time  `json:"deletedAt" bun:"deleted_at"`
	ReconciledAt  *time.Time  `json:"-" bun:"reconciled_at"`
}

func (File) IdentityPrefix() string {
	return "file"
}

func (o *File) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.FileId.IsZero() {
			o.FileId = NewID[File]()
		}

		// Fixes weird bug in tests where the time gets truncated by the database
		if o.CreatedAt.IsZero() {
			o.CreatedAt = time.Now()
		}
	}

	return nil
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
