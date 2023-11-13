package models

import "time"

type File struct {
	tableName string `pg:"files"`

	FileId          uint64       `json:"fileId" pg:"file_id,notnull,pk,type:'bigserial'"`
	AccountId       uint64       `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigserial'"`
	Account         *Account     `json:"-" pg:"rel:has-one"`
	BankAccountId   uint64       `json:"bankAccountId" pg:"bank_account_id,notnull,on_delete:RESTRICT,type:'bigint'"`
	BankAccount     *BankAccount `json:"-" pg:"rel:has-one"`
	Name            string       `json:"name" pg:"name,notnull"`
	ContentType     string       `json:"contentType" pg:"content_type,notnull"`
	Size            uint64       `json:"size" pg:"size,notnull"`
	ObjectUri       string       `json:"-" pg:"object_uri,notnull"`
	CreatedAt       time.Time    `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId uint64       `json:"createdByUserId" pg:"created_by_user_id,notnull,on_delete:CASCADE"`
	CreatedByUser   *User        `json:"-" pg:"rel:has-one,fk:created_by_user_id"`
}
