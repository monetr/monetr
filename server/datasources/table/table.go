package table

import (
	"context"
	"encoding/base32"
	"encoding/csv"
	"hash/fnv"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/monetr/monetr/server/currency"
	"github.com/pkg/errors"
)

type Row struct {
	RowNumber int       `json:"rowNumber"`
	ID        string    `json:"id"`
	Amount    int64     `json:"amount"`
	Memo      string    `json:"memo"`
	Merchant  *string   `json:"merchant,omitempty"`
	Date      time.Time `json:"date"`
	Posted    bool      `json:"posted"`
	Balance   int64     `json:"balance"`
}

var (
	_ TableReader = &csv.Reader{}
)

type TableReader interface {
	Read() (record []string, err error)
}

type Table struct {
	currency        string
	timezone        *time.Location
	index           int
	firstRowHeaders bool
	headers         []string
	mapping         *Mapping
	reader          TableReader
}

func NewTable(
	reader TableReader,
	mapping *Mapping,
	firstRowHeaders bool,
) *Table {
	return &Table{
		currency:        "USD",    // TODO!!!
		timezone:        time.UTC, // TODO!!!
		firstRowHeaders: firstRowHeaders,
		reader:          reader,
		mapping:         mapping,
	}
}

func (t *Table) Read() (*Row, error) {
	if t.firstRowHeaders && t.index == 0 {
		if err := t.mapping.Validate(context.TODO()); err != nil {
			return nil, err
		}

		data, err := t.reader.Read()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		t.headers = data
		if !slices.Equal(t.headers, t.mapping.Headers) {
			return nil, errors.New("headers in file do not match headers in mapping!")
		}
		t.index++
	} else if !t.firstRowHeaders && len(t.headers) == 0 {
		if err := t.mapping.Validate(context.TODO()); err != nil {
			return nil, err
		}

		t.headers = t.mapping.Headers
	}

	defer func() {
		t.index++
	}()

	data, err := t.reader.Read()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	id, err := t.getID(data)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to derive ID from table for row %d",
			t.index,
		)
	}

	amount, err := t.getAmount(data)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to derive amount from table for row %d",
			t.index,
		)
	}

	date, err := t.getDate(data)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to derive date from table for row %d",
			t.index,
		)
	}

	balance, err := t.getBalance(data)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to derive balance from table for row %d",
			t.index,
		)
	}

	row := Row{
		RowNumber: t.index,
		ID:        id,
		Amount:    amount,
		Memo:      t.getField(data, t.mapping.Memo),
		Merchant:  t.getMerchant(data),
		Date:      date,
		Posted:    t.getPosted(data),
		Balance:   balance,
	}

	return &row, nil
}

func (t *Table) getID(data []string) (string, error) {
	switch t.mapping.ID.Kind {
	case IDSpecKindNative:
		values := make([]string, len(t.mapping.ID.Fields))
		for i, field := range t.mapping.ID.Fields {
			values[i] = t.getField(data, field)
		}
		return strings.Join(values, "::"), nil
	case IDSpecKindHashed:
		hash := fnv.New64()
		for _, field := range t.mapping.ID.Fields {
			_, err := hash.Write([]byte(t.getField(data, field)))
			if err != nil {
				return "", errors.WithStack(err)
			}
		}

		return base32.StdEncoding.EncodeToString(hash.Sum(nil)), nil
	default:
		panic("invalid id spec")
	}
}

func (t *Table) getAmount(data []string) (int64, error) {
	var amount int64
	var err error
	switch t.mapping.Amount.Kind {
	case AmountKindSign:
		// Assumes that the mapping struct is 100% valid!
		field := t.mapping.Amount.Fields[0]
		value := t.getField(data, field)
		amount, err = currency.ParseFriendlyToAmount(value, t.currency)
	case AmountKindType:
		value := t.getField(data, t.mapping.Amount.Fields[0])
		direction := t.getField(data, t.mapping.Amount.Fields[1])
		// err is handled, just not right here!
		amount, err = currency.ParseFriendlyToAmount(value, t.currency)

		switch direction {
		case t.mapping.Amount.Credit:
			// Amounts should be negative if they are credits, this might need to be
			// reversed but I need real data to play with first.
			amount *= -1
		case t.mapping.Amount.Debit:
			// No-op?
		default:
			panic("invalid amount direction mapping")
		}
	case AmountKindColumn:
		debit := t.getField(data, t.mapping.Amount.Fields[0])
		credit := t.getField(data, t.mapping.Amount.Fields[1])

		switch {
		case debit != "" && credit == "":
			amount, err = currency.ParseFriendlyToAmount(debit, t.currency)
		case debit == "" && credit != "":
			amount, err = currency.ParseFriendlyToAmount(credit, t.currency)
		case debit == "" && credit == "":
			return 0, errors.Errorf("invalid amount row %d, both debit and credit columns are empty", t.index)
		default:
			return 0, errors.Errorf("invalid amount row %d, both debit and credit columns are populated", t.index)
		}
	default:
		panic("invalid amount spec")
	}

	// handle the error from the parse amount.
	if err != nil {
		return 0, err
	}

	// If the user wants to invert the value then do that here before we return
	// it. Do this after everything!
	if t.mapping.Amount.Invert {
		amount *= -1
	}

	return amount, nil
}

func (t *Table) getMerchant(data []string) *string {
	if t.mapping.Merchant != nil {
		// TODO Simplify to new(t.getField(data, *t.mapping.Merchant)) once golang
		// pulls its head out of its ass.
		value := t.getField(data, *t.mapping.Merchant)
		return &value
	}

	return nil
}

func (t *Table) getDate(data []string) (time.Time, error) {
	value := t.getField(data, t.mapping.Date.Fields[0])
	format := t.mapping.Date.GetTimeFormat()
	date, err := time.ParseInLocation(format, value, t.timezone)
	if err != nil {
		return date, errors.WithStack(err)
	}

	return date, nil
}

func (t *Table) getPosted(data []string) bool {
	// If we don't have a posted spec then we always assume that the transaction
	// is infact posted and not pending. This seems to be the default behavior
	// with all of the CSV exports I've been able to find with banks?
	if t.mapping.Posted == nil {
		return true
	}
	value := t.getField(data, t.mapping.Posted.Fields[0])
	return value == t.mapping.Posted.Posted
}

func (t *Table) getBalance(data []string) (int64, error) {
	switch t.mapping.Balance.Kind {
	case BalanceKindNone:
		return 0, nil
	case BalanceKindField:
		// Assumes that the mapping struct is 100% valid!
		field := t.mapping.Balance.Fields[0]
		value := t.getField(data, field)
		amount, err := currency.ParseFriendlyToAmount(value, t.currency)
		if err != nil {
			return 0, err
		}
		return amount, nil
	case BalanceKindSum:
		// TODO I'm not sure how to implement this and it might need to be a post
		// processing stage?
		// Essentially what I would need to do this is a balance at a specific point
		// in time. Like end of day balance of a specific date, and then take the sum
		// of all the rows that have come after that date to calculate the current
		// running balance. But none of that exists yet so this might just be a no-op
		// for the mean time.
		return 0, nil
	default:
		panic("invalid balance spec")
	}
}

func (t *Table) getField(data []string, field FieldRef) string {
	switch field.DerivedKind {
	case DerivedKindRowNumber:
		// TODO IS this the right order? This would be top of file = 0
		return strconv.Itoa(t.index)
	default:
		index := slices.Index(t.headers, field.Name)
		return strings.TrimSpace(data[index])
	}
}
