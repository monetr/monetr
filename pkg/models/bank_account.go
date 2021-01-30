package models

type BankAccount struct {
	tableName string `pg:"bank_accounts"`

	BankAccountId     uint64   `json:"bankAccountId" pg:"bank_account_id,notnull,pk,type:'bigserial'"`
	AccountId         uint64   `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE"`
	Account           *Account `json:"-" pg:"rel:has-one"`
	LinkId            uint64   `json:"linkId" pg:"link_id,notnull,on_delete:CASCADE"`
	Link              *Link    `json:"-" pg:"rel:has-one"`
	PlaidAccountId    string   `json:"-" pg:"plaid_account_id,notnull"`
	AvailableBalance  int64    `json:"availableBalance" pg:"available_balance,notnull,use_zero"`
	CurrentBalance    int64    `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	Mask              string   `json:"mask" pg:"mask,notnull"`
	Name              string   `json:"name,omitempty" pg:"name"`
	PlaidName         string   `json:"originalName" pg:"original_name,notnull"`
	PlaidOfficialName string   `json:"officialName" pg:"official_name,notnull"`
	Type              string   `json:"accountType" pg:"account_type,notnull"`
	SubType           string   `json:"accountSubType" pg:"account_sub_type,notnull"`
}
