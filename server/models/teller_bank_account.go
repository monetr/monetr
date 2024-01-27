package models

import "time"

type TellerBankAccountStatus uint8

//go:generate stringer -type=TellerBankAccountStatus -output=teller_bank_account.strings.go
const (
	TellerBankAccountStatusUnknown TellerBankAccountStatus = 0
	TellerBankAccountStatusOpen    TellerBankAccountStatus = 1
	TellerBankAccountStatusClosed  TellerBankAccountStatus = 2
)

type TellerBankAccount struct {
	tableName string `pg:"teller_bank_accounts"`

	TellerBankAccountId uint64                  `json:"-" pg:"teller_bank_account_id"`
	AccountId           uint64                  `json:"-" pg:"account_id,notnull,type:'bigint',unique:per_account"`
	Account             *Account                `json:"-" pg:"rel:has-one"`
	TellerLinkId        uint64                  `json:"-" pg:"teller_link_id,type:'bigint'"`
	TellerLink          *TellerLink             `json:"-" pg:"rel:has-one"`
	TellerId            string                  `json:"-" pg:"teller_id,notnull"`
	InstitutionId       string                  `json:"institutionId" pg:"intitution_id,notnull"`
	InstitutionName     string                  `json:"institutionName" pg:"institution_name,notnull"`
	Mask                string                  `json:"mask" pg:"mask,notnull"`
	Name                string                  `json:"name" pg:"name,notnull"`
	Type                string                  `json:"type" pg:"type,notnull"`
	SubType             string                  `json:"subType" pg:"sub_type,notnull"`
	Status              TellerBankAccountStatus `json:"status" pg:"status,notnull,default:0"`
	UpdatedAt           time.Time               `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedAt           time.Time               `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId     uint64                  `json:"createdByUserId" pg:"created_by_user_id,notnull"`
	CreatedByUser       *User                   `json:"-" pg:"rel:has-one,fk:created_by_user_id"`
}
