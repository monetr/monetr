package models

import (
	"github.com/pkg/errors"
	"time"
)

type Account struct {
	tableName string `pg:"accounts"`

	AccountId uint64 `json:"accountId" pg:"account_id,notnull,pk,type:'bigserial'"`
	Timezone  string `json:"timezone" pg:"timezone,notnull,default:'UTC'"`
}

func (a *Account) GetTimezone() (*time.Location, error) {
	location, err := time.LoadLocation(a.Timezone)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse account timezone as location")
	}

	return location, nil
}
