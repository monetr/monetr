package schema

import "github.com/google/jsonschema-go/jsonschema"

func RRule(id string) *jsonschema.Schema {
	return &jsonschema.Schema{
		ID:          id,
		Type:        "string",
		Comment:     "RRule pattern for defining the frequency that something occurs",
		Description: "RRule pattern for defining the frequency that something occurs",
		Pattern:     `^DTSTART:\d{4}\d{2}\d{2}T\d+Z\nRRULE:((FREQ|INTERVAL|UNTIL|COUNT|BYDAY|BYMONTH|BYMONTHDAY|BYYEARDAY|BYWEEKNO|BYSETPOS|WKST)=([^;\\s]+))+`,
		Examples: []any{
			"DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			"DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1",
		},
	}
}
