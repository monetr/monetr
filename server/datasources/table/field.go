package table

import (
	"context"
	"fmt"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

type DerivedKind string

const (
	// DerivedKindRowNumber is used to generate a 0 indexed row number starting
	// from the bottom of the file to the top of the file. This row number isn't
	// exactly stable since new rows are added and old rows are eventually
	// removed.
	DerivedKindRowNumber DerivedKind = "rowNumber"
	// // DerivedKindRowNumberPerDay is the 0 indexed count of rows per date value.
	// // This is from the bottom to the top of the file.
	// DerivedKindRowNumberPerDay DerivedKind = "rowNumberPerDay"
	// // DerivedKindRowNumberPerDayPerAmount is the 0 indexed count of rows per date
	// // value per amount value. This way if days have all unique transaction
	// // amounts then deduplication is easy. If days have duplicate transaction
	// // amounts then they are deduplicated in order they are observed.
	// DerivedKindRowNumberPerDayPerAmount DerivedKind = "rowNumberPerDayPerAmount"
)

// FieldRef references a field within the document, or it references a derived
// field. Not both. Derived fields are generated at processing time and can't be
// previewed.
type FieldRef struct {
	// Name just tells us the name of the field to derive the value from. This can
	// be blank if we are deriving a value instead.
	Name string `json:"name,omitempty"`
	// DerivedKind specifies that this value should come from a calculated field
	// instead of from a named field.
	DerivedKind DerivedKind `json:"derivedKind,omitempty"`
}

func (s *FieldRef) Validate(ctx context.Context) error {
	return validators.OneOfStruct(
		ctx,
		s,
		[]*validation.FieldRules{
			validation.Field(
				&s.Name,
				validators.In(getColumns(ctx)...),
				validators.PrintableUnicode,
				validation.Required,
			),
			validation.Field(
				&s.DerivedKind,
				validation.Empty,
			),
		},
		[]*validation.FieldRules{
			validation.Field(
				&s.Name,
				validation.Empty,
			),
			validation.Field(
				&s.DerivedKind,
				validators.In(
					DerivedKindRowNumber,
					// DerivedKindRowNumberPerDay,
					// DerivedKindRowNumberPerDayPerAmount,
				),
				validation.Required,
			),
		},
	)
}

func (s FieldRef) String() string {
	if s.DerivedKind != "" {
		return fmt.Sprintf("derived::%s", s.DerivedKind)
	}

	return s.Name
}
