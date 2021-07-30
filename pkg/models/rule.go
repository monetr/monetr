package models

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/go-pg/pg/v10/types"
	"github.com/nleeper/goment"
	"github.com/teambition/rrule-go"
)

var (
	_ types.ValueAppender = &Rule{}
	_ types.ValueScanner  = &Rule{}
	_ json.Marshaler      = &Rule{}
	_ json.Unmarshaler    = &Rule{}
)

type Rule struct {
	rrule.RRule
}

func NewRule(input string) (*Rule, error) {
	rule, err := rrule.StrToRRule(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse rule")
	}

	return &Rule{
		RRule: *rule,
	}, nil
}

func (r *Rule) UnmarshalJSON(input []byte) error {
	rule, err := rrule.StrToRRule(strings.Trim(string(input), `"`))
	if err != nil {
		return errors.Wrap(err, "failed to parse rule")
	}

	moment, _ := goment.New()
	now := moment.Local().StartOf("day").ToTime()
	rule.DTStart(now)

	r.RRule = *rule
	return nil
}

func (r *Rule) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, r.RRule.OrigOptions.RRuleString())), nil
}

func (r *Rule) AppendValue(b []byte, flags int) ([]byte, error) {
	if flags == 1 {
		b = append(b, '\'')
	}
	b = append(b, r.RRule.OrigOptions.RRuleString()...)
	if flags == 1 {
		b = append(b, '\'')
	}
	return b, nil
}

func (r *Rule) ScanValue(rd types.Reader, n int) error {
	if n <= 0 {
		return nil
	}

	tmp, err := rd.ReadFullTemp()
	if err != nil {
		return err
	}

	rule, err := rrule.StrToRRule(string(tmp))
	if err != nil {
		return err
	}

	moment, _ := goment.New()
	now := moment.Local().StartOf("day").ToTime()
	rule.DTStart(now)

	r.RRule = *rule

	return nil
}
