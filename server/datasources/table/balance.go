package table

import (
	"context"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

type BalanceKind string

const (
	// BalanceKindNone means that balance will not be adjusted at all as part of
	// this import.
	BalanceKindNone BalanceKind = "none"
	// BalanceKindField means that the balance will be derived from the most
	// recent field in the import.
	BalanceKindField BalanceKind = "field"
	// BalanceKindSum means the balance will be adjusted based on the sum of the
	// transactions being added. Transactions that already exist will not count
	// towards the balance adjustment.
	BalanceKindSum BalanceKind = "sum"
)

// BalanceSpec determines how we will figure out balance adjustments for the
// imported file. Potentially a no-op depending on [BalanceSpec.Kind].
type BalanceSpec struct {
	// Kind determines how balance will be treated for this document.
	Kind BalanceKind `json:"kind"`
	// Field is the name of the balance field if we are deriving a balance from a
	// specific field.
	Fields []FieldRef `json:"fields,omitempty"`
}

func (s *BalanceSpec) Validate(ctx context.Context) error {
	return validators.OneOfStruct(
		ctx,
		s,
		// If [BalanceSpec.Kind] is [BalanceKindNone] or [BalanceKindSum] then
		// don't allow a field to be provided.
		[]*validation.FieldRules{
			validation.Field(
				&s.Kind,
				validators.In(
					BalanceKindNone,
					BalanceKindSum,
				),
				validation.Required,
			),
			validation.Field(
				&s.Fields,
				validation.Empty,
			),
		},
		// If [BalanceSpec.Kind] is [BalanceKindField] then require that a field
		// is specified.
		[]*validation.FieldRules{
			validation.Field(
				&s.Kind,
				validators.Eq(BalanceKindField),
				validation.Required,
			),
			validation.Field(
				&s.Fields,
				validation.Each(
					validators.By(func(ctx context.Context, field FieldRef) error {
						return field.Validate(ctx)
					}),
				),
				validation.Length(1, 1),
				validators.Unique[FieldRef](),
				validation.Required,
			),
		},
	)
}
