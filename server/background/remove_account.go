package background

import (
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/sirupsen/logrus"
)

const (
	RemoveAccount = "RemoveAccount"
)

type (
	RemoveAccountHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		configuration config.Configuration
	}
)
