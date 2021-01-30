package models

type Login struct {
	tableName string `pg:"logins"`

	LoginId      uint64 `json:"loginId" pg:"login_id,notnull,pk,type:'bigserial'"`
	Email        string `json:"email" pg:"email,notnull,unique"`
	PasswordHash string `json:"-" pg:"password_hash,notnull"`

	Users []User `json:"users,omitempty" pg:"rel:has-many"`
}
