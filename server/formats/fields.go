package formats

type Field uint32

const (
	FieldIgnore          Field = 0
	FieldUniqueId        Field = 1
	FieldName            Field = 2
	FieldDescription     Field = 3
	FieldNotes           Field = 4
	FieldDate            Field = 5
	FieldPending         Field = 6
	FieldStatus          Field = 7
	FieldCategory        Field = 8
	FieldCurrencyCodeISO Field = 9
	FieldAmountCombined  Field = 10
	FieldAmountDebit     Field = 11
	FieldAmountCredit    Field = 12
)
