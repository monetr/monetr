package tasks

import (
	"context"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/internal/platypus"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RepairPlaidWebhooksTaskOptions struct {
	Log           *logrus.Entry
	DryRun        bool
	AccountId     uint64
	LinkId        uint64
	NewWebhookURL string
}

type RepairPlaidWebhooksTask struct {
	log           *logrus.Entry
	dryRun        bool
	accountId     uint64
	linkId        uint64
	newWebhookURL string
	db            pg.DBI
	plaid         platypus.Plaid
}

func NewRepairPlaidWebhooksTask(options RepairPlaidWebhooksTaskOptions, db pg.DBI, plaid platypus.Plaid) *RepairPlaidWebhooksTask {
	return &RepairPlaidWebhooksTask{
		log:           options.Log,
		dryRun:        options.DryRun,
		accountId:     options.AccountId,
		linkId:        options.LinkId,
		newWebhookURL: options.NewWebhookURL,
		db:            db,
		plaid:         plaid,
	}
}

func (r *RepairPlaidWebhooksTask) Execute(ctx context.Context) (linksUpdated, updatesFailed int, _ error) {
	if r.accountId > 0 {
		r.log = r.log.WithField("accountId", r.accountId)
		if r.linkId > 0 {
			r.log = r.log.WithField("linkId", r.linkId)
			r.log.Info("Plaid link webhook URL will be checked and updated")
		} else {
			r.log.Info("all Plaid links webhook's URL for the specified account will be checked and updated")
		}
	} else if r.linkId > 0 {
		return 0, 0, errors.Errorf("specific links can only be updated when an account is specified")
	} else {
		r.log.Info("all Plaid links webhook's URL will be checked and updated")
	}

	links, err := r.getLinks(ctx)
	if err != nil {
		return 0, 0, err
	}

	if len(links) == 0 {
		r.log.Info("no links were found")
		return 0, 0, nil
	}

	for _, link := range links {
		if didUpdate, err := r.updateLink(ctx, link); err != nil {
			r.log.WithField("linkId", link.LinkId).WithError(err).Error("failed to update link")
			updatesFailed++
			continue
		} else if didUpdate {
			linksUpdated++
		}
	}

	return linksUpdated, updatesFailed, nil
}

func (r *RepairPlaidWebhooksTask) getLinks(ctx context.Context) ([]models.Link, error) {
	links := make([]models.Link, 0)
	query := r.db.Model(&links).
		Relation("PlaidLink").
		Join(`INNER JOIN "accounts" AS "account"`).
		JoinOn(`"account"."account_id" = "link"."account_id"`).
		Where(`"link"."link_type" = ?`, models.PlaidLinkType)

	if r.accountId > 0 {
		query = query.Where(`"link"."account_id" = ?`, r.accountId)

		if r.linkId > 0 {
			query = query.Where(`"link"."link_id" = ?`, r.linkId)
		}
	} else {
		// If we are not querying by a specific account, then make sure we only retrieve accounts that are currently
		// active.
		query = query.Where(`("account"."subscription_active_until" IS NULL OR "account"."subscription_active_until" > now())`)
	}

	if err := query.Select(&links); err != nil {
		r.log.WithError(err).Errorf("failed to retrieve links for webhook URL update")
		return nil, errors.Wrap(err, "failed to retrieve links for webhook URL update")
	}

	return links, nil
}

func (r *RepairPlaidWebhooksTask) updateLink(ctx context.Context, link models.Link) (didUpdate bool, _ error) {
	log := r.log.WithFields(logrus.Fields{
		"accountId": link.AccountId,
		"linkId":    link.LinkId,
	})

	if link.PlaidLink == nil {
		return false, errors.Errorf("cannot update link without Plaid Link")
	}

	log = log.WithFields(logrus.Fields{
		"institutionId": link.PlaidLink.InstitutionId,
		"institution":   link.PlaidLink.InstitutionName,
	})

	// Check to see if the current webhook url for the link matches the new one. If it does then we can skip it.
	if strings.EqualFold(link.PlaidLink.WebhookUrl, r.newWebhookURL) {
		log.Info("no need to update link, webhook URL is up to date")
		return false, nil
	}

	log.Debugf("creating plaid client for link")
	plaid, err := r.plaid.NewClientFromItemId(ctx, link.PlaidLink.ItemId)
	if err != nil {
		return false, errors.Wrap(err, "failed to create plaid client for link")
	}

	if r.dryRun {
		log.Infof("Plaid webhook URL will be updated from: %s to: %s", link.PlaidLink.WebhookUrl, r.newWebhookURL)
		return true, nil
	}

	if err = plaid.UpdateWebhook(ctx, r.newWebhookURL); err != nil {
		return false, errors.Wrap(err, "failed to update webhook URL")
	}

	result, err := r.db.Model(&models.PlaidLink{}).
		Set(`"webhook_url" = ?`, r.newWebhookURL).
		Where(`"plaid_link"."item_id" = ?`, link.PlaidLink.ItemId).
		Update()
	if err != nil {
		return true, errors.Wrap(err, "failed to update Plaid Link records with new webhook URL")
	}

	if result.RowsAffected() == 0 {
		return true, errors.Errorf("no rows were updated in the database for the Plaid Link")
	}

	return true, nil
}
