package schema

import (
	"github.com/Oudwins/zog"
	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
)

var (
	CreateLink = zog.Struct(zog.Shape{
		"institutionName": Name().Required(),
		"description":     Description().Optional(),
		"lunchFlowLinkId": z.Ptr(ID[models.LunchFlowLink]().Optional()),
	})

	PatchLink = zog.Struct(zog.Shape{
		"institutionName": Name().Optional(),
		"description":     Description().Optional(),
	})
)
