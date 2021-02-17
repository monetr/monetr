//+build !vault

package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreatePlaidLink(link *models.PlaidLink) error {
	_, err := r.txn.Model(link).Insert(link)
	return errors.Wrap(err, "failed to create plaid link")
}
