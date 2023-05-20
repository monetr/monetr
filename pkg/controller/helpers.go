package controller

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/util"
	"github.com/pkg/errors"
)

func (c *Controller) mustGetTimezone(ctx echo.Context) *time.Location {
	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		panic(err)
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		panic(err)
	}

	return timezone
}

func (c *Controller) midnightInLocal(ctx echo.Context, input time.Time) (time.Time, error) {
	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return input, errors.Wrap(err, "failed to retrieve account's timezone")
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		return input, errors.Wrap(err, "failed to parse account's timezone")
	}

	return util.MidnightInLocal(input, timezone), nil
}
