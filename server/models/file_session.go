package models

import (
	"time"

	"github.com/monetr/monetr/server/formats"
)

type FileSession struct {
	tableName string `pg:"file_sessions"`

	FileSessionId   uint64              `json:"fileSessionId" pg:"file_session_id,notnull,pk,type:'bigserial'"`
	AccountId       uint64              `json:"-" pg:"account_id,notnull,pk,type:'bigserial'"`
	Account         *Account            `json:"-" pg:"rel:has-one"`
	BankAccountId   uint64              `json:"bankAccountId" pg:"bank_account_id,notnull,pk,type:'bigint'"`
	BankAccount     *BankAccount        `json:"-" pg:"rel:has-one"`
	FileId          uint64              `json:"fileId" pg:"file_id,notnull,type:'bigint'"`
	File            *File               `json:"-" pg:"rel:has-one"`
	Fields          *formats.FieldIndex `json:"fields" pg:"fields,type:'INT[]'"`
	CreatedAt       time.Time           `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId uint64              `json:"createdByUserId" pg:"created_by_user_id,notnull"`
	CreatedByUser   *User               `json:"-" pg:"rel:has-one,fk:created_by_user_id"`
}
