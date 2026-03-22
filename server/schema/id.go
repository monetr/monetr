package schema

import (
	"fmt"
	"regexp"

	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

func ID[T models.Identifiable](options ...z.TestOption) *z.StringSchema[models.ID[T]] {
	inst := *new(T)
	return z.StringLike[models.ID[T]]().
		Match(
			regexp.MustCompile(fmt.Sprintf(`^%s_[0-7][0-9a-hjkmnp-tv-z]{25}$`, inst.IdentityPrefix())),
			// Merge the options that we default to with the options the caller
			// provides.
			append(
				[]z.TestOption{
					z.IssueCode("invalid_id"),
					z.Message("Must provide a valid ID"),
				},
				options...,
			)...,
		)
}
