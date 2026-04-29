package table

import (
	"context"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
	"github.com/pkg/errors"
)

// TODO FOR LATER: Row number per date must be in descending order. That is the
// row at the top of the file for a given date must be 0. this way as rows fall
// off the bottom of the file the row number for that date and transaction that
// remains is not affected. This will also need to be per amount
// OR Do it from the bottom up, but then after the first import you might need
// to throw away the oldest date as it could be partial data. So always go to
// the next date in the file to ensure uniqueness can be preserved.

// Mapping stores the actual mapping data that we will both persist to the
// database as well as receive from a client's request. Validating this is going
// to be a pain in the ass.
type Mapping struct {
	ID       IDSpec      `json:"id"`
	Amount   AmountSpec  `json:"amount"`
	Memo     FieldRef    `json:"memo"`
	Merchant *FieldRef   `json:"merchant,omitempty"`
	Date     DateSpec    `json:"date"`
	Posted   *PostedSpec `json:"posted,omitempty"`
	Balance  BalanceSpec `json:"balance"`
	Headers  []string    `json:"headers"`
}

func (m *Mapping) Validate(ctx context.Context) error {
	innerCtx := WithColumns(ctx, m.Headers)
	return errors.Wrapf(
		validation.ValidateStructWithContext(
			innerCtx,
			m,
			validation.Field(
				&m.ID,
				validators.By(func(ctx context.Context, field IDSpec) error {
					return field.Validate(ctx)
				}),
				validation.Required,
			),
			validation.Field(
				&m.Amount,
				validators.By(func(ctx context.Context, field AmountSpec) error {
					return field.Validate(ctx)
				}),
				validation.Required,
			),
			validation.Field(
				&m.Memo,
				validators.By(func(ctx context.Context, field FieldRef) error {
					return field.Validate(ctx)
				}),
				validation.Required,
			),
			validation.Field(
				&m.Merchant,
				validators.By(func(ctx context.Context, field *FieldRef) error {
					if field == nil {
						return nil
					}
					return field.Validate(ctx)
				}),
			),
			validation.Field(
				&m.Date,
				validators.By(func(ctx context.Context, field DateSpec) error {
					return field.Validate(ctx)
				}),
				validation.Required,
			),
			validation.Field(
				&m.Posted,
				validators.By(func(ctx context.Context, field *PostedSpec) error {
					if field == nil {
						return nil
					}
					return field.Validate(ctx)
				}),
			),
			validation.Field(
				&m.Balance,
				validators.By(func(ctx context.Context, field BalanceSpec) error {
					return field.Validate(ctx)
				}),
				validation.Required,
			),
			validation.Field(
				&m.Headers,
				validation.Length(1, 20),
				validators.Unique[string](),
				validation.Each(
					is.PrintableASCII,
					validation.Length(1, 100),
					validation.Required,
				),
				validation.Required,
			),
		),
		"failed to validate %T",
		m,
	)
}

func WithColumns(ctx context.Context, columns []string) context.Context {
	return context.WithValue(ctx, FieldRef{}, columns)
}

func getColumns(ctx context.Context) []string {
	columns, ok := ctx.Value(FieldRef{}).([]string)
	if ok {
		return columns
	}

	return []string{}
}
