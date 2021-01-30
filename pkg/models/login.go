package models

type Login struct {
	tableName string `sql:"logins"`

	LoginId      uint64 `json:"loginId" sql:"login_id,notnull,pk,type:'bigserial'"`
	Email        string `json:"email" sql:"email,notnull,unique"`
	PasswordHash string `json:"-" sql:"password_hash,notnull"`

	Users []User `json:"users,omitempty" sql:"rel:has-many"`
}
