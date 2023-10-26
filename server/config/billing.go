package config

import (
	"github.com/monetr/monetr/server/feature"
)

type Plan struct {
	Visible       bool
	StripePriceId string
	Features      []feature.Feature
	Default       bool
}
