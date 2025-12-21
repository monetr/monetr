package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/sirupsen/logrus"
)

const (
	RemoveAccount = "RemoveAccount"
)

func TriggerRemoveAccount(
	ctx context.Context,
	backgroundJobs JobController,
	arguments SyncPlaidArguments,
) error {
	if arguments.Trigger == "" {
		arguments.Trigger = "manual"
	}
	return backgroundJobs.EnqueueJob(ctx, SyncPlaid, arguments)
}

type (
	RemoveAccountHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		configuration config.Configuration
		kms           secrets.KeyManagement
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	RemoveAccountArguments struct {
		AccountId ID[Account] `json:"accountId"`
	}

	RemoveAccountJob struct {
		args    RemoveAccountArguments
		log     *logrus.Entry
		repo    repository.BaseRepository
		secrets repository.SecretsRepository
		clock   clock.Clock
	}
)
