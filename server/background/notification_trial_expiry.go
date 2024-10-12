package background

import (
	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	. "github.com/monetr/monetr/server/models"
	"github.com/sirupsen/logrus"
)

const (
	NotificationTrialExpiry = "NotificationTrialExpiry"
)

type (
	NotificationTrialExpiryHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		clock        clock.Clock
		enqueuer     JobEnqueuer
		unmarshaller JobUnmarshaller
	}

	NotificationTrialExpiryArguments struct {
		AccountId ID[Account]
	}
)
