package models

type BankAccount struct {
	tableName string `sql:"bank_accounts"`

	BankAccountId     uint64   `json:"bankAccountId" sql:"bank_account_id,notnull,pk,type:'bigserial'"`
	AccountId         uint64   `json:"-" sql:"account_id,notnull,pk,on_delete:CASCADE"`
	Account           *Account `json:"-" sql:"rel:has-one"`
	LinkId            uint64   `json:"linkId" sql:"link_id,notnull,on_delete:CASCADE"`
	Link              *Link    `json:"-" sql:"rel:has-one"`
	PlaidAccountId    string   `json:"-" sql:"plaid_account_id,notnull"`
	AvailableBalance  int64    `json:"availableBalance" sql:"available_balance,notnull,use_zero"`
	CurrentBalance    int64    `json:"currentBalance" sql:"current_balance,notnull,use_zero"`
	Mask              string   `json:"mask" sql:"mask,notnull"`
	Name              string   `json:"name,omitempty" sql:"name,null"`
	PlaidName         string   `json:"originalName" sql:"original_name,notnull"`
	PlaidOfficialName string   `json:"officialName" sql:"official_name,notnull"`
	Type              string   `json:"accountType" sql:"account_type,notnull"`
	SubType           string   `json:"accountSubType" sql:"account_sub_type,notnull"`
}
