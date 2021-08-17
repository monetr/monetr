package jobs

import (
	"context"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/rest-api/pkg/internal/plaid_helper"
	"github.com/monetr/rest-api/pkg/metrics"
	"github.com/monetr/rest-api/pkg/pubsub"
	"github.com/monetr/rest-api/pkg/secrets"
	"github.com/sirupsen/logrus"
	"time"
)

type nonDistributedJobManager struct {
	log          *logrus.Entry
	db           *pg.DB
	plaidClient  plaid_helper.Client
	plaidSecrets secrets.PlaidSecretsProvider
	stats        *metrics.Stats
	ps           pubsub.PublishSubscribe
}

func (n *nonDistributedJobManager) TriggerPullHistoricalTransactions(accountId, linkId uint64) (jobId string, err error) {
	panic("implement me")
}

func (n *nonDistributedJobManager) TriggerPullInitialTransactions(accountId, userId, linkId uint64) (jobId string, err error) {
	panic("implement me")
}

func (n *nonDistributedJobManager) TriggerPullLatestTransactions(accountId, linkId uint64, numberOfTransactions int64) (jobId string, err error) {
	panic("implement me")
}

func (n *nonDistributedJobManager) TriggerRemoveTransactions(accountId, linkId uint64, removedTransactions []string) (jobId string, err error) {
	panic("implement me")
}

func (n *nonDistributedJobManager) TriggerRemoveLink(accountId, userId, linkId uint64) (jobId string, err error) {
	log := n.log.WithFields(logrus.Fields{
		"accountId": accountId,
		"linkId":    linkId,
		"userId":    userId,
	})

	runner := &RemoveLinkJob{
		accountId: accountId,
		linkId:    linkId,
		userId:    userId,
		log:       log,
		db:        n.db,
		notify:    n.ps,
	}

	return fmt.Sprintf("%s:%X", RemoveLink, time.Now().Unix()), runner.Run(context.Background())
}

func (n *nonDistributedJobManager) Close() error {
	return nil
}
