package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-pg/pg/v10/types"
	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
)

var (
	_ types.ValueAppender = &PhoneNumber{}
	_ types.ValueScanner  = &PhoneNumber{}
	_ json.Marshaler      = &PhoneNumber{}
	_ json.Unmarshaler    = &PhoneNumber{}
)

type PhoneNumber struct {
	number phonenumbers.PhoneNumber
}

func (p *PhoneNumber) E164() string {
	return phonenumbers.Format(&p.number, phonenumbers.E164)
}

func (p *PhoneNumber) UnmarshalJSON(bytes []byte) error {
	str := strings.Trim(string(bytes), `"`)
	number, err := phonenumbers.Parse(str, "US")
	if err != nil {
		return errors.Wrap(err, "failed to parse phone number")
	}

	*p = PhoneNumber{
		number: *number,
	}
	return nil
}

func (p *PhoneNumber) MarshalJSON() ([]byte, error) {
	format := phonenumbers.Format(&p.number, phonenumbers.NATIONAL)
	return []byte(fmt.Sprintf(`"%s"`, format)), nil
}

func (p *PhoneNumber) AppendValue(b []byte, flags int) ([]byte, error) {
	if flags == 1 {
		b = append(b, '\'')
	}
	format := phonenumbers.Format(&p.number, phonenumbers.NATIONAL)
	b = append(b, format...)
	if flags == 1 {
		b = append(b, '\'')
	}
	return b, nil
}

func (p *PhoneNumber) ScanValue(rd types.Reader, n int) error {
	if n <= 0 {
		return nil
	}

	tmp, err := rd.ReadFullTemp()
	if err != nil {
		return err
	}

	number, err := phonenumbers.Parse(string(tmp), "US")
	if err != nil {
		return errors.Wrap(err, "failed to parse phone number")
	}

	*p = PhoneNumber{
		number: *number,
	}

	return nil
}
