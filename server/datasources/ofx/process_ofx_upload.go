package ofx

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
)

type ProcessOFXUploadArguments struct {
	AccountId           models.ID[models.Account]           `json:"accountId"`
	BankAccountId       models.ID[models.BankAccount]       `json:"bankAccountId"`
	TransactionUploadId models.ID[models.TransactionUpload] `json:"transactionUploadId"`
}

func ProcessOFXUpload(ctx queue.Context, args ProcessOFXUploadArguments) error {
	return nil
}
