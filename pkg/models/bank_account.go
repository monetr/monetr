package models

type BankAccount struct {
	tableName string `pg:"bank_accounts"`

	BankAccountId     uint64   `json:"bankAccountId" pg:"bank_account_id,notnull,pk,type:'bigserial'"`
	AccountId         uint64   `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE"`
	Account           *Account `json:"-" pg:"rel:has-one"`
	LinkId            uint64   `json:"linkId" pg:"link_id,notnull,on_delete:CASCADE"`
	Link              *Link    `json:"link,omitempty" pg:"rel:has-one"`
	PlaidAccountId    string   `json:"-" pg:"plaid_account_id"`
	AvailableBalance  int64    `json:"availableBalance" pg:"available_balance,notnull,use_zero"`
	CurrentBalance    int64    `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	Mask              string   `json:"mask" pg:"mask"`
	Name              string   `json:"name,omitempty" pg:"name,notnull"`
	PlaidName         string   `json:"originalName" pg:"plaid_name"`
	PlaidOfficialName string   `json:"officialName" pg:"plaid_official_name"`
	Type              string   `json:"accountType" pg:"account_type"`
	SubType           string   `json:"accountSubType" pg:"account_sub_type"`
}
