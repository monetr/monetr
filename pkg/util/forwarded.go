package util

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// GetForwardedFor will return the IP address provided by the request header X-Forwarded-For or X-Real-Ip.
func GetForwardedFor(ctx echo.Context) string {
	values := []string{
		ctx.Request().Header.Get("Cf-Connecting-Ip"),
		ctx.Request().Header.Get("X-Original-Forwarded-For"),
		ctx.Request().Header.Get("X-Forwarded-For"),
		ctx.Request().Header.Get("X-Real-Ip"),
	}
	for _, value := range values {
		// The value of the forwared for header can be comma delimited coming from a cloud load balancer.
		items := strings.Split(value, ",")
		if len(items) > 0 && items[0] != "" {
			return items[0]
		}
	}

	return ""
}
