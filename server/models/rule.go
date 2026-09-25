package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"
)

var (
	_ driver.Valuer    = &RuleSet{}
	_ sql.Scanner      = &RuleSet{}
	_ json.Marshaler   = &RuleSet{}
	_ json.Unmarshaler = &RuleSet{}
)

type RuleSet struct {
	rrule.Set
}

func NewRuleSet(input string) (*RuleSet, error) {
	set, err := rrule.StrToRRuleSet(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ruleset")
	}

	return &RuleSet{
		Set: *set,
	}, nil
}

// Value implements driver.Valuer.
func (r *RuleSet) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return r.Set.String(), nil
}

// Scan implements sql.Scanner.
func (r *RuleSet) Scan(src any) error {
	var input string
	switch value := src.(type) {
	case nil:
		return nil
	case string:
		input = value
	case []byte:
		input = string(value)
	default:
		return errors.Errorf("cannot scan %T into a RuleSet", src)
	}
	if input == "" {
		return nil
	}

	set, err := rrule.StrToRRuleSet(input)
	if err != nil {
		return errors.Wrap(err, "failed to parse ruleset")
	}

	r.Set = *set
	return nil
}

// MarshalJSON implements json.Marshaler.
func (r *RuleSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *RuleSet) UnmarshalJSON(input []byte) error {
	inputStr := string(input)
	// Need to remove leading and trailing double quotes too.
	inputStr = strings.Trim(inputStr, `"`)
	// Need to do this because the `\n` comes escaped in JSON but the parser need the real one.
	inputStr = strings.ReplaceAll(inputStr, `\n`, "\n")
	set, err := rrule.StrToRRuleSet(inputStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse ruleset")
	}

	r.Set = *set
	return nil
}

// Clone will make a complete copy of the rule set, this is necessary because
// even though the top level rule set object is not typically a pointer. The
// ruleset object itself contains pointers and these cannot be controlled. As a
// result; simply dereferencing and passing a pointer to the new "object" of the
// ruleset can still cause some odd back-propagation. Like if DTSTART is
// changed, that change will cascade back up. This clone will prevent that by
// creating a copy from the string representation of the rule.
func (r *RuleSet) Clone() *RuleSet {
	ruleset, err := rrule.StrToRRuleSet(r.String())
	if err != nil {
		panic(fmt.Sprintf("failed to clone rule set! %+v", err))
	}

	return &RuleSet{*ruleset}
}
