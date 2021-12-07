package ui

import (
	"github.com/monetr/monetr/pkg/application"
)

var (
	_ application.Controller = &UIController{}
)

type UIController struct{}

func NewUIController() *UIController {
	return &UIController{}
}
