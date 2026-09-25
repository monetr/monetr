package models

import (
	"context"

	"github.com/uptrace/bun"
)

// UserRole is also a PostgreSQL type `user_role`.
type UserRole string

const (
	UserRoleMember UserRole = "member"
	UserRoleOwner  UserRole = "owner"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:user"`

	UserId    ID[User]    `json:"userId" bun:"user_id,notnull,pk"`
	LoginId   ID[Login]   `json:"loginId" bun:"login_id,notnull,unique:per_account,nullzero"`
	Login     *Login      `json:"login,omitempty" bun:"rel:belongs-to,join:login_id=login_id"`
	AccountId ID[Account] `json:"accountId" bun:"account_id,notnull,unique:per_account,nullzero"`
	Account   *Account    `json:"account" bun:"rel:belongs-to,join:account_id=account_id"`
	Role      UserRole    `json:"role" bun:"role,notnull,nullzero"`
}

var (
	_ bun.BeforeAppendModelHook = (*User)(nil)
)

func (o *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.UserId.IsZero() {
			o.UserId = NewID[User]()
		}
	}

	return nil
}

func (User) IdentityPrefix() string {
	return "user"
}
