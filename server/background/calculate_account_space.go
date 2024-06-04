package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/sirupsen/logrus"
)

type (
	CalculateAccountSpaceHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	CalculateAccountSpaceArguments struct {
		AccountId ID[Account] `json:"accountId"`
	}

	CalculateAccountSpaceJob struct {
		args  CalculateAccountSpaceArguments
		log   *logrus.Entry
		repo  repository.BaseRepository
		clock clock.Clock
	}
)

func (h CalculateAccountSpaceHandler) DefaultSchedule() string {
	return "0 0 0 * * 0" // At 00:00 on Sunday.
}

func (h *CalculateAccountSpaceHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := h.log.WithContext(ctx)

	log.Info("retrieving accounts to calculate space used")
	return nil
}
