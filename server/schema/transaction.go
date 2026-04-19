package schema

import (
	"github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

var (
	CreateManualTransaction = zog.Struct(zog.Shape{
		"spendingId":   zog.Ptr(ID[models.Spending]().Optional()),
		"name":         Name().Required(),
		"merchantName": Name().Optional(),
		"date":         Date().Required(),
		"amount":       zog.Int64().Required().Not().EQ(0),
		"isPending":    zog.Bool().Default(false).Required(),
	})

	AdjustsBalance = zog.Struct(zog.Shape{
		// Meta fields, these fields affect things besides the transaction when the
		// transaction is created.
		"adjustsBalance": zog.Bool().Default(false).Required(),
	})

	// PatchTransaction is the schema of fields that can be updated on any
	// transaction in monetr, this is a safe subset of fields that applies to any
	// datasource monetr supports.
	PatchTransaction = zog.Struct(zog.Shape{
		"spendingId": zog.Ptr(ID[models.Spending]().Optional()),
		"name":       Name().Optional(),
	})

	// PatchManualTransaction is the schema of fields that can be updated for
	// manual links. This allows more fields to be changed after the creation of
	// the transaction in order to allow for easier budgeting.
	PatchManualTransaction = zog.Struct(zog.Shape{
		"spendingId":   zog.Ptr(ID[models.Spending]().Optional()),
		"name":         Name().Optional(),
		"merchantName": Name().Optional(),
		"date":         Date().Optional(),
		"amount":       zog.Int64().Optional().Not().EQ(0),
		"isPending":    zog.Bool().Optional(),
	})
)

func isValidTransaction(val any, ctx zog.Ctx) bool {
	switch val := val.(type) {
	case *models.Transaction:
		if val.IsAddition() && val.SpendingId != nil {
			ctx.AddIssue(ctx.Issue().
				SetCode("invalid_spending").
				SetPath([]string{"spendingId"}).
				SetMessage("cannot spend on a deposit").
				SetParams(map[string]any{
					"spendingId": val.SpendingId,
					"amount":     val.Amount,
				}),
			)
			return false
		}
	}

	return true
}
