package models

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
)

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
	Name          string      `json:"name" pg:"name,notnull"`
	ContentType   string      `json:"contentType" pg:"content_type,notnull"`
	Size          uint64      `json:"size" pg:"size,notnull"`
	BlobUri       string      `json:"-" pg:"blob_uri,notnull"`
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
		o.FileId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
