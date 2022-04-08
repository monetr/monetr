package ui

import (
	"github.com/monetr/monetr/pkg/application"
	"github.com/monetr/monetr/pkg/config"
)

var (
	_ application.Controller = &UIController{}
)

type UIController struct {
	configuration config.Configuration
}

func NewUIController(configuration config.Configuration) *UIController {
	return &UIController{
		configuration: configuration,
	}
}
