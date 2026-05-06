package table

import (
	"context"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

type AmountKind string

const (
	// AmountKindSign means that the single amount column can be used to derive
	// the sign of the amount. If it is negative then its a debit, positive means
	// credit.
	AmountKindSign AmountKind = "sign"
	// AmountKindType means that there is a column that contains some values, such
	// as DEBIT or CREDIT which we must map off of.
	AmountKindType AmountKind = "type"
	// AmountKindColumn means there are two separate columns to derive the amount
	// from, one column for debits, one for credits.
	AmountKindColumn AmountKind = "column"
)

// AmountSpec derives the amount of the transaction from the document row. It
// allows for the amount to be specified as a standalone column in the first
// value of [AmountSpec.Fields], or specified as a pair of columns
// ([AmountSpec.Fields]). The pair of columns is either an amount and a
// direction, or each column represents a specific direction independently.
// monetr treats amounts as inverted, deposits are negative and debits are
// positive. So the [AmountSpec.Invert] field is not there to adjust for
// monetr's representation. The invert field is there to adjust only the input
// representation.
type AmountSpec struct {
	Kind AmountKind `json:"kind"`
	// Invert the sign after parsing, for example; negative values become
	// positive.
	Invert bool `json:"invert"`

	// If it is [AmountKindSign] then this must be a single value.
	// If it is [AmountKindType] then the first field here is the amount, the
	// second is the type of amount.
	// If it is [AmountKindColumn] then the first field is debit and the second
	// field is credit.
	Fields []FieldRef `json:"fields,omitempty"`

	// Credit is the value we look for (case-insensitively) in the second field
	// when we are [AmountKindType].
	Credit string `json:"credit,omitempty"`
	// Debit is the value we look for (case-insensitively) in the second field
	// when we are [AmountKindType].
	Debit string `json:"debit,omitempty"`
}

func (s *AmountSpec) Validate(ctx context.Context) error {
	return validators.OneOfStruct(
		ctx,
		s,
		// When we are [AmountKindSign] then only [AmountSpec.Fields] should be
		// provided with exactly one item. [AmountSpec.Credit] and
		// [AmountSpec.Debit] must be empty.
		[]*validation.FieldRules{
			validation.Field(
				&s.Kind,
				validators.Eq(AmountKindSign),
				validation.Required,
			),
			validation.Field(
				&s.Invert,
				validation.In(true, false),
			),
			validation.Field(
				&s.Fields,
				validation.Each(
					validators.By(func(ctx context.Context, field FieldRef) error {
						return field.Validate(ctx)
					}),
				),
				validation.Required,
				validation.Length(1, 1),
				validators.Unique[FieldRef](),
			),
			validation.Field(
				&s.Credit,
				validation.Empty.Error(`when kind is "sign" credit cannot be specified`),
			),
			validation.Field(
				&s.Debit,
				validation.Empty.Error(`when kind is "sign" debit cannot be specified`),
			),
		},
		// When we are [AmountKindType] then [AmountSpec.Credit] and
		// [AmountSpec.Debit] must both be provided, and [AmountSpec.Fields] must
		// have exactly two items.
		[]*validation.FieldRules{
			validation.Field(
				&s.Kind,
				validators.Eq(AmountKindType),
				validation.Required,
			),
			validation.Field(
				&s.Invert,
				validation.In(true, false),
			),
			validation.Field(
				&s.Fields,
				validation.Each(
					validators.By(func(ctx context.Context, field FieldRef) error {
						return field.Validate(ctx)
					}),
				),
				validation.Required,
				validation.Length(2, 2),
				validators.Unique[FieldRef](),
			),
			validation.Field(
				&s.Credit,
				validation.Required,
				validators.PrintableUnicode,
				validation.NotIn(s.Debit),
				validation.Length(1, 100),
			),
			validation.Field(
				&s.Debit,
				validation.Required,
				validators.PrintableUnicode,
				validation.NotIn(s.Credit),
				validation.Length(1, 100),
			),
		},
		// When we are [AmountKindColumn] then The [AmountSpec.Credit] and
		// [AmountSpec.Debit] fields must be empty, and [AmountSpec.Fields] must
		// have exactly two unique items.
		[]*validation.FieldRules{
			validation.Field(
				&s.Kind,
				validators.Eq(AmountKindColumn),
				validation.Required,
			),
			validation.Field(
				&s.Invert,
				validation.In(true, false),
			),
			validation.Field(
				&s.Fields,
				validation.Each(
					validators.By(func(ctx context.Context, field FieldRef) error {
						return field.Validate(ctx)
					}),
				),
				validation.Required,
				validation.Length(2, 2),
				validators.Unique[FieldRef](),
			),
			validation.Field(
				&s.Credit,
				validation.Empty.Error(`when kind is "column" credit cannot be specified`),
			),
			validation.Field(
				&s.Debit,
				validation.Empty.Error(`when kind is "column" debit cannot be specified`),
			),
		},
	)
}
