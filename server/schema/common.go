package schema

import z "github.com/Oudwins/zog"

func Name(options ...z.TestOption) *z.StringSchema[string] {
	return z.String().
		Min(1,
			append(
				[]z.TestOption{
					z.IssueCode("name_too_short"),
					z.IssuePath([]string{"name"}),
					z.Message("Name must be at least one character"),
				},
				options...,
			)...,
		).
		Max(300,
			append(
				[]z.TestOption{
					z.IssueCode("name_too_long"),
					z.IssuePath([]string{"name"}),
					z.Message("Name cannot be longer than 300 characters"),
				},
				options...,
			)...,
		).
		Required(
			append(
				[]z.TestOption{
					z.IssueCode("name_required"),
					z.IssuePath([]string{"name"}),
					z.Message("Name is required"),
				},
				options...,
			)...,
		)
}
