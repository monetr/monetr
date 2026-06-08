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
			is.PrintableUnicode.Error("Name must be printable unicode"),
			validation.Length(1, 300).Error("Name cannot be longer than 300 characters"),
			validation.Required.Error("Name is required"),
		),
		validation.Key("merchantName",
			is.PrintableUnicode.Error("Merchant name must be printable unicode"),
			validation.Length(0, 300).Error("Merchant name cannot be longer than 300 characters"),
		).Required(false),
		validation.Key("date",
			validation.Date(time.RFC3339).Error("Date must be in a valid format"),
			validation.Required.Error("Date is required"),
		),
		validation.Key("amount",
			Amount(),
			validation.NotEq(0).Error("Amount cannot be 0"),
			validation.Required.Error("Amount is required"),
		),
		validation.Key("spendingId",
			validation.OneOf(
				ValidID[models.Spending](),
				validation.Nil,
			),
		).Required(false),
		validation.Key("adjustsBalance",
			validation.OneOf(
				validation.In(true, false).Error("Adjusts balance must be a valid boolean if specified"),
				validation.Never.Error("Validation must be not be specified or must be a valid boolean"),
			),
		).Required(false),
		validation.Key("isPending",
			validation.In(true, false).Error("Is pending must be a valid boolean if provided"),
		).Required(false),
	)
)
