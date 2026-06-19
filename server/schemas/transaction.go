package schemas

import (
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

var (
	CreateTransactionSchema = validation.Map(
		validation.Key("name",
			Name(),
		).Required(Require),
		validation.Key("merchantName",
			TextField(),
		).Required(Optional),
		validation.Key("date",
			validation.Date(time.RFC3339).Error("Date must be in a valid format"),
			validation.Required.Error("Date is required"),
		).Required(Require),
		validation.Key("amount",
			Amount(),
			validation.NotEq(0).Error("Amount cannot be 0"),
			validation.Required.Error("Amount is required"),
		).Required(Require),
		validation.Key("spendingId",
			validation.OneOf(
				validation.Nil.Error("must be nil"),
				ValidID[models.Spending](),
			),
		).Required(Optional),
		validation.Key("adjustsBalance",
			validation.OneOf(
				validation.In(true, false).Error("Adjusts balance must be a valid boolean if specified"),
				validation.Never.Error("Validation must be not be specified or must be a valid boolean"),
			),
		).Required(Optional),
		validation.Key("isPending",
			is.Boolean,
			validation.In(true, false).Error("Is pending must be a valid boolean if provided"),
		).Required(Optional),
	)

	PatchTransaction = validation.Map(
		validation.Key("name",
			Name(),
		).Required(Optional),
		// Even on a non-manual link we still let the user clean up the merchant
		// name for display. The original merchant name from Plaid is preserved
		// separately so this is purely cosmetic and safe to allow here.
		validation.Key("merchantName",
			TextField(),
		).Required(Optional),
		validation.Key("spendingId",
			validation.OneOf(
				validation.Nil.Error("must be nil"),
				ValidID[models.Spending](),
			),
		).Required(Optional),
	)

	PatchManualTransaction = validation.Map(
		validation.Key("name",
			Name(),
		).Required(Optional),
		validation.Key("merchantName",
			TextField(),
		).Required(Optional),
		validation.Key("date",
			is.String,
			validation.Date(time.RFC3339).Error("Date must be in a valid format"),
			validation.Required.Error("Date is required"),
		).Required(Optional),
		validation.Key("amount",
			Amount(),
			validation.Required.Error("Amount is required"),
		).Required(Optional),
		validation.Key("spendingId",
			validation.OneOf(
				validation.Nil.Error("must be nil"),
				ValidID[models.Spending](),
			),
		).Required(Optional),
		validation.Key("isPending",
			// Do NOT use validation.Required here. is pending is a boolean and false
			// is its zero value, so Required would reject a perfectly valid attempt
			// to set is pending back to false.
			is.Boolean,
			validation.In(true, false).Error("Is pending must be a valid boolean if provided"),
		).Required(Optional),

		// NOTE adjustsBalance is intentionally not accepted here. monetr does not
		// yet recalculate balances when a manual transaction's amount changes (see
		// the TODO in the controller), and it is not a field on the transaction
		// model so it would fail the merge anyway. Until balance adjustment is
		// implemented we reject it rather than silently accept something we do not
		// honor.
	)
)
