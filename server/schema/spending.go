package schema

import (
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
)

var (
	CreateSpending = myownsanity.MUST((&jsonschema.Schema{
		ID:          "Spending Creation",
		Title:       "Spending",
		Description: "Create a Spending object",
		Comment:     "Create a Spending object",
		Type:        "object",
		Properties: map[string]*jsonschema.Schema{
			"fundingScheduleId": ID[models.FundingSchedule]("fundingScheduleId"),
			"name": {
				ID:        "name",
				Comment:   "The name of the Spending object",
				Type:      "string",
				MaxLength: myownsanity.Pointer(300),
				MinLength: myownsanity.Pointer(1),
				Pattern:   `^\S+(?:\s+\S+)*$`,
			},
			"description": {
				ID:        "name",
				Comment:   "The name of the Spending object",
				Type:      "string",
				MaxLength: myownsanity.Pointer(300),
				MinLength: myownsanity.Pointer(0),
				Pattern:   `^\S+(?:\s+\S+)*$`,
			},
			"spendingType": {
				ID:   "spendingType",
				Type: "integer",
				Enum: []any{
					models.SpendingTypeGoal,
					models.SpendingTypeExpense,
				},
			},
			"targetAmount": {
				ID:      "targetAmount",
				Type:    "integer",
				Minimum: myownsanity.Pointer(0.0),
			},
			"currentAmount": {
				ID:      "currentAmount",
				Type:    "integer",
				Minimum: myownsanity.Pointer(0.0),
			},
			"nextRecurrence": {
				ID:      "nextRecurrence",
				Type:    "string",
				Format:  "date",
				Pattern: `^(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d\.\d+([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))$`,
			},
			"ruleset": RRule("ruleset"),
		},
		Required: []string{
			"fundingScheduleId",
			"name",
			"spendingType",
			"targetAmount",
			"nextRecurrence",
		},
		Extra: map[string]any{},
		AdditionalProperties: &jsonschema.Schema{
			Not: &jsonschema.Schema{},
		},
		DependentSchemas: map[string]*jsonschema.Schema{
			"ruleset": {
				ID: "ruleset",
				Properties: map[string]*jsonschema.Schema{
					"spendingType": {
						ID:    "spendingType",
						Const: myownsanity.Pointer[any](models.SpendingTypeExpense),
					},
				},
			},
		},
	}).Resolve, resolveOptions)
)
