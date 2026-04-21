package schema

import (
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

func ID[T models.Identifiable]() *z.StringSchema[models.ID[T]] {
	var inst T
	prefix := inst.IdentityPrefix()
	msg := fmt.Sprintf("expected id with prefix %q", prefix)

	return z.StringLike[models.ID[T]]().
		// TODO Improve this even more
		TestFunc(
			func(val *models.ID[T], ctx z.Ctx) bool {
				return strings.HasPrefix(string(*val), prefix+"_")
			},
			z.IssueCode("invalid_id"),
			z.Message(msg),
		)
}
