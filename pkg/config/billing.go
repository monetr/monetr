package config

import (
	"github.com/monetrapp/rest-api/pkg/feature"
)

type Plan struct {
	FreeTrialDays int32
	Visible       bool
	StripePriceId string
	Features      []feature.Feature
	Default       bool
}
