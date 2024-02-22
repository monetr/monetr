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

	TellerBankAccountId uint64                  `json:"-" pg:"teller_bank_account_id,pk,notnull"`
	AccountId           uint64                  `json:"-" pg:"account_id,pk,notnull,type:'bigint',unique:per_account"`
	Account             *Account                `json:"-" pg:"rel:has-one"`
	TellerLinkId        uint64                  `json:"-" pg:"teller_link_id,type:'bigint'"`
	TellerLink          *TellerLink             `json:"-" pg:"rel:has-one"`
	TellerId            string                  `json:"-" pg:"teller_id,notnull"`
	InstitutionId       string                  `json:"institutionId" pg:"institution_id,notnull"`
	InstitutionName     string                  `json:"institutionName" pg:"institution_name,notnull"`
	Mask                string                  `json:"mask" pg:"mask,notnull"`
	Name                string                  `json:"name" pg:"name,notnull"`
	Type                string                  `json:"type" pg:"type,notnull"`
	SubType             string                  `json:"subType" pg:"sub_type,notnull"`
	Status              TellerBankAccountStatus `json:"status" pg:"status,notnull,default:0"`
	LedgerBalance       *int64                  `json:"ledgerBalance" pg:"ledger_balance,use_zero"`
	UpdatedAt           time.Time               `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedAt           time.Time               `json:"createdAt" pg:"created_at,notnull"`
	BalancedAt          *time.Time              `json:"balancedAt" pg:"balanced_at"`
}

func (t TellerBankAccount) GetIsCredit() bool {
	return t.Type == "credit" && t.SubType == "credit_card"
}
