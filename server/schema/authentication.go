package schema

import (
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/internals"
	locale "github.com/elliotcourant/go-lclocale"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/zoneinfo"
)

var (
	AuthenticationLogin = z.Struct(z.Shape{
		"email": z.String().
			Email().
			Required().
			Trim().
			Transform(func(valPtr *string, ctx internals.Ctx) error {
				*valPtr = strings.ToLower(*valPtr)
				return nil
			}),
		"password": z.String().
			Required().
			Trim().
			Min(8).
			Max(71),
	})

	AuthenticationRegister = z.Struct(z.Shape{
		"email": z.String().
			Email().
			Required().
			Trim().
			Transform(func(valPtr *string, ctx internals.Ctx) error {
				*valPtr = strings.ToLower(*valPtr)
				return nil
			}),
		"password": z.String().
			Required().
			Trim().
			Min(8).
			Max(71),
		"firstName": z.String().Required().Trim().Max(250),
		"lastName":  z.String().Optional().Trim().Max(250),
		"timezone": z.String().
			Default("UTC").
			Required().
			// TODO Make this its own schema type?
			TestFunc(func(val *string, ctx internals.Ctx) bool {
				_, err := zoneinfo.Timezone(*val)
				if err != nil {
					ctx.AddIssue(ctx.Issue().
						SetCode("timezone_invalid").
						SetPath([]string{"timezone"}).
						SetMessage("timezone not recognized by server").
						SetParams(map[string]any{
							"timezone": val,
						}),
					)
					return false
				}

				return true
			}),
		"locale": z.String().
			Default(consts.DefaultLocale).
			Required().
			Transform(func(valPtr *string, ctx internals.Ctx) error {
				if _, err := locale.GetLConv(*valPtr); err != nil {
					// If the provided locale is invalid, fallback to the default
					*valPtr = consts.DefaultLocale
				}

				return nil
			}),
		// TODO Make this a union
		"betaCode": z.String().Optional(),
	})
)
