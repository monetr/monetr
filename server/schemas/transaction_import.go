package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

var (
	PatchTransactionImport = validation.OneOf(
		// When we do not yet have a mapping then they must specify a new mapping ID
		// and say that we are moving to the pending preview status.
		validation.Map(
			validation.Key(
				"transactionImportMappingId",
				ValidID[models.TransactionImportMapping](),
				validation.Required,
			),
			validation.Key(
				"status",
				validators.In(string(models.TransactionImportStatusPendingPreview)),
				validation.Required,
			),
		),
		// Otherwise the user can only progress the import to pending processing.
		validation.Map(
			validation.Key(
				"status",
				validators.In(string(models.TransactionImportStatusPendingProcessing)),
				validation.Required,
			),
		),
	)
)
