package schema

import "github.com/google/jsonschema-go/jsonschema"

func Timestamp(id string) *jsonschema.Schema {
	return &jsonschema.Schema{
		ID:      id,
		Type:    "string",
		Format:  "date",
		Pattern: `^(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d\.\d+([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))$`,
	}
}
