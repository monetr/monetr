package schema

import (
	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

var CreateTransactionSchema = z.Struct(z.Shape{
	"bankAccountId":  ID[models.BankAccount](z.IssuePath([]string{"bankAccountId"})).Required(),
	"spendingId":     ID[models.Spending](z.IssuePath([]string{"spendingId"})).Optional(),
	"spendingAmount": z.Int64().GTE(0).Optional(),
	"amount":         z.Int64().Required(),
	"date":           z.Time().Required(),
	"name":           Name(),
	"isPending":      z.Bool().Default(false).Optional(),
})

// PatchManualTransactionSchema is the part of the schema that is considered
// valid when updating transactions for manual links. It can be merged with
// [PatchTransactionSchema] to build a complete schema.
var PatchManualTransactionSchema = z.Struct(z.Shape{
	"amount":    z.Int64().Optional(),
	"date":      z.Time().Optional(),
	"isPending": z.Bool().Default(false).Optional(),
})

// PatchTransactionSchema is the schema used for validating transaction updates.
// This is valid for any kind of transaction.
var PatchTransactionSchema = z.Struct(z.Shape{
	"spendingId":     ID[models.Spending](z.IssuePath([]string{"spendingId"})).Optional(),
	"spendingAmount": z.Int64().GTE(0).Optional(),
	"name":           Name().Optional(),
})
