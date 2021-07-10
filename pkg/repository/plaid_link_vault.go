//+build vault

package repository

import (
	"fmt"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
)

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
