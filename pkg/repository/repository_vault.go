//+build vault

package repository

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

type repositoryBase struct {
	userId, accountId uint64
	txn               *pg.Tx
	vault             *api.Client
}

func (r *repositoryBase) CreatePlaidLink(link *models.PlaidLink) error {
	_, err := r.txn.Model(link).Insert(link)
	if err != nil {
		return errors.Wrap(err, "failed to create plaid link")
	}

	_, err = r.vault.Logical().Write(fmt.Sprintf("plaidLink/%d", link.PlaidLinkID), map[string]interface{}{
		"accessToken":   link.AccessToken,
		"itemId":        link.ItemId,
		"institutionId": link.InstitutionId,
	})
	if err != nil {
		return errors.Wrap(err, "failed to store access token")
	}

	return nil
}
