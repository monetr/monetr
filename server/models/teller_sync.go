package models

import "time"

type TellerSync struct {
	tableName string `pg:"teller_syncs"`

	TellerSyncId        uint64             `json:"tellerSyncId" pg:"teller_sync_id,notnull,pk,type:'bigserial'"`
	AccountId           uint64             `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE"`
	Account             *Account           `json:"-" pg:"rel:has-one"`
	TellerBankAccountId uint64             `json:"-" pg:"teller_bank_account_id,notnull"`
	TellerBankAccount   *TellerBankAccount `json:"-" pg:"rel:has-one"`
	Timestamp           time.Time          `json:"timestamp" pg:"timestamp,notnull"`
	Trigger             string             `json:"trigger" pg:"trigger,notnull"`
	ImmutableTimestamp  time.Time          `json:"immutableTimestamp" pg:"immutable_timestamp,notnull"`
	Added               int                `json:"added" pg:"added,notnull,use_zero"`
	Modified            int                `json:"modified" pg:"modified,notnull,use_zero"`
	Removed             int                `json:"removed" pg:"removed,notnull,use_zero"`
}
