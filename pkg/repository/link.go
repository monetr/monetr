package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
	"time"
)

func (r *repositoryBase) GetLink(linkId uint64) (*models.Link, error) {
	var link models.Link
	err := r.txn.Model(&link).
		Relation("PlaidLink").
		Relation("BankAccounts").
		Where(`"link"."link_id" = ? AND "link"."account_id" = ?`, linkId, r.AccountId()).
		Limit(1).
		Select(&link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get link")
	}

	return &link, nil
}

func (r *repositoryBase) GetLinks() ([]models.Link, error) {
	var result []models.Link
	err := r.txn.Model(&result).
		Where(`"link"."account_id" = ?`, r.accountId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve links")
	}

	return result, nil
}

func (r *repositoryBase) CreateLink(link *models.Link) error {
	userId := r.UserId()
	now := time.Now().UTC()
	link.AccountId = r.AccountId()
	link.CreatedByUserId = userId
	link.UpdatedByUserId = &userId
	link.CreatedAt = now
	link.UpdatedAt = now

	_, err := r.txn.Model(link).Insert(link)
	return errors.Wrap(err, "failed to insert link")
}

func (r *repositoryBase) UpdateLink(link *models.Link) error {
	userId := r.UserId()
	link.AccountId = r.AccountId()
	link.UpdatedByUserId = &userId
	link.UpdatedAt = time.Now().UTC()

	_, err := r.txn.Model(link).WherePK().Returning(`*`).UpdateNotZero(link)
	return errors.Wrap(err, "failed to update link")
}
