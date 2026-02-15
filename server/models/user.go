package models

import (
	"context"

	"github.com/go-pg/pg/v10"
)

// UserRole is also a PostgreSQL type `user_role`.
type UserRole string

const (
	UserRoleMember UserRole = "member"
	UserRoleOwner  UserRole = "owner"
)

type User struct {
	tableName string `pg:"users"`

	UserId    ID[User]    `json:"userId" pg:"user_id,notnull,pk"`
	LoginId   ID[Login]   `json:"loginId" pg:"login_id,notnull,unique:per_account"`
	Login     *Login      `json:"login,omitempty" pg:"rel:has-one"`
	AccountId ID[Account] `json:"accountId" pg:"account_id,notnull,unique:per_account"`
	Account   *Account    `json:"account" pg:"rel:has-one"`
	Role      UserRole    `json:"role" pg:"role,notnull"`
}

var (
	_ pg.BeforeInsertHook = (*User)(nil)
)

func (o *User) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.UserId.IsZero() {
		o.UserId = NewID[User]()
	}

	return ctx, nil
}

func (User) IdentityPrefix() string {
	return "user"
}
