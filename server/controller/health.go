package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/build"
)

func (c *Controller) handleHealth(ctx echo.Context) error {
	status := http.StatusOK
	err := c.DB.Ping(ctx.Request().Context())
	if err != nil {
		c.getLog(ctx).WithError(err).Warn("failed to ping database")
		status = http.StatusInternalServerError
	}

	result := map[string]any{
		"dbHealthy":  err == nil,
		"apiHealthy": true,
		"revision":   build.Revision,
		"buildTime":  build.BuildTime,
		"serverTime": c.Clock.Now().UTC(),
	}

	if build.Release != "" {
		result["release"] = build.Release
	} else {
		result["release"] = nil
	}

	return ctx.JSON(status, result)
}
