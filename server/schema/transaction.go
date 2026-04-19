package schema

import (
	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

var (
	CreateManualTransaction = z.Struct(z.Shape{
		"spendingId":   z.Ptr(ID[models.Spending]().Optional()),
		"name":         Name().Required(),
		"merchantName": Name().Optional(),
		"date":         Date().Required(),
		"amount":       z.Int64().Required().Not().EQ(0),
		"isPending":    z.Bool().Default(false).Required(),
	})

	AdjustsBalance = z.Struct(z.Shape{
		// Meta fields, these fields affect things besides the transaction when the
		// transaction is created.
		"adjustsBalance": z.Bool().Default(false).Required(),
	})

	// PatchTransaction is the schema of fields that can be updated on any
	// transaction in monetr, this is a safe subset of fields that applies to any
	// datasource monetr supports.
	PatchTransaction = z.Struct(z.Shape{
		"spendingId": z.Ptr(ID[models.Spending]().Optional()),
		"name":       Name().Optional(),
	})

	// PatchManualTransaction is the schema of fields that can be updated for
	// manual links. This allows more fields to be changed after the creation of
	// the transaction in order to allow for easier budgeting.
	PatchManualTransaction = z.Struct(z.Shape{
		"spendingId":   z.Ptr(ID[models.Spending]().Optional()),
		"name":         Name().Optional(),
		"merchantName": Name().Optional(),
		"date":         Date().Optional(),
		"amount":       z.Int64().Optional().Not().EQ(0),
		"isPending":    z.Bool().Optional(),
	})
)
