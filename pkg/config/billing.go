package config

import (
	"github.com/monetr/monetr/pkg/feature"
)

type Plan struct {
	Visible       bool
	StripePriceId string
	Features      []feature.Feature
	Default       bool
}
