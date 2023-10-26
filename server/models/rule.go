package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/go-pg/pg/v10/types"
	"github.com/nleeper/goment"
	"github.com/teambition/rrule-go"
)

var (
	_ types.ValueAppender = &RuleSet{}
	_ types.ValueScanner  = &RuleSet{}
	_ json.Marshaler      = &RuleSet{}
	_ json.Unmarshaler    = &RuleSet{}
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

// AppendValue implements types.ValueAppender.
func (r *RuleSet) AppendValue(b []byte, flags int) ([]byte, error) {
	if flags == 1 {
		b = append(b, '\'')
	}
	b = append(b, []byte(r.Set.String())...)
	if flags == 1 {
		b = append(b, '\'')
	}
	return b, nil
}

// ScanValue implements types.ValueScanner.
func (r *RuleSet) ScanValue(rd types.Reader, n int) error {
	if n <= 0 {
		return nil
	}

	tmp, err := rd.ReadFullTemp()
	if err != nil {
		return err
	}

	set, err := rrule.StrToRRuleSet(string(tmp))
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
	// This will make sure that we can work with times that are in the past. I started working on monetr in february
	// 2021 I think. So really there shouldn't be anything before that time.
	// TODO This is actually going to cause bugs, the DTStart should be specified
	//      by the client, not by some hardcoded date.
	rule.DTStart(time.Date(2021, 2, 14, 0, 0, 0, 0, time.UTC))

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

func (r *Rule) String() string {
	return fmt.Sprintf("%s <%s>", r.RRule.After(time.Now(), false), r.RRule.String())
}
