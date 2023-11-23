package recurring

import (
	"github.com/monetr/monetr/server/models"
)

type TransactionReader interface {
	Read() (*models.Transaction, error)
	Close() error
}
